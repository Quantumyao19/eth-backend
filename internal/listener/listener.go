package listener

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/metrics"
	"eth-backend/internal/model"
	"eth-backend/internal/repository"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Listener struct {
	client       *ethclient.Client
	interval     time.Duration
	transferRepo *repository.TransferRepository
	stateRepo    *repository.ListenerStateRepository
	cleanupRepo  *repository.CleanupRepository
	redis        *redis.Client
	metrics      *metrics.Metrics

	lastProcessedBlock uint64
}

const (
	listenerInterval            = 1 * time.Minute
	cleanupInterval             = 10 * time.Minute
	transferListenerName        = "transfer_listener"
	maxBlockRange        uint64 = 10
	confirmations        uint64 = 12
	retainRecords        int    = 10000
)

func NewListener(client *ethclient.Client, transferRepo *repository.TransferRepository, stateRepo *repository.ListenerStateRepository, cleanupRepo *repository.CleanupRepository, redis *redis.Client, metrics *metrics.Metrics) *Listener {
	return &Listener{
		client:       client,
		interval:     listenerInterval,
		transferRepo: transferRepo,
		stateRepo:    stateRepo,
		cleanupRepo:  cleanupRepo,
		redis:        redis,
		metrics:      metrics,
	}
}

func (l *Listener) Start(ctx context.Context) {
	if err := l.initState(ctx); err != nil {
		logger.Log.Error("failed to init listener state", zap.Error(err))
		return
	}

	block, err := l.stateRepo.GetLastProcessedBlock(ctx, transferListenerName)
	if err != nil {
		logger.Log.Error("failed to get checkpoint", zap.Error(err))
		return
	}

	l.lastProcessedBlock = uint64(block)

	l.metrics.ListenerLastProcessedBlock.Set(float64(block))

	go l.loop(ctx)
}

func (l *Listener) loop(ctx context.Context) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	cleanupTicker := time.NewTicker(cleanupInterval)
	defer cleanupTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			if err := l.processCycle(ctx); err != nil {
				logger.Log.Error("listener cycle failed", zap.Error(err))
			}

		case <-cleanupTicker.C:
			err := l.cleanup(ctx)
			if err != nil {
				logger.Log.Error("cleaup failed")
			}
		}
	}
}

func (l *Listener) processCycle(ctx context.Context) error {
	start := time.Now()
	defer func() {
		l.metrics.ListenerProcessingDuration.Observe(time.Since(start).Seconds())
	}()

	ctxCycle, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	latestBlock, err := l.client.BlockNumber(ctxCycle)
	if err != nil {
		l.metrics.ListenerErrorsTotal.WithLabelValues("get_block").Inc()
		return err
	}

	if latestBlock <= confirmations {
		return nil
	}

	latestBlock -= confirmations

	fromBlock := l.lastProcessedBlock + 1

	if fromBlock > latestBlock {
		logger.Log.Info("no new blocks", zap.Uint64("latest", latestBlock))
		return nil
	}

	if latestBlock-fromBlock > maxBlockRange {
		latestBlock = fromBlock + maxBlockRange
	}

	logs, err := l.fetchLogs(ctxCycle, fromBlock, latestBlock)
	if err != nil {
		l.metrics.ListenerErrorsTotal.WithLabelValues("fetch_logs").Inc()
		return err
	}

	var transfers []*model.Transfer
	addressSet := make(map[string]struct{})

	for _, vLog := range logs {
		if t := l.parseTransfer(vLog); t != nil {
			transfers = append(transfers, t)
			addressSet[t.From] = struct{}{}
			addressSet[t.To] = struct{}{}
		}
	}

	inserted, err := l.persistTransfers(ctxCycle, transfers, latestBlock)
	if err != nil {
		return err
	}

	l.lastProcessedBlock = latestBlock

	if len(transfers) > 0 {
		l.invalidateTransaferCache(ctxCycle, addressSet)
		l.metrics.ListenerEventsProcessedTotal.Add(float64(len(transfers)))
	}

	l.metrics.ListenerLastProcessedBlock.Set(float64(latestBlock))
	l.metrics.ListenerBlocksProcessedTotal.Inc()

	logger.Log.Info("listener cycle completed", zap.Uint64("from_block", fromBlock), zap.Uint64("to_block", latestBlock), zap.Int("transfers", len(transfers)), zap.Int64("inserted", inserted))

	return nil
}

func (l *Listener) persistTransfers(ctx context.Context, transfers []*model.Transfer, latestBlock uint64) (int64, error) {
	tx, err := l.transferRepo.Begin()
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()

	var inserted int64

	if len(transfers) > 0 {
		inserted, err = l.transferRepo.InsertManyTx(ctx, tx, transfers)
		if err != nil {
			return 0, err
		}
	}

	if err := l.stateRepo.UpdateLastProcessedBlockTx(ctx, tx, transferListenerName, int64(latestBlock)); err != nil {
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return inserted, nil
}

func (l *Listener) fetchLogs(ctx context.Context, from, to uint64) ([]types.Log, error) {
	transferSigHash := crypto.Keccak256Hash([]byte("Transfer(address,address,uint256)"))

	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(from)),
		ToBlock:   big.NewInt(int64(to)),
		Topics: [][]common.Hash{
			{
				transferSigHash,
			},
		},
	}

	return l.client.FilterLogs(ctx, query)
}

func (l *Listener) parseTransfer(vLog types.Log) *model.Transfer {
	if len(vLog.Topics) < 3 {
		return nil
	}

	from := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	value := new(big.Int).SetBytes(vLog.Data)
	bigInt := model.BigInt{Int: value}

	transfer := &model.Transfer{
		TxHash:       vLog.TxHash.Hex(),
		LogIndex:     uint(vLog.Index),
		BlockNumber:  vLog.BlockNumber,
		TokenAddress: vLog.Address.Hex(),
		From:         from.Hex(),
		To:           to.Hex(),
		Value:        bigInt,
	}

	return transfer
}

func (l *Listener) invalidateTransaferCache(ctx context.Context, addressSet map[string]struct{}) {
	for address := range addressSet {
		indexKey := fmt.Sprintf("transfer:index:%s", address)

		keys, err := l.redis.SMembers(ctx, indexKey).Result()
		if err != nil {
			logger.Log.Warn("get cache index failed", zap.Error(err), zap.String("address", address))
			return
		}

		if len(keys) > 0 {
			if err := l.redis.Del(ctx, indexKey).Err(); err != nil {
				logger.Log.Warn("delete index key failed", zap.Error(err), zap.String("indexKey", indexKey))
			}

		}
	}
}

func (l *Listener) initState(ctx context.Context) error {
	if err := l.stateRepo.EnsureState(ctx, transferListenerName); err != nil {
		return err
	}

	latest, err := l.client.BlockNumber(ctx)
	if err != nil {
		return err
	}

	if latest > confirmations {
		latest -= confirmations
	}

	return l.stateRepo.UpdateLastProcessedBlock(ctx, transferListenerName, int64(latest))
}

func (l *Listener) cleanup(ctx context.Context) error {
	return l.cleanupRepo.DeleteOldTransfers(ctx, retainRecords)

}

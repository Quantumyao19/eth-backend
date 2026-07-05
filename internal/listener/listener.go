package listener

import (
	"context"
	"eth-backend/internal/logger"
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
	client   *ethclient.Client
	interval time.Duration
	repo     *repository.TransferRepository
	redis    *redis.Client
}

const (
	defaultInterval = 5 * time.Minute
)

func NewListener(client *ethclient.Client, repo *repository.TransferRepository, redis *redis.Client) *Listener {
	return &Listener{
		client:   client,
		interval: defaultInterval,
		repo:     repo,
		redis:    redis,
	}
}

func (l *Listener) Start(ctx context.Context) {
	go l.loop(ctx)
}

func (l *Listener) loop(ctx context.Context) {
	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	var lastBlock uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			func() {
				ctxCycle, cancel := context.WithTimeout(ctx, 10*time.Second)
				defer cancel()

				latestBlock, err := l.client.BlockNumber(ctxCycle)

				if err != nil {
					logger.Log.Error("get block error", zap.Error(err))
					return
				}

				if lastBlock == 0 {
					lastBlock = latestBlock - 5
					return
				}

				logs, err := l.fetchLogs(ctxCycle, lastBlock, latestBlock)
				if err != nil {
					logger.Log.Error("fetch logs error", zap.Error(err))
					return
				}

				var transfers []*model.Transfer
				addressSet := make(map[string]struct{})

				for _, vLog := range logs {
					t := l.parseTransfer(vLog)
					if t != nil {
						transfers = append(transfers, t)

						addressSet[t.From] = struct{}{}
						addressSet[t.To] = struct{}{}
					}
				}

				if len(transfers) > 0 {
					inserted, err := l.repo.InsertMany(ctxCycle, transfers)
					if err != nil {
						logger.Log.Error("batch insert error", zap.Error(err))
						return
					} else {
						logger.Log.Info("batch insert completed", zap.Int("requested", len(transfers)), zap.Int64("inserted", inserted))
						l.invalidateTransaferCache(ctxCycle, addressSet)
					}
				}

				lastBlock = latestBlock

			}()

		}
	}
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
		LogIndex:     uint(vLog.TxIndex),
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

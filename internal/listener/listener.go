package listener

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/model"
	"eth-backend/internal/repository"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"go.uber.org/zap"
)

type Listener struct {
	client   *ethclient.Client
	interval time.Duration
	repo     *repository.TransferRepository
}

const (
	defaultInterval = 5 * time.Second
)

func NewListener(client *ethclient.Client, repo *repository.TransferRepository) *Listener {
	return &Listener{
		client:   client,
		interval: defaultInterval,
		repo:     repo,
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
			latestBlock, err := l.client.BlockNumber(ctx)
			if err != nil {
				logger.Log.Error("get block error", zap.Error(err))
				continue
			}

			if lastBlock == 0 {
				lastBlock = latestBlock - 5
				continue
			}

			logs, err := l.fetchLogs(ctx, lastBlock, latestBlock)
			if err != nil {
				logger.Log.Error("fetch logs error", zap.Error(err))
				continue
			}

			var transfers []*model.Transfer
			for _, vLog := range logs {
				t := l.parseTransfer(vLog)
				if t != nil {
					transfers = append(transfers, t)
				}
			}

			if len(transfers) > 0 {
				inserted, err := l.repo.InsertMany(ctx, transfers)
				if err != nil {
					logger.Log.Error("batch insert error", zap.Error(err))
				} else {
					logger.Log.Info("batch insert completed", zap.Int("requested", len(transfers)), zap.Int64("inserted", inserted))
				}
			}

			lastBlock = latestBlock
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

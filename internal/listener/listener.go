package listener

import (
	"context"
	"eth-backend/internal/logger"
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
}

func NewListener(client *ethclient.Client) *Listener {
	return &Listener{
		client:   client,
		interval: 5 * time.Second,
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

			for _, vLog := range logs {
				l.handleLog(vLog)
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

func (l *Listener) handleLog(vLog types.Log) {
	if len(vLog.Topics) < 3 {
		return
	}

	from := common.HexToAddress(vLog.Topics[1].Hex())
	to := common.HexToAddress(vLog.Topics[2].Hex())

	value := new(big.Int).SetBytes(vLog.Data)

	logger.Log.Info("Token", zap.String("token", vLog.Address.Hex()))
	logger.Log.Info("From", zap.String("from", from.String()))
	logger.Log.Info("To", zap.String("to", to.String()))
	logger.Log.Info("Value", zap.String("value", value.String()))

}

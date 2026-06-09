package eth

import (
	"context"
	"eth-backend/utils"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Service struct {
	client *Client
}

func NewService(client *Client) *Service {
	return &Service{client: client}
}

func (s *Service) GetBlockNumber(ctx context.Context) (uint64, error) {
	return s.client.rpc.BlockNumber(ctx)
}

func (s *Service) GetBalance(ctx context.Context, addr string) (string, string, error) {
	balance, err := s.client.rpc.BalanceAt(ctx, common.HexToAddress(addr), nil)
	if err != nil {
		return "", "", err
	}

	wei := balance.String()
	eth := utils.WeiToETH(balance)
	return wei, eth, nil
}

func (s *Service) GetTransaction(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := s.client.TransactionByHash(ctx, hash)
	if err != nil {
		log.Fatal(err)
	}
	return tx, isPending, nil
}

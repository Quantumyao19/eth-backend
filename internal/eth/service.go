package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
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

func (s *Service) GetBalance(ctx context.Context, addr string) (string, error) {
	balance, err := s.client.rpc.BalanceAt(ctx, common.HexToAddress(addr), nil)
	if err != nil {
		return "", err
	}
	return balance.String(), nil
}

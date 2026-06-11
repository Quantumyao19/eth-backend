package eth

import (
	"context"
	"eth-backend/config"
	"eth-backend/utils"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Service struct {
	client  *Client
	chainID *big.Int
}

func NewService(client *Client, cfg config.EthConfig) (*Service, error) {
	ctx := context.Background()
	chainID, err := client.rpc.NetworkID(ctx)
	if err != nil {
		return nil, err
	}

	//parse chainID from EthConfig
	configChainID := new(big.Int)
	_, ok := configChainID.SetString(cfg.ChainID, 10)
	if !ok {
		return nil, fmt.Errorf("invalid CHAIN_ID: %s", cfg.ChainID)
	}

	//compare chainID
	if configChainID != nil && configChainID.Cmp(chainID) != 0 {
		log.Fatalf("chainID mismatch: config=%s rpc=%s",
			configChainID.String(),
			chainID.String(),
		)
	}

	return &Service{
		client:  client,
		chainID: chainID,
	}, nil
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

func (s *Service) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	return s.client.rpc.TransactionReceipt(ctx, hash)
}

func (s *Service) GetTransactionSender(ctx context.Context, tx *types.Transaction) (common.Address, error) {
	signer := types.LatestSignerForChainID(s.chainID)
	return types.Sender(signer, tx)
}

package eth

import (
	"context"
	"eth-backend/config"
	"eth-backend/utils"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const erc20MetaABI = `[
  {
    "constant": true,
    "inputs": [],
    "name": "decimals",
    "outputs": [{"name": "", "type": "uint8"}],
    "type": "function"
  },
  {
    "constant": true,
    "inputs": [],
    "name": "symbol",
    "outputs": [{"name": "", "type": "string"}],
    "type": "function"
  }
]`

type Service struct {
	client  *Client
	chainID *big.Int
	metaABI abi.ABI
}

func NewService(client *Client, cfg config.EthConfig) (*Service, error) {
	ctx := context.Background()
	chainID, err := client.rpc.ChainID(ctx)
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
	if configChainID.Cmp(chainID) != 0 {
		return nil, fmt.Errorf("chainID mismatch: config=%s rpc=%s",
			configChainID.String(),
			chainID.String(),
		)
	}

	parsedABI, err := abi.JSON(strings.NewReader(erc20MetaABI))
	if err != nil {
		return nil, err
	}

	return &Service{
		client:  client,
		chainID: chainID,
		metaABI: parsedABI,
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
		return nil, false, err
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

func (s *Service) GetTokenMeta(ctx context.Context, token common.Address) (string, uint8, error) {
	data, _ := s.metaABI.Pack("decimals")
	res, err := s.client.rpc.CallContract(ctx, ethereum.CallMsg{
		To:   &token,
		Data: data,
	}, nil)
	if err != nil {
		return "", 0, err
	}

	var decimals uint8
	err = s.metaABI.UnpackIntoInterface(&decimals, "decimals", res)
	if err != nil {
		return "", 0, err
	}

	data, _ = s.metaABI.Pack("symbol")
	res, err = s.client.rpc.CallContract(ctx, ethereum.CallMsg{
		To:   &token,
		Data: data,
	}, nil)
	if err != nil {
		return "", 0, err
	}

	var symbol string
	err = s.metaABI.UnpackIntoInterface(&symbol, "symbol", res)
	if err != nil {
		return "", 0, err
	}

	return symbol, decimals, nil
}

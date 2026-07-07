package eth

import (
	"context"
	"eth-backend/internal/metrics"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

const (
	rpcMethodBlockNumber           = "eth_blockNumber"
	rpcMethodGetBalance            = "eth_getBalance"
	rpcMethodTransactionByHash     = "eth_getTransactionByHash"
	rpcMethodGetTransactionReceipt = "eth_getTransactionReceipt"
	rpcMethodCallContract          = "eth_call"
)

type Client struct {
	rpc     *ethclient.Client
	metrics *metrics.Metrics
}

func NewClient(rpcURL string, metrics *metrics.Metrics) (*Client, error) {
	c, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{rpc: c, metrics: metrics}, nil
}

func (c *Client) Raw() *ethclient.Client {
	return c.rpc
}

func (c *Client) Close() {
	c.rpc.Close()
}

func (c *Client) ChainID(ctx context.Context) (*big.Int, error) {
	return c.rpc.ChainID(ctx)
}

func (c *Client) GetBlockNumber(ctx context.Context) (block uint64, err error) {
	defer c.observeRPC(rpcMethodBlockNumber, time.Now(), &err)
	block, err = c.rpc.BlockNumber(ctx)
	return
}

func (c *Client) GetBalance(ctx context.Context, addr string) (balance *big.Int, err error) {
	defer c.observeRPC(rpcMethodGetBalance, time.Now(), &err)
	balance, err = c.rpc.BalanceAt(ctx, common.HexToAddress(addr), nil)
	return
}

func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (tx *types.Transaction, isPending bool, err error) {
	defer c.observeRPC(rpcMethodTransactionByHash, time.Now(), &err)
	tx, isPending, err = c.rpc.TransactionByHash(ctx, hash)
	return
}

func (c *Client) GetTransactionReceipt(ctx context.Context, hash common.Hash) (receipt *types.Receipt, err error) {
	defer c.observeRPC(rpcMethodGetTransactionReceipt, time.Now(), &err)
	receipt, err = c.rpc.TransactionReceipt(ctx, hash)
	return
}

func (c *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) (result []byte, err error) {
	defer c.observeRPC(rpcMethodCallContract, time.Now(), &err)
	result, err = c.rpc.CallContract(ctx, msg, blockNumber)
	return
}

func (c *Client) observeRPC(method string, start time.Time, err *error) {
	if c.metrics == nil {
		return
	}

	duration := time.Since(start).Seconds()

	c.metrics.RPCRequestsTotal.WithLabelValues(method).Inc()
	c.metrics.RPCRequestDuration.WithLabelValues(method).Observe(duration)

	if err != nil && *err != nil {
		c.metrics.RPCRequestsErrors.WithLabelValues(method).Inc()
	}
}

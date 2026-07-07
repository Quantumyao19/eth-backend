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
	rpcMethodTransactionByHash     = "eth_transactionByHash"
	rpcMethodGetTransactionReceipt = "eth_getTransactionReceipt"
	rpcMethodCallContract          = "eth_callContract"
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

func (c *Client) GetBlockNumber(ctx context.Context) (uint64, error) {
	defer c.observeRPC(rpcMethodBlockNumber, time.Now())
	return c.rpc.BlockNumber(ctx)
}

func (c *Client) GetBalance(ctx context.Context, addr string) (*big.Int, error) {
	defer c.observeRPC(rpcMethodGetBalance, time.Now())
	return c.rpc.BalanceAt(ctx, common.HexToAddress(addr), nil)
}

func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	defer c.observeRPC(rpcMethodTransactionByHash, time.Now())
	return c.rpc.TransactionByHash(ctx, hash)
}

func (c *Client) GetTransactionReceipt(ctx context.Context, hash common.Hash) (*types.Receipt, error) {
	defer c.observeRPC(rpcMethodGetTransactionReceipt, time.Now())
	return c.rpc.TransactionReceipt(ctx, hash)
}

func (c *Client) CallContract(ctx context.Context, msg ethereum.CallMsg, blockNumber *big.Int) ([]byte, error) {
	defer c.observeRPC(rpcMethodCallContract, time.Now())
	return c.rpc.CallContract(ctx, msg, blockNumber)
}

func (c *Client) observeRPC(method string, start time.Time) {
	if c.metrics == nil {
		return
	}

	duration := time.Since(start).Seconds()

	c.metrics.RPCRequestsTotal.WithLabelValues(method).Inc()
	c.metrics.RPCRequestDuration.WithLabelValues(method).Observe(duration)
}

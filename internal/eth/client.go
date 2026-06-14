package eth

import (
	"context"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Client struct {
	rpc *ethclient.Client
}

func NewClient(rpcURL string) (*Client, error) {
	c, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}
	return &Client{rpc: c}, nil
}

func (c *Client) Raw() *ethclient.Client {
	return c.rpc
}

func (c *Client) Close() {
	c.rpc.Close()
}

func (c *Client) TransactionByHash(ctx context.Context, hash common.Hash) (*types.Transaction, bool, error) {
	tx, isPending, err := c.rpc.TransactionByHash(ctx, hash)
	if err != nil {
		return nil, false, err
	}
	return tx, isPending, nil
}

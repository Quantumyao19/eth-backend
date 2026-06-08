package eth

import "github.com/ethereum/go-ethereum/ethclient"

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

func (c *Client) Close() {
	c.rpc.Close()
}

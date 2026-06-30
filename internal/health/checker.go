package health

import (
	"context"
	"eth-backend/internal/eth"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/redis/go-redis/v9"
)

const (
	timeoutForHealthCheck = 300 * time.Millisecond
)

type Checker struct {
	DB    *goqu.Database
	Redis *redis.Client
	RPC   *eth.Service
}

func NewChecker(db *goqu.Database, redisClient *redis.Client, rpc *eth.Service) *Checker {
	return &Checker{
		DB:    db,
		Redis: redisClient,
		RPC:   rpc,
	}
}

func (c *Checker) checkDB(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutForHealthCheck)
	defer cancel()

	var one int
	_, err := c.DB.Select(goqu.L("1")).ScanValContext(ctx, &one)
	return err

}

func (c *Checker) checkRedis(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutForHealthCheck)
	defer cancel()

	return c.Redis.Ping(ctx).Err()
}

func (c *Checker) checkRPC(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutForHealthCheck)
	defer cancel()

	_, err := c.RPC.GetBlockNumber(ctx)
	return err
}

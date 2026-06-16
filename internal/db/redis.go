package db

import (
	"context"
	"time"

	"eth-backend/internal/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedisClient(addr string, password string, db int) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,

		PoolSize:     20,
		MinIdleConns: 5,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Log.Fatal("Failed to connect to Redis",
			zap.String("addr", addr),
			zap.Error(err))
	}

	return rdb
}

package deps

import (
	"context"
	"eth-backend/internal/health"

	"github.com/redis/go-redis/v9"
)

type RedisDependency struct {
	redis *redis.Client
}

func NewRedisDependency(redis *redis.Client) health.Dependency {
	return &RedisDependency{redis: redis}
}

func (r *RedisDependency) Name() string {
	return "Redis"
}

func (r *RedisDependency) Check(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, timeoutForHealthCheck)
	defer cancel()

	return r.redis.Ping(ctx).Err()
}

func (r *RedisDependency) Weight() int {
	return 20
}

func (r *RedisDependency) Critical() bool {
	return false
}

package handler

import (
	"context"
	"encoding/json"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"eth-backend/internal/repository"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type TransferHandler struct {
	repo  *repository.TransferRepository
	redis *redis.Client
}

const (
	defaultPageSize      = 20
	transferListCacheTTL = 1 * time.Minute
	emptyResultCacheTTL  = 1 * time.Minute
	transferLockTTL      = 10 * time.Second
	cacheWaitRetries     = 6
	cacheMinInterval     = 30 * time.Millisecond
	cacheMaxInterval     = 300 * time.Millisecond
)

func NewTransferHandler(repo *repository.TransferRepository, redisClient *redis.Client) *TransferHandler {
	return &TransferHandler{
		repo:  repo,
		redis: redisClient,
	}
}

func (handler *TransferHandler) ListTransfer(c *gin.Context) {
	requestID, _ := c.Request.Context().Value(middleware.RequestIDKey).(string)
	ctx := c.Request.Context()
	address := c.Query("address")
	if address == "" {
		writeError(c, "address required", http.StatusBadRequest)
		return
	}

	if !isValidEthereumAddress(address) {
		writeError(c, "invalid ethereum address format", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	cacheKey := fmt.Sprintf("transfer:list:%s:%d:%d", address, page, pageSize)
	lockKey := fmt.Sprintf("transfer:lock:%s:%d:%d", address, page, pageSize)

	cached, err := handler.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		logger.Log.Debug("transfer cache hit", zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
		writeCachedJSON(c, cached)
		return
	} else if err != redis.Nil {
		logger.Log.Warn("redis cache read failed", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
	}

	// tried to obtain Redis distributed lock
	lockValue := uuid.NewString()
	locked, err := handler.redis.SetNX(ctx, lockKey, lockValue, transferLockTTL).Result()
	if err != nil {
		logger.Log.Warn("failed to acquire transfer cache lock", zap.Error(err), zap.String("lock_key", lockKey), zap.String("request_id", requestID))
	}

	if locked {
		defer handler.releaseTransferLock(ctx, lockKey, lockValue, requestID)
		resp, err := handler.queryAndCacheTransfers(ctx, cacheKey, address, page, pageSize, requestID)
		if err != nil {
			handleError(c, err)
			return
		}
		writeJSON(c, resp)
		return
	}

	if err == nil {
		cacheWaitInterval := cacheMinInterval
		for i := 0; i < cacheWaitRetries; i++ {
			timer := time.NewTimer(jitterCacheWaitInterval(cacheWaitInterval))
			select {
			case <-ctx.Done():
				timer.Stop()
				handleError(c, ctx.Err())
				return
			case <-timer.C:
			}

			cached, err := handler.redis.Get(ctx, cacheKey).Result()
			if err == nil {
				logger.Log.Debug("transfer cache hit after wait", zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
				writeCachedJSON(c, cached)
				return
			}
			if err != redis.Nil {
				logger.Log.Warn("redis cache retry failed", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
				break
			}

			cacheWaitInterval *= 2
			if cacheWaitInterval > cacheMaxInterval {
				cacheWaitInterval = cacheMaxInterval
			}
		}
	}

	resp, err := handler.queryAndCacheTransfers(ctx, cacheKey, address, page, pageSize, requestID)
	if err != nil {
		handleError(c, err)
		return
	}
	writeJSON(c, resp)
}

func jitterCacheWaitInterval(interval time.Duration) time.Duration {
	jitterRange := interval * 4 / 10
	if jitterRange <= 0 {
		return interval
	}

	waitInterval := interval*8/10 + time.Duration(rand.Int63n(int64(jitterRange)))
	if waitInterval > cacheMaxInterval {
		return cacheMaxInterval
	}

	return waitInterval
}

func writeCachedJSON(c *gin.Context, cached string) {
	c.Data(http.StatusOK, "application/json", []byte(cached))
}

func (handler *TransferHandler) queryAndCacheTransfers(ctx context.Context, cacheKey string, address string, page int, pageSize int, requestID string) (map[string]interface{}, error) {
	// retry to obtain data from redis cache
	cached, err := handler.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		var resp map[string]interface{}

		if err := json.Unmarshal([]byte(cached), &resp); err == nil {
			return resp, nil
		}

		logger.Log.Warn("failed to unmarshal cached transfer response", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))

	} else if err != redis.Nil {
		logger.Log.Warn("redis cache read failed", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))

	}

	data, total, err := handler.repo.ListByAddress(ctx, address, page, pageSize)
	if err != nil {
		logger.Log.Error("transfer repo list failed", zap.Error(err), zap.String("address", address), zap.String("request_id", requestID))
		return nil, err
	}

	resp := map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"data":      data,
	}

	jsonBytes, err := json.Marshal(resp)
	if err != nil {
		logger.Log.Error("failed to marshal transfer response", zap.Error(err), zap.String("request_id", requestID))
		return nil, err
	}

	cacheTTL := transferListCacheTTL
	if total == 0 {
		cacheTTL = emptyResultCacheTTL
	}

	if err := handler.redis.Set(ctx, cacheKey, jsonBytes, cacheTTL).Err(); err != nil {
		logger.Log.Warn("failed to set transfer cache", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
	}

	indexKey := fmt.Sprintf("transfer:index:%s", address)
	if err := handler.redis.SAdd(ctx, indexKey, cacheKey).Err(); err != nil {
		logger.Log.Warn("failed to add cache index", zap.Error(err))
	}

	handler.redis.Expire(ctx, indexKey, transferListCacheTTL)

	return resp, nil
}

func (handler *TransferHandler) releaseTransferLock(ctx context.Context, lockKey string, lockValue string, requestID string) {
	// use LUA script to guarantee atomic operations
	const unlockScript = `
	if redis.call("GET", KEYS[1]) == ARGV[1] then
    	return redis.call("DEL", KEYS[1])
	end
	return 0
	`

	if err := handler.redis.Eval(ctx, unlockScript, []string{lockKey}, lockValue).Err(); err != nil {
		logger.Log.Warn("failed to release transfer cache lock", zap.Error(err), zap.String("lock_key", lockKey), zap.String("request_id", requestID))
	}
}

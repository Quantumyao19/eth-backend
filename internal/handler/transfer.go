package handler

import (
	"context"
	"encoding/json"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"eth-backend/internal/repository"
	"fmt"
	"net/http"
	"strconv"
	"time"

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
	transferLockTTL      = 3 * time.Second
	cacheWaitRetries     = 6
	cacheWaitInterval    = 50 * time.Millisecond
)

func NewTransferHandler(repo *repository.TransferRepository, redisClient *redis.Client) *TransferHandler {
	return &TransferHandler{
		repo:  repo,
		redis: redisClient,
	}
}

func (handler *TransferHandler) ListTransfer(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)
	ctx := r.Context()
	address := r.URL.Query().Get("address")
	if address == "" {
		writeError(w, "address required", http.StatusBadRequest)
		return
	}

	if !isValidEthereumAddress(address) {
		writeError(w, "invalid ethereum address format", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	cacheKey := fmt.Sprintf("transfer:list:%s:%d:%d", address, page, pageSize)
	lockKey := fmt.Sprintf("transfer:lock:%s:%d:%d", address, page, pageSize)

	cached, err := handler.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		logger.Log.Debug("transfer cache hit", zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
		writeCachedJSON(w, cached)
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
			handleError(w, err)
			return
		}
		writeJSON(w, resp)
		return
	}

	if err == nil {
		for i := 0; i < cacheWaitRetries; i++ {
			timer := time.NewTimer(cacheWaitInterval)
			select {
			case <-ctx.Done():
				timer.Stop()
				handleError(w, ctx.Err())
				return
			case <-timer.C:
			}

			cached, err := handler.redis.Get(ctx, cacheKey).Result()
			if err == nil {
				logger.Log.Debug("transfer cache hit after wait", zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
				writeCachedJSON(w, cached)
				return
			}
			if err != redis.Nil {
				logger.Log.Warn("redis cache retry failed", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
				break
			}
		}
	}

	resp, err := handler.queryAndCacheTransfers(ctx, cacheKey, address, page, pageSize, requestID)
	if err != nil {
		handleError(w, err)
		return
	}
	writeJSON(w, resp)
}

func writeCachedJSON(w http.ResponseWriter, cached string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(cached))
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

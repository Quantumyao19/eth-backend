package handler

import (
	"encoding/json"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"eth-backend/internal/repository"
	"fmt"
	"net/http"
	"strconv"
	"time"

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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	cacheKey := fmt.Sprintf("transfer:list:%s:%d:%d", address, page, pageSize)

	cached, err := handler.redis.Get(ctx, cacheKey).Result()
	if err == nil {
		logger.Log.Debug("transfer cache hit", zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(cached))
		return
	}

	if err != nil && err != redis.Nil {
		logger.Log.Warn("redis cache read failed", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
	}

	data, total, err := handler.repo.ListByAddress(ctx, address, page, pageSize)
	if err != nil {
		logger.Log.Error("transfer repo list failed", zap.Error(err), zap.String("address", address), zap.String("request_id", requestID))
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
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
		handleError(w, err)
		return
	}

	if err := handler.redis.Set(ctx, cacheKey, jsonBytes, transferListCacheTTL).Err(); err != nil {
		logger.Log.Warn("failed to set transfer cache", zap.Error(err), zap.String("cache_key", cacheKey), zap.String("request_id", requestID))
	}

	writeJSON(w, resp)
}

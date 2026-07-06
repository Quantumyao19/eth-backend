package handler

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const defaultTimeout = 5 * time.Second

func (h *Handler) BlockNumber(c *gin.Context) {
	requestID, _ := c.Request.Context().Value(middleware.RequestIDKey).(string)

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultTimeout)
	defer cancel()

	block, err := h.service.GetBlockNumber(ctx)
	if err != nil {
		logger.Log.Error("get blocknumber error", zap.Error(err), zap.String("request_id", requestID))
		handleError(c, err)
		return
	}

	writeJSON(c, map[string]uint64{
		"block": block,
	})
}

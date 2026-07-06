package handler

import (
	"context"
	"eth-backend/internal/eth"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type Handler struct {
	service *eth.Service
}

func NewHandler(s *eth.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Balance(c *gin.Context) {
	requestID, _ := c.Request.Context().Value(middleware.RequestIDKey).(string)

	addr := c.Query("address")
	if addr == "" {
		writeError(c, "missing address", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultTimeout)
	defer cancel()

	wei, eth, err := h.service.GetBalance(ctx, addr)
	if err != nil {
		logger.Log.Error("GetBalance error", zap.Error(err), zap.String("request_id", requestID))
		handleError(c, err)
		return
	}

	writeJSON(c, map[string]string{
		"address":     addr,
		"balance_wei": wei,
		"balance_eth": eth,
	})
}

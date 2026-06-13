package handler

import (
	"context"
	"eth-backend/internal/eth"
	"eth-backend/internal/logger"
	"net/http"

	"go.uber.org/zap"
)

type Handler struct {
	service *eth.Service
}

func NewHandler(s *eth.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("address")
	if addr == "" {
		writeError(w, "missing address", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
	defer cancel()

	wei, eth, err := h.service.GetBalance(ctx, addr)
	if err != nil {
		logger.Log.Fatal("GetBalance error", zap.Error(err))
		handleError(w, err)
		return
	}

	writeJSON(w, map[string]string{
		"address":     addr,
		"balance_wei": wei,
		"balance_eth": eth,
	})
}

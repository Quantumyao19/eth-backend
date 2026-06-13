package handler

import (
	"context"
	"encoding/json"
	"eth-backend/internal/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

const defaultTimeout = 5 * time.Second

func (h *Handler) BlockNumber(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
	defer cancel()

	block, err := h.service.GetBlockNumber(ctx)
	if err != nil {
		logger.Log.Fatal("get blocknumber error", zap.Error(err))
		handleError(w, err)
		return
	}

	writeJSON(w, map[string]uint64{
		"block": block,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

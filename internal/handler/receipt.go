package handler

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"go.uber.org/zap"
)

type ReceiptResponse struct {
	TxHash  string `json:"tx_hash"`
	Status  uint64 `json:"status"`
	GasUsed uint64 `json:"gas_used"`
}

func (h *Handler) Receipt(w http.ResponseWriter, r *http.Request) {
	requestID, _ := r.Context().Value(middleware.RequestIDKey).(string)

	hash := r.URL.Query().Get("hash")
	if hash == "" {
		writeError(w, "missing hash", http.StatusBadRequest)
		return
	}

	if !common.IsHexHash(hash) {
		writeError(w, "invalid hash format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
	defer cancel()

	receipt, err := h.service.GetTransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		logger.Log.Error("get receipt error", zap.Error(err), zap.String("request_id", requestID))
		handleError(w, err)
		return
	}

	if receipt == nil {
		writeError(w, "receipt not found", http.StatusNotFound)
		return
	}

	resp := ReceiptResponse{
		TxHash:  receipt.TxHash.Hex(),
		Status:  receipt.Status,
		GasUsed: receipt.GasUsed,
	}

	writeJSON(w, resp)
}

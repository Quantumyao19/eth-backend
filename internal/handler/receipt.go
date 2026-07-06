package handler

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ReceiptResponse struct {
	TxHash  string `json:"tx_hash"`
	Status  uint64 `json:"status"`
	GasUsed uint64 `json:"gas_used"`
}

func (h *Handler) Receipt(c *gin.Context) {
	requestID, _ := c.Request.Context().Value(middleware.RequestIDKey).(string)

	hash := c.Query("hash")
	if hash == "" {
		writeError(c, "missing hash", http.StatusBadRequest)
		return
	}

	if !common.IsHexHash(hash) {
		writeError(c, "invalid hash format", http.StatusBadRequest)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), defaultTimeout)
	defer cancel()

	receipt, err := h.service.GetTransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		logger.Log.Error("get receipt error", zap.Error(err), zap.String("request_id", requestID))
		handleError(c, err)
		return
	}

	if receipt == nil {
		writeError(c, "receipt not found", http.StatusNotFound)
		return
	}

	resp := ReceiptResponse{
		TxHash:  receipt.TxHash.Hex(),
		Status:  receipt.Status,
		GasUsed: receipt.GasUsed,
	}

	writeJSON(c, resp)
}

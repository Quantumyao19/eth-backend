package handler

import (
	"context"
	"encoding/hex"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TransactionResponse struct {
	Pending  bool   `json:"pending"`
	Hash     string `json:"hash"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasLimit uint64 `json:"gas_limit"`
	Nonce    uint64 `json:"nonce"`
	Input    string `json:"input"`
	GasPrice string `json:"gas_price"`
}

func (h *Handler) Transaction(c *gin.Context) {
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

	tx, isPending, err := h.service.GetTransaction(ctx, common.HexToHash(hash))
	if err != nil {
		logger.Log.Error("gettransaction error", zap.Error(err), zap.String("request_id", requestID))
		handleError(c, err)
		return
	}

	if tx == nil {
		writeError(c, "transaction not found", http.StatusNotFound)
		return
	}

	resp := buildTransactionResponse(tx, isPending)
	writeJSON(c, resp)
}

func buildTransactionResponse(tx *types.Transaction, isPending bool) TransactionResponse {
	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	return TransactionResponse{
		Pending:  isPending,
		Hash:     tx.Hash().Hex(),
		To:       to,
		Value:    tx.Value().String(),
		GasLimit: tx.Gas(),
		Nonce:    tx.Nonce(),
		Input:    hex.EncodeToString(tx.Data()),
		GasPrice: tx.GasPrice().String(),
	}
}

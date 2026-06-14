package handler

import (
	"context"
	"encoding/hex"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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

func (h *Handler) Transaction(w http.ResponseWriter, r *http.Request) {
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

	tx, isPending, err := h.service.GetTransaction(ctx, common.HexToHash(hash))
	if err != nil {
		logger.Log.Error("gettransaction error", zap.Error(err), zap.String("request_id", requestID))
		handleError(w, err)
		return
	}

	if tx == nil {
		writeError(w, "transaction not found", http.StatusNotFound)
		return
	}

	resp := buildTransactionResponse(tx, isPending)
	writeJSON(w, resp)
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

package handler

import (
	"encoding/hex"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
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
	hash := r.URL.Query().Get("hash")
	if hash == "" {
		http.Error(w, "missing hash", http.StatusBadRequest)
		return
	}

	if !common.IsHexHash(hash) {
		http.Error(w, "invalid hash format", http.StatusBadRequest)
		return
	}

	tx, isPending, err := h.service.GetTransaction(r.Context(), common.HexToHash(hash))
	if err != nil {
		http.Error(w, "failed to get transaction", http.StatusInternalServerError)
		return
	}

	if tx == nil {
		http.Error(w, "transaction not found", http.StatusNotFound)
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

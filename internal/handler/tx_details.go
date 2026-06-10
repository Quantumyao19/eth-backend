package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

type TxDetailResponse struct {
	Hash     string `json:"hash"`
	To       string `json:"to"`
	Value    string `json:"value"`
	GasLimit uint64 `json:"gas_limit"`
	GasUsed  uint64 `json:"gas_used"`
	Status   string `json:"status"`
}

func (h *Handler) TxDetail(w http.ResponseWriter, r *http.Request) {
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

	//obtain transaction
	tx, _, err := h.service.GetTransaction(ctx, common.HexToHash(hash))
	if err != nil {
		log.Println("GetTransaction error:", err)
		handleError(w, err)
		return
	}

	if tx == nil {
		writeError(w, "transaction not found", http.StatusNotFound)
		return
	}

	//obtain receipt
	receipt, err := h.service.GetTransactionReceipt(ctx, common.HexToHash(hash))
	if err != nil {
		log.Println("GetReceipt error:", err)
		handleError(w, err)
		return
	}

	gasUsed := uint64(0)
	status := "pending"

	if receipt != nil {
		gasUsed = receipt.GasUsed
		if receipt.Status == 1 {
			status = "success"
		} else {
			status = "failed"
		}
	}

	to := ""
	if tx.To() != nil {
		to = tx.To().Hex()
	}

	resp := TxDetailResponse{
		Hash:     tx.Hash().Hex(),
		To:       to,
		Value:    tx.Value().String(),
		GasLimit: tx.Gas(),
		GasUsed:  gasUsed,
		Status:   status,
	}

	writeJSON(w, resp)

}

package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
)

type ReceiptResponse struct {
	TxHash  string `json:"tx_hash"`
	Status  uint64 `json:"status"`
	GasUsed uint64 `json:"gas_used"`
}

func (h *Handler) Receipt(w http.ResponseWriter, r *http.Request) {
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
		log.Println("GetReceipt error:", err)
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

package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) BlockNumber(w http.ResponseWriter, r *http.Request) {
	block, err := h.service.GetBlockNumber(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]uint64{
		"block": block,
	})
}

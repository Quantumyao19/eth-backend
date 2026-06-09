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

	writeJSON(w, map[string]uint64{
		"block": block,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

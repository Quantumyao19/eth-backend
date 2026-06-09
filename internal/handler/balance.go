package handler

import (
	"eth-backend/internal/eth"
	"net/http"
)

type Handler struct {
	service *eth.Service
}

func NewHandler(s *eth.Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) Balance(w http.ResponseWriter, r *http.Request) {
	addr := r.URL.Query().Get("address")
	if addr == "" {
		http.Error(w, "missing address", http.StatusBadRequest)
		return
	}

	wei, eth, err := h.service.GetBalance(r.Context(), addr)
	if err != nil {
		http.Error(w, "failed to get balance", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{
		"address":     addr,
		"balance_wei": wei,
		"balance_eth": eth,
	})
}

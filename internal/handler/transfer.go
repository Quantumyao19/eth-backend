package handler

import (
	"eth-backend/internal/repository"
	"net/http"
	"strconv"
)

type TransferHandler struct {
	repo *repository.TransferRepository
}

const (
	defaultPageSize = 20
)

func NewTransferHandler(repo *repository.TransferRepository) *TransferHandler {
	return &TransferHandler{
		repo: repo,
	}
}

func (handler *TransferHandler) ListTransfer(w http.ResponseWriter, r *http.Request) {
	address := r.URL.Query().Get("address")
	if address == "" {
		writeError(w, "address required", http.StatusBadRequest)
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	data, total, err := handler.repo.ListByAddress(r.Context(), address, page, pageSize)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"page":      page,
		"page_size": pageSize,
		"total":     total,
		"data":      data,
	}

	writeJSON(w, resp)

}

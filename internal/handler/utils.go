package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

var ethereumAddressRegex = regexp.MustCompile("^0x[a-fA-F0-9]{40}$")

func isValidEthereumAddress(address string) bool {
	return ethereumAddressRegex.MatchString(address)
}

func writeError(w http.ResponseWriter, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: message,
	})
}

func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func handleError(w http.ResponseWriter, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		writeError(w, "request timeout", http.StatusGatewayTimeout)
		return
	}

	writeError(w, "internal server error", http.StatusInternalServerError)
}

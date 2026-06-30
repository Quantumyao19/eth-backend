package health

import (
	"encoding/json"
	"net/http"
)

func (c *Checker) ReadyHandler(w http.ResponseWriter, r *http.Request) {
	result := c.CheckReadiness(r.Context())

	w.Header().Set("Content-Type", "application/json")

	if result.Status == StatusUnhealthy {
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": string(result.Status),
	})
}

package health

import (
	"encoding/json"
	"net/http"
)

func (c *Checker) StartupHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "started",
	})
}

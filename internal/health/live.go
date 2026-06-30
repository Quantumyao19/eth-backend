package health

import (
	"encoding/json"
	"net/http"
)

func (c *Checker) LiveHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "alive",
	})
}

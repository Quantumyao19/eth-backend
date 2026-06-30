package health

import "net/http"

type HealthHandler struct {
	Checker *Checker
}

func NewHealthHandler(checker *Checker) *HealthHandler {
	return &HealthHandler{
		Checker: checker,
	}
}

func (h *HealthHandler) Ready(w http.ResponseWriter, r *http.Request) {
	h.Checker.ReadyHandler(w, r)
}

func (h *HealthHandler) Live(w http.ResponseWriter, r *http.Request) {
	h.Checker.LiveHandler(w, r)
}

func (h *HealthHandler) Startup(w http.ResponseWriter, r *http.Request) {
	h.Checker.StartupHandler(w, r)
}

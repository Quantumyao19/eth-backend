package health

import "github.com/gin-gonic/gin"

type HealthHandler struct {
	Checker *Checker
}

func NewHealthHandler(checker *Checker) *HealthHandler {
	return &HealthHandler{
		Checker: checker,
	}
}

func (h *HealthHandler) Ready(c *gin.Context) {
	h.Checker.ReadyHandler(c)
}

func (h *HealthHandler) Live(c *gin.Context) {
	h.Checker.LiveHandler(c)
}

func (h *HealthHandler) Startup(c *gin.Context) {
	h.Checker.StartupHandler(c)
}

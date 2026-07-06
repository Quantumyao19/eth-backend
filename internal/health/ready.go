package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Checker) ReadyHandler(ctx *gin.Context) {
	result := c.CheckReadiness(ctx.Request.Context())

	if result.Status == StatusUnhealthy {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{"status": string(result.Status), "score": result.Score})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"status": string(result.Status), "score": result.Score})
}

package health

import "github.com/gin-gonic/gin"

func (c *Checker) ReadyHandler(ctx *gin.Context) {
	result := c.CheckReadiness(ctx.Request.Context())

	if result.Status == StatusUnhealthy {
		ctx.JSON(503, gin.H{"status": string(result.Status), "score": result.Score})
		return
	}

	ctx.JSON(200, gin.H{"status": string(result.Status), "score": result.Score})
}

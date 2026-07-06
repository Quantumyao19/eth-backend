package health

import "github.com/gin-gonic/gin"

func (c *Checker) StartupHandler(ctx *gin.Context) {
	ctx.JSON(200, gin.H{"status": "started"})
}

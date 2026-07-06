package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (c *Checker) StartupHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"status": "started"})
}

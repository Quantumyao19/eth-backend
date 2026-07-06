package handler

import (
	"context"
	"errors"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Data interface{} `json:"data"`
}

var ethereumAddressRegex = regexp.MustCompile("^0x[a-fA-F0-9]{40}$")

func isValidEthereumAddress(address string) bool {
	return ethereumAddressRegex.MatchString(address)
}

func writeError(c *gin.Context, message string, status int) {
	c.JSON(status, ErrorResponse{Error: message})
}

func writeJSON(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, SuccessResponse{Data: data})
}

func handleError(c *gin.Context, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		writeError(c, "request timeout", http.StatusGatewayTimeout)
		return
	}

	writeError(c, "internal server error", http.StatusInternalServerError)
}

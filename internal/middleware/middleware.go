package middleware

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/metrics"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	RouteKey     contextKey = "route"
)

func WithRequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.New().String()

		ctx := context.WithValue(c.Request.Context(), RequestIDKey, requestID)

		c.Request = c.Request.WithContext(ctx)
		c.Set(RequestIDKey.String(), requestID)
		c.Next()
	}
}

func WithRouteContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		c.Set(string(RouteKey), route)
		c.Next()
	}
}

func WithMetrics(m *metrics.Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status := strconv.Itoa(c.Writer.Status())

		route := c.FullPath()
		if route == "" {
			route = "unknown"
		}

		m.HTTPRequestsTotal.WithLabelValues(c.Request.Method, route, status).Inc()
		m.HTTPRequestsDuration.WithLabelValues(c.Request.Method, route, status).Observe(duration)
		if c.Writer.Status() >= 500 {
			m.HTTPRequestsErrors.WithLabelValues(c.Request.Method, route, status)
		}
	}
}

func WithLogging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Milliseconds()

		requestID, _ := c.Request.Context().Value(RequestIDKey).(string)

		logger.Log.Info("http request",
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", duration),
		)
	}
}

func WithRecover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Log.Error("panic recovered", zap.Any("error", err))

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			}
		}()
		c.Next()
	}
}

func (k contextKey) String() string {
	return string(k)
}

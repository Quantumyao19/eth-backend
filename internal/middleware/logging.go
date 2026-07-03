package middleware

import (
	"eth-backend/internal/logger"
	"net/http"
	"time"

	"go.uber.org/zap"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, //default to 200
		}

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Milliseconds()

		requestID, _ := r.Context().Value(RequestIDKey).(string)

		logger.Log.Info("http request",
			zap.String("request_id", requestID),
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.statusCode),
			zap.Int64("duration_ms", duration),
		)
	})
}

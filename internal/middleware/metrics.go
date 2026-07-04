package middleware

import (
	"eth-backend/internal/metrics"
	"net/http"
	"strconv"
	"time"
)

func Metrics(m *metrics.Metrics) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{
				ResponseWriter: w,
				statusCode:     http.StatusOK, // default to 200
			}

			start := time.Now()
			next.ServeHTTP(rw, r)
			duration := time.Since(start).Seconds()

			status := strconv.Itoa(rw.statusCode)

			path := GetRoute(r)

			m.HTTPRequestsTotal.WithLabelValues(r.Method, path, status).Inc()
			m.HTTPRequestDuration.WithLabelValues(r.Method, path, status).Observe(duration)
		})
	}
}

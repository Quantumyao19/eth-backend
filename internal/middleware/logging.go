package middleware

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

type LogEntry struct {
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, //default to 200
		}

		next.ServeHTTP(rw, r)

		entry := LogEntry{
			Method:     r.Method,
			Path:       r.URL.Path,
			Status:     rw.statusCode,
			DurationMs: time.Since(start).Milliseconds(),
		}

		jsonData, err := json.Marshal(entry)
		if err != nil {
			log.Printf("failed to marshal log entry: %v", err)
			return
		}
		log.Println(string(jsonData))

	})
}

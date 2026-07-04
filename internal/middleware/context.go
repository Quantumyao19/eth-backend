package middleware

import (
	"net/http"
)

type contextKey string

const (
	RequestIDKey contextKey = "request_id"
	RouteKey     contextKey = "route"
)

func GetRoute(r *http.Request) string {
	if v := r.Context().Value(RouteKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return "unknown"
}

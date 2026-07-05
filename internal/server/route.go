package server

import (
	"context"
	"eth-backend/internal/logger"
	"eth-backend/internal/middleware"
	"net/http"

	"go.uber.org/zap"
)

func withRoute(route string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Info("withRoute called for route", zap.String("route", route))
		ctx := context.WithValue(r.Context(), middleware.RouteKey, route)
		h(w, r.WithContext(ctx))
	}
}

func routeContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := routeFromPath(r.URL.Path)
		ctx := context.WithValue(r.Context(), middleware.RouteKey, route)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func routeFromPath(path string) string {
	staticRoutes := map[string]string{
		"/balance":        "balance",
		"/block":          "block",
		"/tx":             "tx",
		"/receipt":        "receipt",
		"/tx/detail":      "tx_detail",
		"/health/live":    "health_live",
		"/health/ready":   "health_ready",
		"/health/startup": "health_startup",
		"/metrics":        "metrics",
	}

	if route, ok := staticRoutes[path]; ok {
		return route
	}

	return "unknown"
}

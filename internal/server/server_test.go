package server

import (
	"testing"

	"eth-backend/internal/handler"
	"eth-backend/internal/health"
)

func TestNewServerRegistersGinRoutes(t *testing.T) {
	s := NewServer(&handler.Handler{}, &handler.TransferHandler{}, &health.HealthHandler{}, nil)

	routes := make(map[string]bool)
	for _, route := range s.engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}

	expected := []string{
		"GET /balance",
		"GET /block",
		"GET /tx",
		"GET /receipt",
		"GET /tx/detail",
		"GET /health/live",
		"GET /health/ready",
		"GET /health/startup",
		"GET /metrics",
	}

	for _, route := range expected {
		if !routes[route] {
			t.Fatalf("expected gin route %s to be registered", route)
		}
	}
}

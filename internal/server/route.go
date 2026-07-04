package server

import (
	"context"
	"net/http"
)

type ctxKey string

const routeKey ctxKey = "route"

func withRoute(route string, h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), routeKey, route)
		h(w, r.WithContext(ctx))
	}
}

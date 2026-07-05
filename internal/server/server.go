package server

import (
	"context"
	"eth-backend/internal/handler"
	"eth-backend/internal/health"
	"eth-backend/internal/metrics"
	"eth-backend/internal/middleware"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	handler         *handler.Handler
	transferHandler *handler.TransferHandler
	healthHandler   *health.HealthHandler
	httpServer      *http.Server
	metrics         *metrics.Metrics
}

func NewServer(h *handler.Handler, transferHandler *handler.TransferHandler, healthHandler *health.HealthHandler, metrics *metrics.Metrics) *Server {
	return &Server{
		handler:         h,
		transferHandler: transferHandler,
		healthHandler:   healthHandler,
		metrics:         metrics,
	}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/balance", withRoute("balance", s.handler.Balance))
	mux.HandleFunc("/block", withRoute("block", s.handler.BlockNumber))
	mux.HandleFunc("/tx", withRoute("tx", s.handler.Transaction))
	mux.HandleFunc("/receipt", withRoute("receipt", s.handler.Receipt))
	mux.HandleFunc("/tx/detail", withRoute("tx_detail", s.handler.TxDetail))
	mux.HandleFunc("/transfers", withRoute("transfers", s.transferHandler.ListTransfer))
	mux.HandleFunc("/health/live", withRoute("health_live", s.healthHandler.Live))
	mux.HandleFunc("/health/ready", withRoute("health_ready", s.healthHandler.Ready))
	mux.HandleFunc("/health/startup", withRoute("health_startup", s.healthHandler.Startup))
	mux.HandleFunc("/metrics", withRoute("metrics", promhttp.Handler().ServeHTTP))

	var h http.Handler = mux
	h = routeContext(h)
	h = middleware.Logging(h)
	h = middleware.Metrics(s.metrics)(h)
	h = middleware.Recover(h)
	h = middleware.RequestID(h)

	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: h,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

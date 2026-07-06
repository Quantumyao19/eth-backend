package server

import (
	"context"
	"eth-backend/internal/handler"
	"eth-backend/internal/health"
	"eth-backend/internal/metrics"
	"eth-backend/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	engine          *gin.Engine
	handler         *handler.Handler
	transferHandler *handler.TransferHandler
	healthHandler   *health.HealthHandler
	metrics         *metrics.Metrics
	httpServer      *http.Server
}

func NewServer(h *handler.Handler, transferHandler *handler.TransferHandler, healthHandler *health.HealthHandler, metrics *metrics.Metrics) *Server {
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.WithRequestID())
	r.Use(middleware.WithRouteContext())
	r.Use(middleware.WithMetrics(metrics))
	r.Use(middleware.WithLogging())
	r.Use(middleware.WithRecover())

	s := &Server{
		engine:          r,
		handler:         h,
		transferHandler: transferHandler,
		healthHandler:   healthHandler,
		metrics:         metrics,
	}

	s.registerRoutes()
	return s
}

func (s *Server) registerRoutes() {
	if s.handler != nil {
		s.engine.GET("/balance", s.handler.Balance)
		s.engine.GET("/block", s.handler.BlockNumber)
		s.engine.GET("/tx", s.handler.Transaction)
		s.engine.GET("/receipt", s.handler.Receipt)
		s.engine.GET("/tx/detail", s.handler.TxDetail)
	}
	if s.transferHandler != nil {
		s.engine.GET("/transfers", s.transferHandler.ListTransfer)
	}
	if s.healthHandler != nil {
		s.engine.GET("/health/live", s.healthHandler.Live)
		s.engine.GET("/health/ready", s.healthHandler.Ready)
		s.engine.GET("/health/startup", s.healthHandler.Startup)
	}
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))
}

func (s *Server) Start(port string) error {
	s.httpServer = &http.Server{
		Addr:    ":" + port,
		Handler: s.engine,

		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}

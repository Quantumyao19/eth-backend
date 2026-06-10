package server

import (
	"eth-backend/internal/handler"
	"eth-backend/internal/middleware"
	"net/http"
)

type Server struct {
	handler *handler.Handler
}

func NewServer(h *handler.Handler) *Server {
	return &Server{handler: h}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/balance", s.handler.Balance)
	mux.HandleFunc("/block", s.handler.BlockNumber)
	mux.HandleFunc("/tx", s.handler.Transaction)
	mux.HandleFunc("/receipt", s.handler.Receipt)
	mux.HandleFunc("/tx/detail", s.handler.TxDetail)

	wrappedMux := middleware.Recover(
		middleware.Logging(mux),
	)
	return http.ListenAndServe(":"+port, wrappedMux)
}

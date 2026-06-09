package server

import (
	"eth-backend/internal/handler"
	"net/http"
)

type Server struct {
	handler *handler.Handler
}

func NewServer(h *handler.Handler) *Server {
	return &Server{handler: h}
}

func (s *Server) Start(port string) error {
	http.HandleFunc("/balance", s.handler.Balance)
	http.HandleFunc("/block", s.handler.BlockNumber)
	http.HandleFunc("/tx", s.handler.Transaction)

	return http.ListenAndServe(":"+port, nil)
}

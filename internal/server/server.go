package server

import (
	"eth-backend/internal/handler"
	"eth-backend/internal/middleware"
	"net/http"
)

type Server struct {
	handler         *handler.Handler
	transferHandler *handler.TransferHandler
}

func NewServer(h *handler.Handler, transferHandler *handler.TransferHandler) *Server {
	return &Server{
		handler:         h,
		transferHandler: transferHandler,
	}
}

func (s *Server) Start(port string) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/balance", s.handler.Balance)
	mux.HandleFunc("/block", s.handler.BlockNumber)
	mux.HandleFunc("/tx", s.handler.Transaction)
	mux.HandleFunc("/receipt", s.handler.Receipt)
	mux.HandleFunc("/tx/detail", s.handler.TxDetail)
	mux.HandleFunc("/transfers", s.transferHandler.ListTransfer)

	var h http.Handler = mux
	h = middleware.Logging(h)
	h = middleware.Recover(h)
	h = middleware.RequestID(h)
	return http.ListenAndServe(":"+port, h)
}

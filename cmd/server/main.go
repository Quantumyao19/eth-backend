package main

import (
	"eth-backend/config"
	"eth-backend/internal/eth"
	"eth-backend/internal/handler"
	"eth-backend/internal/server"
	"log"
)

func main() {
	cfg := config.Load()

	client, err := eth.NewClient(cfg.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	service := eth.NewService(client)
	h := handler.NewHandler(service)
	srv := server.NewServer(h)

	log.Println("Server running on port", cfg.Port)
	log.Fatal(srv.Start(cfg.Port))

}

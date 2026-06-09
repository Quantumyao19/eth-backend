package main

import (
	"eth-backend/config"
	"eth-backend/internal/eth"
	"eth-backend/internal/handler"
	"eth-backend/internal/server"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	cfg := config.Load()

	client, err := eth.NewClient(cfg.Eth.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	service := eth.NewService(client)
	h := handler.NewHandler(service)
	srv := server.NewServer(h)

	log.Println("Server running on port", cfg.Server.Port)
	log.Fatal(srv.Start(cfg.Server.Port))
}

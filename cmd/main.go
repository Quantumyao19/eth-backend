package main

import (
	"context"
	"eth-backend/config"
	"eth-backend/internal/eth"
	"eth-backend/internal/handler"
	"eth-backend/internal/listener"
	"eth-backend/internal/logger"
	"eth-backend/internal/server"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger.Init()
	defer logger.Sync()

	logger.Log.Info("starting application")

	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("failed to load .env file", zap.Error(err))
	}

	cfg := config.Load()

	client, err := eth.NewClient(cfg.Eth.RPCURL)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	service, err := eth.NewService(client, cfg.Eth)
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := listener.NewListener(client.Raw())
	l.Start(ctx)

	h := handler.NewHandler(service)
	srv := server.NewServer(h)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		sig := <-sigCh
		logger.Log.Info("shutdown signal received", zap.String("signal", sig.String()))

		cancel()
	}()

	logger.Log.Info("server ready", zap.String("port", cfg.Server.Port))

	if err := srv.Start(cfg.Server.Port); err != nil {
		logger.Log.Fatal("server stopped uunexpectedly", zap.Error(err))
	}
}

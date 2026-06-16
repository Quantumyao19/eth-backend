package main

import (
	"context"
	"database/sql"
	"eth-backend/config"
	"eth-backend/internal/db"
	"eth-backend/internal/eth"
	"eth-backend/internal/handler"
	"eth-backend/internal/listener"
	"eth-backend/internal/logger"
	"eth-backend/internal/repository"
	"eth-backend/internal/server"
	"os"
	"os/signal"
	"syscall"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/jackc/pgx/v5/stdlib"
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
		logger.Log.Fatal("failed to create eth client", zap.Error(err))
	}
	defer client.Close()

	service, err := eth.NewService(client, cfg.Eth)
	if err != nil {
		logger.Log.Fatal("failed to create eth service", zap.Error(err))
	}

	dbPool, err := db.NewPostgres(cfg.DB.Postgres.URL)
	if err != nil {
		logger.Log.Fatal("failed to connect postgres", zap.Error(err))
	}
	defer dbPool.Close()

	sqlDB, err := sql.Open("pgx", cfg.DB.Postgres.URL)
	if err != nil {
		logger.Log.Fatal("failed to open sql DB", zap.Error(err))
	}
	defer sqlDB.Close()

	gdb := goqu.New("postgres", sqlDB)

	transferRepo := repository.NewTransferRepository(gdb)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := listener.NewListener(client.Raw(), transferRepo)
	l.Start(ctx)

	h := handler.NewHandler(service)
	redisClient := db.NewRedisClient(cfg.DB.Redis.Addr, cfg.DB.Redis.Password, cfg.DB.Redis.DB)
	repo := handler.NewTransferHandler(transferRepo, redisClient)
	srv := server.NewServer(h, repo)

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

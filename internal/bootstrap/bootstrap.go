package bootstrap

import (
	"context"
	"database/sql"
	"eth-backend/config"
	"eth-backend/internal/db"
	"eth-backend/internal/eth"
	"eth-backend/internal/handler"
	"eth-backend/internal/health"
	"eth-backend/internal/health/deps"
	"eth-backend/internal/listener"
	"eth-backend/internal/logger"
	"eth-backend/internal/metrics"
	"eth-backend/internal/repository"
	"eth-backend/internal/server"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func Run() error {
	logger.Init()
	defer logger.Sync()

	logger.Log.Info("starting application")

	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("failed to load .env file", zap.Error(err))
	}

	cfg := config.Load()

	dbPool, err := db.NewPostgres(cfg.DB.Postgres.URL)
	if err != nil {
		return err
	}
	defer dbPool.Close()

	sqlDB, err := sql.Open("pgx", cfg.DB.Postgres.URL)
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	gdb := goqu.New("postgres", sqlDB)
	transferRepo := repository.NewTransferRepository(gdb)
	stateRepo := repository.NewListenerStateRepository(gdb)

	redisClient := db.NewRedisClient(cfg.DB.Redis.Addr, cfg.DB.Redis.Password, cfg.DB.Redis.DB)
	defer redisClient.Close()

	m := metrics.NewMetrics()
	client, err := eth.NewClient(cfg.Eth.RPCURL, m)
	if err != nil {
		return err
	}
	defer client.Close()

	service, err := eth.NewService(client, cfg.Eth)
	if err != nil {
		return err
	}

	h := handler.NewHandler(service)
	transferHandler := handler.NewTransferHandler(transferRepo, redisClient)

	engine := health.NewEngine([]health.Dependency{
		deps.NewPostgresDependency(gdb),
		deps.NewRedisDependency(redisClient),
	})

	checker := health.NewChecker(engine)
	healthHandler := health.NewHealthHandler(checker)

	srv := server.NewServer(h, transferHandler, healthHandler, m)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	l := listener.NewListener(client.Raw(), transferRepo, stateRepo, redisClient, m)
	l.Start(ctx)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start(cfg.Server.Port)
	}()

	select {
	case err := <-errCh:
		if err == nil || err == http.ErrServerClosed {
			logger.Log.Info("server stopped gracefully")
			return nil
		}
		return err
	case sig := <-sigCh:
		logger.Log.Info("shutdown signal received", zap.String("signal", sig.String()))
		cancel()

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		if shutdownErr := srv.Shutdown(shutdownCtx); shutdownErr != nil {
			logger.Log.Warn("server shutdown failed", zap.Error(shutdownErr))
		}
		return nil
	}
}

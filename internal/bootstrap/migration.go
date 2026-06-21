package bootstrap

import (
	"database/sql"
	"embed"
	"errors"
	"eth-backend/config"
	"eth-backend/internal/logger"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func RunMigrationCommand() error {
	logger.Init()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("failed to load .env file", zap.Error(err))
	}

	cfg := config.Load()
	return RunMigration(cfg.DB.Postgres.URL)
}

func RunMigration(dbURL string) error {
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		logger.Log.Error("failed to create source driver", zap.Error(err))
		return err
	}

	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance(
		"iofs",
		sourceDriver,
		"postgres",
		dbDriver,
	)
	if err != nil {
		logger.Log.Error("failed to create migrate instance", zap.Error(err))
		return err
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			logger.Log.Warn("failed to close migration source", zap.Error(sourceErr))
		}
		if dbErr != nil {
			logger.Log.Warn("failed to close migration database", zap.Error(dbErr))
		}
	}()

	err = m.Up()
	if err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			logger.Log.Info("no new migrations")
			return nil
		}
		logger.Log.Error("migration failed", zap.Error(err))
		return err
	}

	logger.Log.Info("migration applied successfully")
	return nil
}

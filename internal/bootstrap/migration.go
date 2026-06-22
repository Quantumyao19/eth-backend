package bootstrap

import (
	"database/sql"
	"embed"
	"eth-backend/config"
	"eth-backend/internal/logger"
	"fmt"
	"strconv"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

const (
	dbRetryTimes = 5
)

func RunMigrationCommand(cmd string, arg string) error {
	logger.Init()
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		logger.Log.Warn("failed to load .env file", zap.Error(err))
	}

	cfg := config.Load()
	return RunMigration(cfg.DB.Postgres.URL, cmd, arg)
}

func RunMigration(dbURL, cmd, arg string) error {
	sourceDriver, err := iofs.New(migrationFiles, "migrations")
	if err != nil {
		logger.Log.Error("failed to create source driver", zap.Error(err))
		return err
	}

	db, err := openDBWithRetry(dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	dbDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", sourceDriver, "postgres", dbDriver)
	if err != nil {
		return err
	}
	defer m.Close()

	switch cmd {
	case "up":
		err = m.Up()
		if err != nil && err != migrate.ErrNoChange {
			return err
		}
		logger.Log.Info("migration up done")
	case "down":
		steps := 1
		if arg != "" {
			s, err := strconv.Atoi(arg)
			if err != nil {
				return err
			}
			steps = s
		}
		err = m.Steps(-steps)
		if err != nil {
			return err
		}
		logger.Log.Info("migration down done", zap.Int("steps", steps))

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				logger.Log.Info("no migration applied yet")
				return nil
			}
			return err
		}
		logger.Log.Info("migration version", zap.Uint("version", version), zap.Bool("dirty", dirty))

	default:
		logger.Log.Error("unknown migrate command", zap.String("cmd", cmd))
		return fmt.Errorf("unknown migrate command: %s", cmd)
	}

	return nil
}

func openDBWithRetry(dsn string) (*sql.DB, error) {
	var db *sql.DB
	var err error

	for i := 0; i < dbRetryTimes; i++ {
		db, err = sql.Open("pgx", dsn)
		if err == nil && db.Ping() == nil {
			return db, nil
		}

		time.Sleep(2 * time.Second)
	}
	return nil, err
}

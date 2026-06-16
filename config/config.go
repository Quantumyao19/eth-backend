package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	Eth    EthConfig
	DB     Database
}

type ServerConfig struct {
	Port string
}

type EthConfig struct {
	RPCURL  string
	ChainID string
}

type Database struct {
	Postgres PostgreSQL
	Redis    Redis
}

type PostgreSQL struct {
	URL string
}

type Redis struct {
	Addr     string
	Password string
	DB       int
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Eth: EthConfig{
			RPCURL:  mustGetEnv("RPC_URL"),
			ChainID: getEnv("CHAIN_ID", "11155111"),
		},
		DB: Database{
			Postgres: PostgreSQL{
				URL: getEnv("DB_URL", ""),
			},
			Redis: Redis{
				Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
				Password: getEnv("REDIS_PASSWORD", ""),
				DB:       getEnvInt("REDIS_DB", 0),
			},
		},
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic("missing required env: " + key)
	}
	return v
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if intVal, err := strconv.Atoi(v); err == nil {
			return intVal
		}
	}
	return fallback
}

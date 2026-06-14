package config

import (
	"os"
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
	URL string
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
			URL: getEnv("DB_URL", ""),
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

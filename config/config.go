package config

import "os"

type Config struct {
	RPCURL string
	Port   string
}

func Load() *Config {
	return &Config{
		RPCURL: getEnv("RPC_URL", "https://sepolia.infura.io/v3/xxxxxx"),
		Port:   getEnv("PORT", "8080"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

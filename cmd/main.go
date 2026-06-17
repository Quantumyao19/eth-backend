package main

import (
	"eth-backend/internal/bootstrap"
	"os"
)

func main() {
	if err := bootstrap.Run(); err != nil {
		os.Exit(1)
	}
}

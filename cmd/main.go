package main

import (
	"eth-backend/internal/bootstrap"
	"eth-backend/internal/logger"
	"os"

	"go.uber.org/zap"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		cmd := "up"

		if len(os.Args) > 2 {
			cmd = os.Args[2]
		}

		var arg string
		if len(os.Args) > 3 {
			arg = os.Args[3]
		}

		if err := bootstrap.RunMigrationCommand(cmd, arg); err != nil {
			logger.Log.Error("migration error", zap.Error(err))
			os.Exit(1)
		}
		return
	}

	if err := bootstrap.Run(); err != nil {
		os.Exit(1)
	}
}

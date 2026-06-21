package main

import (
	"eth-backend/internal/bootstrap"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "migrate":
			if err := bootstrap.RunMigrationCommand(); err != nil {
				os.Exit(1)
			}
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
			os.Exit(1)
		}
	}

	if err := bootstrap.Run(); err != nil {
		os.Exit(1)
	}
}

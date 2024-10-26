package main

import (
	"os"

	"github.com/vukovlevi/battleship/logger"
)

func main() {
    log := logger.CreateLogger(os.Stdout, os.Stderr)
    log.Info("lajos", "num", 5, "id")
}

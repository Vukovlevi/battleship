package main

import (
	"os"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/logger"
)

func main() {
    log := logger.CreateLogger(os.Stdout, os.Stderr)
    log.Info("lajos", "num", 5, "id", 76)
    log.Warning("zigi", "num", 5, "id", 77)
    log.Error("budi", "num", 5, "id", 78)

    assert.SetLogger(&log)
    assert.Assert(true, "lajos", "id", 5, "name", "szia")
}

package main

import (
	"flag"
	"os"

	"github.com/vukovlevi/battleship/server/assert"
	"github.com/vukovlevi/battleship/server/game"
	"github.com/vukovlevi/battleship/server/logger"
	"github.com/vukovlevi/battleship/server/tcp"
)

func main() {
    debugMode := false
    flag.BoolVar(&debugMode, "debug", false, "if set to true, the debug statements will appear, otherwise not")
    flag.Parse()

    log := logger.CreateLogger(os.Stdout, os.Stdout, debugMode)
    assert.SetLogger(&log)

    gameServer := game.NewGameServer(&log)
    gameServer.Start()

    tcpServer := tcp.NewTcpServer(42069, &log)
    tcpServer.Start(gameServer.IncomingRequestChan)
}

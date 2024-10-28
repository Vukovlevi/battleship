package assert

import (
	"os"

	"github.com/vukovlevi/battleship/server/logger"
)

var log *logger.Logger = nil

func SetLogger(l *logger.Logger) {
    log = l
}

func Assert(statement bool, msg string, data ...any) {
    if log == nil {
        msg := "logger not set in assert\n"
        os.Stderr.Write([]byte(msg))
        panic(msg)
    }

    if !statement {
        log.Error(msg, data...)
        panic("assert triggered")
    }
}

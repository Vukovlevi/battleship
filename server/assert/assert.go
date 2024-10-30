package assert

import (
	"os"

	"github.com/vukovlevi/battleship/server/logger"
)

const assertErrorMsg = "assert triggered"

var log *logger.Logger = nil

func SetLogger(l *logger.Logger) {
    log = l
}

func checkLogger() {
    if log == nil {
        msg := "logger not set in assert\n"
        os.Stderr.Write([]byte(msg))
        panic(msg)
    }
}

func runAssert(msg string, data []any) {
    log.Error(msg, data...)
    panic(assertErrorMsg)
}

func Assert(statement bool, msg string, data ...any) {
    checkLogger()

    if !statement {
        runAssert(msg, data)
    }
}

func Nil(statement any, msg string, data ...any) {
    checkLogger()

    if statement != nil {
        runAssert(msg, data)
    }
}

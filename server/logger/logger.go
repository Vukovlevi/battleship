package logger

import (
	"fmt"
	"io"
	"os"
	"strconv"
)

const (
    blue = "\033[34m"
    yellow = "\033[33m"
    red = "\033[31m"
    cyan = "\033[36m"
    escapeColor = "\033[0m"
)

type color string

type colorEnum struct {
    blue color
    yellow color
    red color
    cyan color
    escapeColor color
}

type Logger struct {
    writer io.Writer
    errorWriter io.Writer
    debugWriter io.Writer
    color colorEnum
    debugMode bool
}

func CreateLogger(w, d, e io.Writer, debugMode bool) Logger {
    return Logger{
        writer: w,
        errorWriter: e,
        debugWriter: d,
        color: colorEnum{
            blue: blue,
            yellow: yellow,
            red: red,
            cyan: cyan,
            escapeColor: escapeColor,
        },
        debugMode: debugMode,
    }
}

func createMsg(msg string, data ...any) (string, error) {
    if len(data) % 2 != 0 {
        errorMsg := "not correctly formatted data in logger, data len should be even, data len: " + strconv.Itoa(len(data)) + "\n"
        os.Stderr.Write([]byte(errorMsg))
        panic(errorMsg)
    }

    str := msg + "\n"
    for i := 0; i < len(data); i += 2 {
        str += fmt.Sprintf("\t%s: %+v\n", data[i], data[i + 1])
    }

    return str, nil
}

func (l *Logger) write(writer io.Writer, color color, level, msg string) {
    writer.Write([]byte(fmt.Sprintf("%s[%s]%s %s\n", color, level, l.color.escapeColor, msg)))
}

func (l *Logger) Info(msg string, data ...any) {
    str, err := createMsg(msg, data...)
    if err != nil {
        return
    }

    l.write(l.writer, l.color.blue, "INFO", str)
}

func (l *Logger) Debug(msg string, data ...any) {
    if !l.debugMode {
        return
    }

    str, err := createMsg(msg, data...)
    if err != nil {
        return
    }

    l.write(l.debugWriter, l.color.cyan, "DEBUG", str)
}

func (l *Logger) Warning(msg string, data ...any) {
    str, err := createMsg(msg, data...)
    if err != nil {
        return
    }

    l.write(l.writer, l.color.yellow, "WARNING", str)
}

func (l *Logger) Error(msg string, data ...any) {
    str, err := createMsg(msg, data...)
    if err != nil {
        return
    }

    l.write(l.errorWriter, l.color.red, "ERROR", str)
}

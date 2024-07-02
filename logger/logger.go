package logger

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

type LogLevel int

const (
	Silent LogLevel = iota
	Error
	Info
	Trace
)

type ILogger interface {
	Info(string, ...any)
	Trace(sql string, bindings []any, begin time.Time)
	Error(string, ...any)
}

type Logger struct {
	LogLevel LogLevel
}

var _ ILogger = (*Logger)(nil)

func (l Logger) Info(s string, v ...any) {
	if l.LogLevel >= Info {
		log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile).Printf(s, v...)
	}
}

func (l Logger) Error(s string, v ...any) {
	if l.LogLevel >= Error {
		log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile).Printf(s, v...)
	}
}

func (l Logger) Trace(sql string, bindings []any, begin time.Time) {
	if l.LogLevel >= Trace {
		runtime := time.Since(begin)
		sql = fmt.Sprintf(strings.Replace(sql, "?", "%v", -1), bindings...)
		log.New(os.Stdout, "\033[34m[trace]\033[0m ", log.LstdFlags|log.Lshortfile).Printf("[runtime: %v]  "+sql, runtime)
	}
}

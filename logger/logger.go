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
	Info(string, ...interface{})
	Trace(sql string, bindings []interface{}, begin time.Time)
	Error(string, ...interface{})
}

type Logger struct {
	LogLevel LogLevel
}

var _ ILogger = (*Logger)(nil)

func (l Logger) Info(s string, v ...interface{}) {
	if l.LogLevel >= Info {
		log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile).Printf(s, v...)
	}
}

func (l Logger) Error(s string, v ...interface{}) {
	if l.LogLevel >= Error {
		log.New(os.Stdout, "\033[31m[error]\033[0m ", log.LstdFlags|log.Lshortfile).Printf(s, v...)
	}
}

func (l Logger) Trace(sql string, bindings []interface{}, begin time.Time) {
	if l.LogLevel >= Trace {
		runtime := time.Since(begin)
		sql = fmt.Sprintf(strings.Replace(sql, "?", "\"%v\"", -1), bindings...)
		log.New(os.Stdout, "\033[34m[info]\033[0m ", log.LstdFlags|log.Lshortfile).Printf("[runtime: %v]  "+sql, runtime)
	}
}

package logger

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

const (
	fatalExitCode = 1
)

type Logger interface {
	Debug(keyvals ...interface{})
	Info(keyvals ...interface{})
	Warn(keyvals ...interface{})
	Error(keyvals ...interface{})
	Fatal(keyvals ...interface{})
}

type myLogger struct {
	logger log.Logger
}

func (l *myLogger) Debug(keyvals ...interface{}) {
	level.Debug(l.logger).Log(keyvals...)
}

func (l *myLogger) Info(keyvals ...interface{}) {
	level.Info(l.logger).Log(keyvals...)
}

func (l *myLogger) Warn(keyvals ...interface{}) {
	level.Warn(l.logger).Log(keyvals...)
}

func (l *myLogger) Error(keyvals ...interface{}) {
	level.Error(l.logger).Log(keyvals...)
}

func (l *myLogger) Fatal(keyvals ...interface{}) {
	level.Error(l.logger).Log(keyvals...)

	os.Exit(fatalExitCode)
}

func New() Logger {
	var logger log.Logger
	w := log.NewSyncWriter(os.Stderr)

	logger = log.NewLogfmtLogger(w)

	return &myLogger{logger: logger}
}

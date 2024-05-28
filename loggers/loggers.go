package loggers

import (
	"log/slog"
	"os"
)

type logInterface interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
}

type Logger struct {
	log logInterface
}

func GenLogger() *Logger {
	return &Logger{
		log: slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

func (l Logger) Info(msg string, args ...any) {
	l.log.Info(msg, args...)
}

func (l Logger) Debug(msg string, args ...any) {
	l.log.Debug(msg, args...)
}

func (l Logger) Error(msg string, args ...any) {
	l.log.Error(msg, args...)
}

func (l Logger) Warn(msg string, args ...any) {
	l.log.Warn(msg, args...)
}

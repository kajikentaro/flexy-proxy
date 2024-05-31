package loggers

import (
	"fmt"
	"log/slog"
	"os"
)

const (
	ERROR = iota + 1
	WARN
	INFO
	DEBUG
)

type logInterface interface {
	Info(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
}

type Logger struct {
	log   logInterface
	level int
}

type LoggerSettings struct {
	LogLevel int
}

func StrToLogLevel(strLogLevel string) (int, error) {
	switch strLogLevel {
	case "INFO", "info":
		return INFO, nil
	case "DEBUG", "debug":
		return DEBUG, nil
	case "ERROR", "error":
		return ERROR, nil
	case "WARNING", "warning":
		return WARN, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", strLogLevel)
	}
}

func getSlog() *slog.Logger {
	var debugLevel = new(slog.LevelVar)
	debugLevel.Set(slog.LevelDebug)

	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: debugLevel}))
}

func GenLogger(settings *LoggerSettings) *Logger {
	s := &LoggerSettings{
		LogLevel: INFO,
	}
	if settings != nil {
		s = settings
	}

	return &Logger{
		log:   getSlog(),
		level: s.LogLevel,
	}
}

func (l Logger) Info(msg string, args ...any) {
	if INFO <= l.level {
		l.log.Info(msg, args...)
	}
}

func (l Logger) Debug(msg string, args ...any) {
	if DEBUG <= l.level {
		l.log.Debug(msg, args...)
	}
}

func (l Logger) Error(msg string, args ...any) {
	if ERROR <= l.level {
		l.log.Error(msg, args...)
	}
}

func (l Logger) Warn(msg string, args ...any) {
	if WARN <= l.level {
		l.log.Warn(msg, args...)
	}
}

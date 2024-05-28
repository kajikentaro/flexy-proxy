package proxy

import (
	"fmt"
	"go-proxy/loggers"
)

type loggerForProxy struct {
	log *loggers.Logger
}

func (l *loggerForProxy) Printf(format string, v ...interface{}) {
	l.log.Info(fmt.Sprintf(format, v...))
}

func GenLoggerForProxy(log *loggers.Logger) *loggerForProxy {
	return &loggerForProxy{
		log: log,
	}
}

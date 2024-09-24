package proxy

import (
	"fmt"

	"github.com/kajikentaro/flexy-proxy/loggers"
)

type loggerForProxy struct {
	log *loggers.Logger
}

func (l *loggerForProxy) Printf(format string, v ...interface{}) {
	l.log.Debug(fmt.Sprintf(format, v...))
}

func GenLoggerForProxy(log *loggers.Logger) *loggerForProxy {
	return &loggerForProxy{
		log: log,
	}
}

package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/kajikentaro/flexy-proxy/loggers"
	"github.com/kajikentaro/flexy-proxy/proxy"
	"github.com/kajikentaro/flexy-proxy/utils"
	"golang.org/x/sync/errgroup"
)

// default logger used until the config file is loaded
var defaultLogger = loggers.GenLogger(nil)

func StartProxy(customConfigPath string, portNum int) error {
	restart := make(chan struct{})
	serverErr := make(chan error)

	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		return utils.WatchFile(ctx, defaultLogger, customConfigPath, restart)
	})
	eg.Go(func() error {
		server := NewServer(customConfigPath, portNum)
		for {
			select {
			case <-restart:
				go func() {
					serverErr <- server.ResetAndServe()
				}()
			case err := <-serverErr:
				if errors.Is(err, ErrConfig) {
					defaultLogger.Error("Failed to parse config. Waiting for changes...", "error", err)
					continue
				}
				if err != nil {
					return err
				}
				defaultLogger.Info("Server stopped successfully")
				continue
			case <-ctx.Done():
				return nil
			}
		}
	})
	restart <- struct{}{} // initial start
	return eg.Wait()
}

type Server struct {
	configPath string
	portNum    int
	srv        *http.Server
}

var ErrConfig = errors.New("config error")

func NewServer(customConfigPath string, portNum int) *Server {
	return &Server{
		configPath: customConfigPath,
		portNum:    portNum,
		srv:        nil,
	}
}

func (s *Server) close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return s.srv.Shutdown(ctx)
}

// start server. if the server is already running, it will be closed and restarted
func (s *Server) ResetAndServe() error {
	if s.srv != nil {
		if err := s.close(); err != nil {
			return fmt.Errorf("failed to close server: %w", err)
		}
	}

	proxyConfig, err := utils.ParseConfig(s.configPath)
	if err != nil {
		return ErrConfig
	}

	proxy := proxy.NewProxy(proxyConfig)
	addr := fmt.Sprintf(":%d", s.portNum)
	s.srv = &http.Server{Addr: addr, Handler: proxy}
	proxyConfig.Logger.Info(fmt.Sprintf("Starting proxy on %s", addr))

	err = s.srv.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

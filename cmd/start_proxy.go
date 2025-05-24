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
	defer close(restart)

	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(func() error {
		return utils.WatchFile(ctx, defaultLogger, customConfigPath, restart)
	})
	eg.Go(func() error {
		restart <- struct{}{} // initial start
		return nil
	})
	eg.Go(func() error {
		var server *server

		for {
			select {
			case <-restart:
				if server != nil {
					if err := server.Close(); err != nil {
						return fmt.Errorf("failed to stop server: %w", err)
					}
				}

				var err error
				server, err = newServer(customConfigPath, portNum)
				if errors.Is(err, errConfig) {
					continue
				}
				if err != nil {
					return err
				}

				eg.Go(func() error {
					return server.Start()
				})
			case <-ctx.Done():
				return nil
			}
		}
	})
	return eg.Wait()
}

type server struct {
	*http.Server
}

var errConfig = errors.New("config error")

func newServer(customConfigPath string, portNum int) (*server, error) {
	proxyConfig, err := utils.ParseConfig(customConfigPath)
	if err != nil {
		defaultLogger.Error("Failed to parse config. Waiting for changes...", "error", err)
		return nil, errConfig
	}

	proxy := proxy.NewProxy(proxyConfig)
	addr := fmt.Sprintf(":%d", portNum)
	srv := &server{&http.Server{Addr: addr, Handler: proxy}}
	proxyConfig.Logger.Info(fmt.Sprintf("Starting proxy on %s", addr))
	return srv, nil
}

func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}
	return s.Server.Close()
}

func (s *server) Start() error {
	err := s.Server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	if err != nil {
		return err
	}
	return nil
}

package utils

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kajikentaro/flexy-proxy/loggers"
)

func WatchFile(ctx context.Context, log *loggers.Logger, targetFilePath string, restart chan<- struct{}) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to start file watcher: %w", err)
	}
	defer watcher.Close()

	err = watcher.Add(filepath.Dir(targetFilePath))
	if err != nil {
		return fmt.Errorf("failed to add directory to watcher: %w", err)
	}

	debounce := NewDebounce(500 * time.Millisecond)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				// channel was closed
				return fmt.Errorf("watcher events channel closed")
			}
			if filepath.Base(event.Name) == filepath.Base(targetFilePath) {
				debounce(func() {
					log.Warn("File changed, restarting server...")
					select {
					case restart <- struct{}{}:
					default:
						log.Warn("Restart already pending. Skipping signal.")
					}
				})
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return fmt.Errorf("watcher errors channel closed")
			}
			return fmt.Errorf("watcher error: %w", err)
		case <-ctx.Done():
			return nil
		}
	}
}

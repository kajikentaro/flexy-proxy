package utils

import (
	"sync"
	"time"
)

func NewDebounce(delay time.Duration) func(func()) {
	var mu sync.Mutex
	var timer *time.Timer

	return func(f func()) {
		mu.Lock()
		defer mu.Unlock()

		if timer != nil {
			timer.Stop()
		}

		timer = time.AfterFunc(delay, f)
	}
}

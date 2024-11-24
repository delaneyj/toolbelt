package toolbelt

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Throttle will only allow the function to be called once every d duration.
func Throttle(d time.Duration, fn CtxErrFunc) CtxErrFunc {
	shouldWait := false
	mu := &sync.RWMutex{}

	checkShoulWait := func() bool {
		mu.RLock()
		defer mu.RUnlock()
		return shouldWait
	}

	return func(ctx context.Context) error {
		if checkShoulWait() {
			return nil
		}

		mu.Lock()
		defer mu.Unlock()
		shouldWait = true

		go func() {
			<-time.After(d)
			shouldWait = false
		}()

		if err := fn(ctx); err != nil {
			return fmt.Errorf("throttled function failed: %w", err)
		}

		return nil
	}
}

// Debounce will only call the function after d duration has passed since the last call.
func Debounce(d time.Duration, fn CtxErrFunc) CtxErrFunc {
	var t *time.Timer
	mu := &sync.RWMutex{}

	return func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()

		if t != nil && !t.Stop() {
			<-t.C
		}

		t = time.AfterFunc(d, func() {
			if err := fn(ctx); err != nil {
				fmt.Printf("debounced function failed: %v\n", err)
			}
		})

		return nil
	}
}

func CallNTimesWithDelay(d time.Duration, n int, fn CtxErrFunc) CtxErrFunc {
	return func(ctx context.Context) error {
		called := 0
		for {
			shouldCall := false
			if n < 0 {
				shouldCall = true
			} else if called < n {
				shouldCall = true
			}
			if !shouldCall {
				break
			}

			if err := fn(ctx); err != nil {
				return fmt.Errorf("call n times with delay failed: %w", err)
			}
			called++

			<-time.After(d)
		}

		return nil
	}
}

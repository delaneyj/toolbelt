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

// DebounceWithMaxWait creates a debounced function that waits for a quiet period
// before executing, but guarantees execution within a maximum wait time.
func DebounceWithMaxWait(waitTime time.Duration, maxWaitTime time.Duration, fn func(context.Context) error) func(context.Context) error {
	var (
		mu          sync.Mutex
		timer       *time.Timer
		maxTimer    *time.Timer
		latestCtx   context.Context
		firstCallAt time.Time
	)

	execute := func() {
		mu.Lock()
		ctx := latestCtx
		mu.Unlock()

		if ctx != nil {
			fn(ctx)
		}

		mu.Lock()
		timer = nil
		maxTimer = nil
		latestCtx = nil
		firstCallAt = time.Time{}
		mu.Unlock()
	}

	return func(ctx context.Context) error {
		mu.Lock()
		defer mu.Unlock()

		latestCtx = ctx

		// First call in this burst
		if firstCallAt.IsZero() {
			firstCallAt = time.Now()
			
			// Start max wait timer
			maxTimer = time.AfterFunc(maxWaitTime, execute)
		}

		// Reset debounce timer
		if timer != nil {
			timer.Stop()
		}
		timer = time.AfterFunc(waitTime, func() {
			mu.Lock()
			if maxTimer != nil {
				maxTimer.Stop()
			}
			mu.Unlock()
			execute()
		})

		return nil
	}
}

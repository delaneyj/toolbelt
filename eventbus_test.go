package toolbelt

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"errors"
)

func TestEventBusSync_Subscribe(t *testing.T) {
	bus := NewEventBusSync[string]()

	t.Run("single subscriber", func(t *testing.T) {
		ctx := context.Background()
		received := false
		cancel := bus.Subscribe(ctx, func(msg string) error {
			received = true
			if msg != "test" {
				t.Errorf("expected 'test', got '%s'", msg)
			}
			return nil
		})
		defer cancel()

		if bus.Count() != 1 {
			t.Errorf("expected 1 subscriber, got %d", bus.Count())
		}

		err := bus.Emit(ctx, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !received {
			t.Error("subscriber did not receive message")
		}
	})

	t.Run("multiple subscribers", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[int]()
		var sum int

		cancel1 := bus.Subscribe(ctx, func(msg int) error {
			sum += msg
			return nil
		})
		defer cancel1()

		cancel2 := bus.Subscribe(ctx, func(msg int) error {
			sum += msg * 2
			return nil
		})
		defer cancel2()

		if bus.Count() != 2 {
			t.Errorf("expected 2 subscribers, got %d", bus.Count())
		}

		err := bus.Emit(ctx, 10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if sum != 30 { // 10 + 20
			t.Errorf("expected sum 30, got %d", sum)
		}
	})
}

func TestEventBusSync_Unsubscribe(t *testing.T) {
	ctx := context.Background()
	bus := NewEventBusSync[string]()

	var count int
	cancel1 := bus.Subscribe(ctx, func(msg string) error {
		count++
		return nil
	})

	cancel2 := bus.Subscribe(ctx, func(msg string) error {
		count++
		return nil
	})

	if bus.Count() != 2 {
		t.Errorf("expected 2 subscribers, got %d", bus.Count())
	}

	cancel1()

	if bus.Count() != 1 {
		t.Errorf("expected 1 subscriber after unsubscribe, got %d", bus.Count())
	}

	count = 0
	err := bus.Emit(ctx, "test")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	cancel2()

	if bus.Count() != 0 {
		t.Errorf("expected 0 subscribers after all unsubscribe, got %d", bus.Count())
	}
}

func TestSubscriptionCancellation(t *testing.T) {
	t.Run("cancel same subscription multiple times", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[string]()

		cancel := bus.Subscribe(ctx, func(msg string) error {
			return nil
		})

		if bus.Count() != 1 {
			t.Errorf("expected 1 subscriber, got %d", bus.Count())
		}

		// Cancel multiple times - should be idempotent
		cancel()
		cancel()
		cancel()

		if bus.Count() != 0 {
			t.Errorf("expected 0 subscribers after cancel, got %d", bus.Count())
		}
	})

	t.Run("cancel between emits", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[int]()

		var order []int
		var cancel1, cancel2, cancel3 context.CancelFunc

		cancel1 = bus.Subscribe(ctx, func(msg int) error {
			order = append(order, 1)
			return nil
		})

		cancel2 = bus.Subscribe(ctx, func(msg int) error {
			order = append(order, 2)
			return nil
		})

		cancel3 = bus.Subscribe(ctx, func(msg int) error {
			order = append(order, 3)
			return nil
		})

		defer cancel1()
		defer cancel3()

		if bus.Count() != 3 {
			t.Errorf("expected 3 subscribers, got %d", bus.Count())
		}

		// First emit - all three should run
		err := bus.Emit(ctx, 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if len(order) != 3 {
			t.Errorf("expected 3 executions, got %d", len(order))
		}

		// Cancel subscriber 2 between emits
		cancel2()

		// Reset and emit again
		order = nil
		err = bus.Emit(ctx, 2)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Only subscribers 1 and 3 should run
		if len(order) != 2 || order[0] != 1 || order[1] != 3 {
			t.Errorf("expected [1, 3], got %v", order)
		}

		if bus.Count() != 2 {
			t.Errorf("expected 2 subscribers remaining, got %d", bus.Count())
		}
	})

	t.Run("cancel all then subscribe new", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[string]()
		cancels := make([]context.CancelFunc, 5)

		// Add 5 subscribers
		for i := 0; i < 5; i++ {
			idx := i
			cancels[i] = bus.Subscribe(ctx, func(msg string) error {
				t.Logf("Subscriber %d received: %s", idx, msg)
				return nil
			})
		}

		if bus.Count() != 5 {
			t.Errorf("expected 5 subscribers, got %d", bus.Count())
		}

		// Cancel all
		for _, cancel := range cancels {
			cancel()
		}

		if bus.Count() != 0 {
			t.Errorf("expected 0 subscribers after cancelling all, got %d", bus.Count())
		}

		// Add new subscriber
		var received bool
		cancel := bus.Subscribe(ctx, func(msg string) error {
			received = true
			return nil
		})
		defer cancel()

		err := bus.Emit(ctx, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if !received {
			t.Error("new subscriber did not receive message")
		}
	})

	t.Run("async bus cancel during concurrent emit", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusAsync[int]()

		// Use channels to coordinate the test
		started := make(chan struct{}, 3)
		proceed := make(chan struct{})

		var cancel2 context.CancelFunc

		bus.Subscribe(ctx, func(msg int) error {
			started <- struct{}{}
			<-proceed
			return nil
		})

		cancel2 = bus.Subscribe(ctx, func(msg int) error {
			started <- struct{}{}
			<-proceed
			return nil
		})

		bus.Subscribe(ctx, func(msg int) error {
			started <- struct{}{}
			<-proceed
			return nil
		})

		// Start emit in goroutine
		go func() {
			bus.Emit(ctx, 1)
		}()

		// Wait for all subscribers to start
		for i := 0; i < 3; i++ {
			<-started
		}

		// Cancel subscriber 2 while emit is in progress
		cancel2()

		// Let all subscribers finish
		close(proceed)

		// Next emit should only have 2 subscribers
		bus.Emit(ctx, 2)

		// Give time for any async operations
		time.Sleep(10 * time.Millisecond)

		if bus.Count() != 2 {
			t.Errorf("expected 2 subscribers after cancel, got %d", bus.Count())
		}
	})

	t.Run("cancel in reverse order", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[int]()
		cancels := make([]context.CancelFunc, 5)
		received := make([]bool, 5)

		// Subscribe 5 handlers
		for i := 0; i < 5; i++ {
			idx := i
			cancels[i] = bus.Subscribe(ctx, func(msg int) error {
				received[idx] = true
				return nil
			})
		}

		// Cancel in reverse order
		for i := 4; i >= 0; i-- {
			cancels[i]()
			expectedCount := i
			if bus.Count() != expectedCount {
				t.Errorf("after cancelling subscriber %d, expected %d subscribers, got %d", i, expectedCount, bus.Count())
			}
		}

		// Emit should not call any subscriber
		clear(received)
		err := bus.Emit(ctx, 1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		for i, r := range received {
			if r {
				t.Errorf("subscriber %d should not have been called", i)
			}
		}
	})
}

func TestEventBusSync_Error(t *testing.T) {
	ctx := context.Background()
	bus := NewEventBusSync[string]()

	expectedErr := errors.New("test error")

	bus.Subscribe(ctx, func(msg string) error {
		return nil
	})

	bus.Subscribe(ctx, func(msg string) error {
		return expectedErr
	})

	err := bus.Emit(ctx, "test")
	if err == nil {
		t.Error("expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("expected error %v, got %v", expectedErr, err)
	}
}

func TestEventBusAsync_Subscribe(t *testing.T) {
	bus := NewEventBusAsync[string]()

	t.Run("single subscriber", func(t *testing.T) {
		ctx := context.Background()
		received := make(chan string, 1)
		cancel := bus.Subscribe(ctx, func(msg string) error {
			received <- msg
			return nil
		})
		defer cancel()

		if bus.Count() != 1 {
			t.Errorf("expected 1 subscriber, got %d", bus.Count())
		}

		err := bus.Emit(ctx, "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		select {
		case msg := <-received:
			if msg != "test" {
				t.Errorf("expected 'test', got '%s'", msg)
			}
		case <-time.After(time.Second):
			t.Error("timeout waiting for message")
		}
	})

	t.Run("multiple subscribers concurrent", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusAsync[int]()
		var sum atomic.Int32
		var wg sync.WaitGroup

		for i := 0; i < 5; i++ {
			wg.Add(1)
			bus.Subscribe(ctx, func(msg int) error {
				defer wg.Done()
				sum.Add(int32(msg))
				return nil
			})
		}

		if bus.Count() != 5 {
			t.Errorf("expected 5 subscribers, got %d", bus.Count())
		}

		err := bus.Emit(ctx, 10)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		wg.Wait()

		if sum.Load() != 50 { // 10 * 5
			t.Errorf("expected sum 50, got %d", sum.Load())
		}
	})
}

func TestEventBusAsync_Error(t *testing.T) {
	ctx := context.Background()
	bus := NewEventBusAsync[string]()

	expectedErr1 := errors.New("error 1")
	expectedErr2 := errors.New("error 2")

	bus.Subscribe(ctx, func(msg string) error {
		return expectedErr1
	})

	bus.Subscribe(ctx, func(msg string) error {
		return nil
	})

	bus.Subscribe(ctx, func(msg string) error {
		return expectedErr2
	})

	err := bus.Emit(ctx, "test")
	if err == nil {
		t.Error("expected error, got nil")
	}

	// Should contain both errors
	errStr := err.Error()
	if !contains(errStr, "error 1") || !contains(errStr, "error 2") {
		t.Errorf("expected both errors in result, got: %v", err)
	}
}

func TestEventBus_ConcurrentSubscribeUnsubscribe(t *testing.T) {
	t.Run("sync bus", func(t *testing.T) {
		bus := NewEventBusSync[int]()
		testConcurrentSubscribeUnsubscribe(t, bus)
	})

	t.Run("async bus", func(t *testing.T) {
		bus := NewEventBusAsync[int]()
		testConcurrentSubscribeUnsubscribe(t, bus)
	})
}

func testConcurrentSubscribeUnsubscribe(t *testing.T, bus EventBus[int]) {
	ctx := context.Background()
	var wg sync.WaitGroup
	cancels := make([]context.CancelFunc, 100)

	// Concurrent subscriptions
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cancel := bus.Subscribe(ctx, func(msg int) error {
				return nil
			})
			cancels[idx] = cancel
		}(i)
	}

	wg.Wait()

	if bus.Count() != 100 {
		t.Errorf("expected 100 subscribers, got %d", bus.Count())
	}

	// Concurrent unsubscriptions
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			cancels[idx]()
		}(i)
	}

	wg.Wait()

	if bus.Count() != 50 {
		t.Errorf("expected 50 subscribers after partial unsubscribe, got %d", bus.Count())
	}
}

func TestEventBus_ContextCancellation(t *testing.T) {
	t.Run("sync bus with manual cancellation", func(t *testing.T) {
		bus := NewEventBusSync[string]()

		var received bool
		cancelFunc := bus.Subscribe(context.Background(), func(msg string) error {
			received = true
			return nil
		})

		// Normal emit
		err := bus.Emit(context.Background(), "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !received {
			t.Error("expected to receive message")
		}

		// Cancel the subscription manually
		cancelFunc()

		// Emit again - subscriber should not be called as it was cancelled
		received = false
		err = bus.Emit(context.Background(), "test")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if received {
			t.Error("cancelled subscriber should not receive message")
		}
	})
}

func TestEventBus_EmptyBus(t *testing.T) {
	t.Run("sync bus", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusSync[string]()
		err := bus.Emit(ctx, "test")
		if err != nil {
			t.Errorf("unexpected error for empty bus: %v", err)
		}
		if bus.Count() != 0 {
			t.Errorf("expected 0 subscribers, got %d", bus.Count())
		}
	})

	t.Run("async bus", func(t *testing.T) {
		ctx := context.Background()
		bus := NewEventBusAsync[string]()
		err := bus.Emit(ctx, "test")
		if err != nil {
			t.Errorf("unexpected error for empty bus: %v", err)
		}
		if bus.Count() != 0 {
			t.Errorf("expected 0 subscribers, got %d", bus.Count())
		}
	})
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}

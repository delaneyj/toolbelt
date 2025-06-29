package toolbelt

import (
	"context"
	"errors"
	"sync"
)

type Subscriber[T any] func(msg T) error

type EventBus[T any] interface {
	Subscribe(ctx context.Context, fn Subscriber[T]) context.CancelFunc
	Emit(ctx context.Context, msg T) error
	Count() int
}

type baseEventBus[T any] struct {
	mu   sync.RWMutex
	subs []*Subscriber[T]
}

func newBaseEventBus[T any]() *baseEventBus[T] {
	return &baseEventBus[T]{
		subs: []*Subscriber[T]{},
	}
}

func (b *baseEventBus[T]) Subscribe(ctx context.Context, fn Subscriber[T]) context.CancelFunc {
	b.mu.Lock()
	defer b.mu.Unlock()

	_, cancelCtx := context.WithCancel(ctx)

	fnPtr := &fn
	b.subs = append(b.subs, fnPtr)
	cancel := func() {
		defer cancelCtx()

		b.mu.Lock()
		defer b.mu.Unlock()

		// Remove the subscriber from the list
		for i, sub := range b.subs {
			if sub == fnPtr {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				break
			}
		}
	}
	return cancel
}

func (b *baseEventBus[T]) Count() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}

type EventBusSync[T any] struct {
	*baseEventBus[T]
}

func NewEventBusSync[T any]() *EventBusSync[T] {
	return &EventBusSync[T]{
		baseEventBus: newBaseEventBus[T](),
	}
}

func (b *EventBusSync[T]) Emit(ctx context.Context, msg T) error {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, sub := range b.subs {
		if err := (*sub)(msg); err != nil {
			return err
		}
	}
	return nil
}

type EventBusAsync[T any] struct {
	*baseEventBus[T]
	errs  []error
	errMu sync.Mutex
	wg    sync.WaitGroup
}

func NewEventBusAsync[T any]() *EventBusAsync[T] {
	return &EventBusAsync[T]{
		baseEventBus: newBaseEventBus[T](),
	}
}

func (b *EventBusAsync[T]) Emit(ctx context.Context, msg T) error {
	// Subs might be modified while we are iterating over them,
	// so we need to copy them first.
	b.mu.RLock()
	subs := make([]*Subscriber[T], len(b.subs))
	copy(subs, b.subs)
	b.mu.RUnlock()

	clear(b.errs)

	b.wg.Add(len(subs))
	for _, sub := range subs {
		go func(sub Subscriber[T]) {
			defer b.wg.Done()
			if err := sub(msg); err != nil {
				b.errMu.Lock()
				b.errs = append(b.errs, err)
				b.errMu.Unlock()
			}
		}(*sub)
	}
	b.wg.Wait()
	return errors.Join(b.errs...)
}

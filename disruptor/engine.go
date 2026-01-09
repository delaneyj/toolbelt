package disruptor

import (
	"errors"
	"io"
	"math"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// -------------------------------------------------------------------------------------------------
// Errors and validation

var ErrMinimumReservationSize = errors.New("the minimum reservation size is 1 slot")

var (
	errMissingWaitStrategy     = errors.New("a wait strategy must be provided")
	errCapacityTooSmall        = errors.New("the capacity must be at least 1")
	errCapacityPowerOfTwo      = errors.New("the capacity be a power of two, e.g. 2, 4, 8, 16")
	errCapacityTooLarge        = errors.New("the capacity exceeds supported limits")
	errMissingConsumers        = errors.New("no consumers have been provided")
	errMissingConsumersInGroup = errors.New("the consumer group does not have any consumers")
	errEmptyConsumer           = errors.New("an empty consumer was specified in the consumer group")
)

func validateConfig[T any](cfg *config[T]) error {
	if cfg.waitStrategy == nil {
		return errMissingWaitStrategy
	}
	if cfg.capacity == 0 {
		return errCapacityTooSmall
	}
	if cfg.capacity&(cfg.capacity-1) != 0 {
		return errCapacityPowerOfTwo
	}
	if cfg.capacity > uint64(math.MaxInt64) {
		return errCapacityTooLarge
	}
	if len(cfg.consumerGroups) == 0 {
		return errMissingConsumers
	}
	for _, group := range cfg.consumerGroups {
		if len(group) == 0 {
			return errMissingConsumersInGroup
		}
		for _, consumer := range group {
			if consumer == nil {
				return errEmptyConsumer
			}
		}
	}
	return nil
}

// -------------------------------------------------------------------------------------------------
// Cursor and barriers

type cursor [8]int64

func newCursor() *cursor {
	var c cursor
	c[0] = defaultCursorValue
	return &c
}

func (c *cursor) Store(value int64) { atomic.StoreInt64(&c[0], value) }
func (c *cursor) Load() int64       { return atomic.LoadInt64(&c[0]) }

const defaultCursorValue = -1

type barrier interface {
	Load() int64
}

type compositeBarrier []*cursor

func newCompositeBarrier(sequences ...*cursor) barrier {
	switch len(sequences) {
	case 0:
		return emptyBarrier{}
	case 1:
		return sequences[0]
	default:
		return compositeBarrier(sequences)
	}
}

func (c compositeBarrier) Load() int64 {
	min := int64(math.MaxInt64)
	for _, seq := range c {
		if value := seq.Load(); value < min {
			min = value
		}
	}
	return min
}

type emptyBarrier struct{}

func (emptyBarrier) Load() int64 { return math.MaxInt64 }

// -------------------------------------------------------------------------------------------------
// Readers

type reader interface {
	Read()
	Close() error
}

func newCompositeReader(readers []reader) reader {
	switch len(readers) {
	case 0:
		return noopReader{}
	case 1:
		return readers[0]
	default:
		return compositeReader(readers)
	}
}

type compositeReader []reader

func (c compositeReader) Read() {
	var wg sync.WaitGroup
	wg.Add(len(c))
	for _, r := range c {
		reader := r
		go func() {
			reader.Read()
			wg.Done()
		}()
	}
	wg.Wait()
}

func (c compositeReader) Close() error {
	var firstErr error
	for _, r := range c {
		if err := r.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

type noopReader struct{}

func (noopReader) Read()        {}
func (noopReader) Close() error { return nil }

type defaultReader[T any] struct {
	state    int64
	current  *cursor
	written  *cursor
	upstream barrier
	waiter   WaitStrategy
	consumer Consumer[T]
	events   Events[T]
	fast     SliceConsumer[T]
}

func newDefaultReader[T any](current, written *cursor, upstream barrier, waiter WaitStrategy, consumer Consumer[T], ring []T, mask uint64) reader {
	var fast SliceConsumer[T]
	if sc, ok := any(consumer).(SliceConsumer[T]); ok {
		fast = sc
	}

	return &defaultReader[T]{
		state:    stateRunning,
		current:  current,
		written:  written,
		upstream: upstream,
		waiter:   waiter,
		consumer: consumer,
		events: Events[T]{
			ring: ring,
			mask: mask,
		},
		fast: fast,
	}
}

func (r *defaultReader[T]) Read() {
	var gateCount, idleCount, lower, upper int64
	current := r.current.Load()

	for {
		lower = current + 1
		upper = r.upstream.Load()

		if lower <= upper {
			if r.fast != nil {
				r.fast.ConsumeSlice(uint64(lower), uint64(upper), r.events.ring, r.events.mask)
			} else {
				r.events.lower = uint64(lower)
				r.events.upper = uint64(upper)
				r.consumer.Consume(&r.events)
			}
			r.current.Store(upper)
			current = upper
		} else if upper = r.written.Load(); lower <= upper {
			gateCount++
			idleCount = 0
			r.waiter.Gate(uint64(gateCount))
		} else if atomic.LoadInt64(&r.state) == stateRunning {
			idleCount++
			gateCount = 0
			r.waiter.Idle(uint64(idleCount))
		} else {
			break
		}
	}

	if closer, ok := r.consumer.(io.Closer); ok {
		_ = closer.Close()
	}
}

func (r *defaultReader[T]) Close() error {
	atomic.StoreInt64(&r.state, stateClosed)
	return nil
}

const (
	stateRunning = iota
	stateClosed
)

// -------------------------------------------------------------------------------------------------
// Writers

// -------------------------------------------------------------------------------------------------
// Wait strategy

// WaitStrategy coordinates how readers wait for writers. Implementations
// can tune the backoff and idling policy to balance latency and CPU usage.
type WaitStrategy interface {
	Gate(count uint64)
	Idle(count uint64)
}

// DefaultWaitStrategy uses short sleeps to avoid monopolising CPU resources
// under contention.
type DefaultWaitStrategy struct{}

// Gate waits briefly when a reader has caught up to a writer.
func (DefaultWaitStrategy) Gate(uint64) { time.Sleep(time.Nanosecond) }

// Idle backs off slightly more when the disruptor is fully idle.
func (DefaultWaitStrategy) Idle(uint64) { time.Sleep(50 * time.Microsecond) }

// PhasedWaitStrategy spins for a configurable number of iterations, then
// yields, and eventually sleeps. This allows tuning for latency-sensitive
// workloads without burning a full CPU.
type PhasedWaitStrategy struct {
	SpinLimit  uint64
	YieldLimit uint64
	Sleep      time.Duration
}

func NewPhasedWaitStrategy(spinLimit, yieldLimit uint64, sleep time.Duration) PhasedWaitStrategy {
	return PhasedWaitStrategy{SpinLimit: spinLimit, YieldLimit: yieldLimit, Sleep: sleep}
}

func (s PhasedWaitStrategy) Gate(count uint64) {
	s.phase(count)
}

func (s PhasedWaitStrategy) Idle(count uint64) {
	s.phase(count)
}

func (s PhasedWaitStrategy) phase(count uint64) {
	if count <= s.SpinLimit {
		if count&63 == 0 {
			runtime.Gosched()
		}
		return
	}
	if count <= s.SpinLimit+s.YieldLimit {
		runtime.Gosched()
		return
	}
	slp := s.Sleep
	if slp <= 0 {
		slp = 50 * time.Microsecond
	}
	time.Sleep(slp)
}

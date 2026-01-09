package disruptor

import (
	"runtime"
	"sync/atomic"
)

const cacheLineSize = 64

type paddedInt64 struct {
	value int64
	_     [cacheLineSize/8 - 1]int64
}

func newPaddedInt64(v int64) paddedInt64 {
	return paddedInt64{value: v}
}

func (p *paddedInt64) Add(delta int64) int64 {
	return atomic.AddInt64(&p.value, delta)
}

func (p *paddedInt64) Load() int64 {
	return atomic.LoadInt64(&p.value)
}

func (p *paddedInt64) Store(v int64) {
	atomic.StoreInt64(&p.value, v)
}

// sequencer coordinates writer reservations and commits.
type sequencer interface {
	next(count int64, gate barrier) (lower, upper int64)
	publish(lower, upper int64)
}

// singleProducerSequencer uses simple atomic operations for a single writer.
type singleProducerSequencer struct {
	capacity int64
	previous paddedInt64
	gate     paddedInt64
	cursor   *cursor
}

func newSingleProducerSequencer(capacity int64, cursor *cursor) *singleProducerSequencer {
	return &singleProducerSequencer{
		capacity: capacity,
		previous: newPaddedInt64(defaultCursorValue),
		gate:     newPaddedInt64(defaultCursorValue),
		cursor:   cursor,
	}
}

func (s *singleProducerSequencer) next(count int64, upstream barrier) (int64, int64) {
	if count <= 0 {
		panic(ErrMinimumReservationSize)
	}
	next := s.previous.Add(count)
	for spin := int64(0); next-s.capacity > s.gate.Load(); spin++ {
		if spin&SpinMask == 0 {
			runtime.Gosched()
		}
		s.gate.Store(upstream.Load())
	}
	upper := next
	lower := upper - (count - 1)
	return lower, upper
}

func (s *singleProducerSequencer) publish(lower, upper int64) {
	s.cursor.Store(upper)
}

type multiSequencer struct {
	capacity    int64
	indexMask   int64
	nextValue   atomic.Int64
	cachedValue atomic.Int64
	cursorValue atomic.Int64
	available   []atomic.Int64
	cursor      *cursor
}

func newMultiSequencer(capacity int64, cursor *cursor) *multiSequencer {
	s := &multiSequencer{
		capacity:  capacity,
		indexMask: capacity - 1,
		available: make([]atomic.Int64, capacity),
		cursor:    cursor,
	}
	s.nextValue.Store(-1)
	s.cachedValue.Store(-1)
	s.cursorValue.Store(-1)
	for i := range s.available {
		s.available[i].Store(-1)
	}
	return s
}

func (s *multiSequencer) next(count int64, gating barrier) (int64, int64) {
	if count <= 0 || count > s.capacity {
		panic("disruptor: reservation count out of range")
	}

	for {
		current := s.nextValue.Load()
		next := current + count
		wrapPoint := next - s.capacity
		cached := s.cachedValue.Load()

		if wrapPoint > cached || cached > current {
			gatingSequence := gating.Load()
			s.cachedValue.Store(gatingSequence)
			if wrapPoint > gatingSequence {
				runtime.Gosched()
				continue
			}
		}

		if s.nextValue.CompareAndSwap(current, next) {
			lower := next - (count - 1)
			return lower, next
		}
	}
}

func (s *multiSequencer) publish(lower, upper int64) {
	for seq := lower; seq <= upper; seq++ {
		idx := seq & s.indexMask
		s.available[idx].Store(seq)
	}
	s.advanceCursor()
}

func (s *multiSequencer) advanceCursor() {
	for {
		current := s.cursorValue.Load()
		next := current + 1

		for s.isAvailable(next) {
			next++
		}

		target := next - 1
		if target == current {
			return
		}

		if s.cursorValue.CompareAndSwap(current, target) {
			s.cursor.Store(target)
			return
		}
	}
}

func (s *multiSequencer) isAvailable(sequence int64) bool {
	if sequence < 0 {
		return false
	}
	idx := sequence & s.indexMask
	return s.available[idx].Load() == sequence
}

package disruptor

// Consumer receives a batch of events produced by the disruptor. Each
// invocation contains the closed range of sequence numbers alongside random
// access to the shared ring buffer.
type Consumer[T any] interface {
	Consume(*Events[T])
}

// SliceConsumer is an optional extension implemented by consumers that prefer
// to work with the raw ring buffer. When provided, the disruptor will invoke
// this method instead of the more ergonomic Consume(*Events[T]) path.
type SliceConsumer[T any] interface {
	ConsumeSlice(lower, upper uint64, ring []T, mask uint64)
}

// ConsumerFunc adapts a plain function so it can be used as a Consumer.
type ConsumerFunc[T any] func(*Events[T])

// Consume invokes the underlying function.
func (f ConsumerFunc[T]) Consume(events *Events[T]) { f(events) }

// Events describes a contiguous range of published items.
type Events[T any] struct {
	lower uint64
	upper uint64
	ring  []T
	mask  uint64
}

// Lower reports the first sequence number in this batch.
func (e *Events[T]) Lower() uint64 { return e.lower }

// Upper reports the last sequence number in this batch.
func (e *Events[T]) Upper() uint64 { return e.upper }

// At returns the entry for the provided absolute sequence number.
func (e *Events[T]) At(sequence uint64) *T {
	return &e.ring[sequence&e.mask]
}

// Range exposes the sequences in this batch as a Go iterator. Consumers can
// range over the returned function just like a slice.
func (e *Events[T]) Range() func(func(uint64, *T) bool) {
	return func(yield func(uint64, *T) bool) {
		for seq := e.lower; seq <= e.upper; seq++ {
			if !yield(seq, &e.ring[seq&e.mask]) {
				return
			}
		}
	}
}

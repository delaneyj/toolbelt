package disruptor

import "fmt"

// SingleProducer exposes the writer and reader coordination primitives for a
// single-writer disruptor. It owns the typed ring buffer used to exchange
// events.
type SingleProducer[T any] struct {
	*baseDisruptor[T]
}

// MultiProducer exposes the disruptor configured for concurrent writers.
type MultiProducer[T any] struct {
	*baseDisruptor[T]
}

// Producer captures the common API shared by both single- and multi-writer
// disruptors. It is useful for helper functions that operate on either variant.
type Producer[T any] interface {
	Entry(sequence uint64) *T
	BufferSize() uint64
	Ring() ([]T, uint64)
	Publish(write func(slot *T)) uint64
	PublishBatch(count uint64, write func(lower, upper uint64, ring []T, mask uint64)) (uint64, uint64)
	Reserve(count uint64) uint64
	ReserveRange(count uint64) (uint64, uint64)
	Commit(lower, upper uint64)
	Read()
	Close() error
}

// NewSingleProducer constructs a single-writer disruptor and panics if
// validation fails.
func NewSingleProducer[T any](options ...any) SingleProducer[T] {
	d, err := NewSingleProducerDisruptor[T](options...)
	if err != nil {
		panic(err)
	}
	return d
}

// NewSingleProducerDisruptor constructs a single-writer disruptor and returns
// any configuration error.
func NewSingleProducerDisruptor[T any](options ...any) (SingleProducer[T], error) {
	base, err := buildDisruptor[T](false, options...)
	if err != nil {
		return SingleProducer[T]{}, err
	}
	return SingleProducer[T]{base}, nil
}

// NewMultiProducer constructs a multi-writer disruptor and panics if
// validation fails.
func NewMultiProducer[T any](options ...any) MultiProducer[T] {
	d, err := NewMultiProducerDisruptor[T](options...)
	if err != nil {
		panic(err)
	}
	return d
}

// NewMultiProducerDisruptor constructs a multi-writer disruptor and returns
// any configuration error.
func NewMultiProducerDisruptor[T any](options ...any) (MultiProducer[T], error) {
	base, err := buildDisruptor[T](true, options...)
	if err != nil {
		return MultiProducer[T]{}, err
	}
	return MultiProducer[T]{base}, nil
}

// baseDisruptor implements the common disruptor behaviour shared between the
// single- and multi-writer variants. The exported types embed this struct so
// that all methods are promoted automatically.
type baseDisruptor[T any] struct {
	reader    reader
	ring      []T
	mask      uint64
	sequencer sequencer
	upstream  barrier
}

func buildDisruptor[T any](multi bool, options ...any) (*baseDisruptor[T], error) {
	cfg := newConfig[T]()
	for _, option := range options {
		switch opt := option.(type) {
		case capacityOption:
			cfg.capacity = opt.value
		case waitOption:
			if opt.value != nil {
				cfg.waitStrategy = opt.value
			}
		case consumerGroupOption[T]:
			cfg.consumerGroups = append(cfg.consumerGroups, opt.consumers)
		case nil:
			// ignore
		default:
			return nil, fmt.Errorf("unsupported option type %T", option)
		}
	}

	if err := validateConfig(cfg); err != nil {
		return nil, err
	}

	ring := make([]T, int(cfg.capacity))
	mask := cfg.capacity - 1

	cursor := newCursor()
	readers, upstream := buildReaders(cfg, cursor, ring, mask)

	var seq sequencer
	if multi {
		seq = newMultiSequencer(int64(cfg.capacity), cursor)
	} else {
		seq = newSingleProducerSequencer(int64(cfg.capacity), cursor)
	}

	return &baseDisruptor[T]{
		reader:    newCompositeReader(readers),
		ring:      ring,
		mask:      mask,
		sequencer: seq,
		upstream:  upstream,
	}, nil
}

func buildReaders[T any](cfg *config[T], writerSequence *cursor, ring []T, mask uint64) ([]reader, barrier) {
	upstream := barrier(writerSequence)
	var readers []reader

	for _, consumerGroup := range cfg.consumerGroups {
		groupSequences := make([]*cursor, 0, len(consumerGroup))
		for _, consumer := range consumerGroup {
			current := newCursor()
			reader := newDefaultReader(current, writerSequence, upstream, cfg.waitStrategy, consumer, ring, mask)
			readers = append(readers, reader)
			groupSequences = append(groupSequences, current)
		}
		upstream = newCompositeBarrier(groupSequences...)
	}

	return readers, upstream
}

// Entry returns a pointer to the event slot corresponding to the provided
// sequence. The caller can populate the slot before invoking Commit.
func (d *baseDisruptor[T]) Entry(sequence uint64) *T {
	return &d.ring[sequence&d.mask]
}

// BufferSize reports the capacity of the underlying ring buffer.
func (d *baseDisruptor[T]) BufferSize() uint64 {
	return uint64(len(d.ring))
}

// Ring returns the live ring backing store and mask. Callers should only write
// to slots they have reserved and must avoid retaining the slice after Close.
func (d *baseDisruptor[T]) Ring() ([]T, uint64) {
	return d.ring, d.mask
}

// Publish reserves a single slot, applies the provided write function, commits
// the sequence, and returns the published sequence number.
func (d *baseDisruptor[T]) Publish(write func(slot *T)) uint64 {
	lower, upper := d.ReserveRange(1)
	write(d.Entry(upper))
	d.Commit(lower, upper)
	return upper
}

// PublishBatch reserves count slots, invokes write with direct ring access,
// commits the reservation, and returns the lower/upper sequence numbers.
func (d *baseDisruptor[T]) PublishBatch(count uint64, write func(lower, upper uint64, ring []T, mask uint64)) (uint64, uint64) {
	lower, upper := d.ReserveRange(count)
	write(lower, upper, d.ring, d.mask)
	d.Commit(lower, upper)
	return lower, upper
}

// Reserve reserves space for count entries and returns the highest sequence.
func (d *baseDisruptor[T]) Reserve(count uint64) uint64 {
	_, upper := d.ReserveRange(count)
	return upper
}

// ReserveRange reserves count entries and returns the inclusive lower/upper
// sequence numbers for the reservation.
func (d *baseDisruptor[T]) ReserveRange(count uint64) (uint64, uint64) {
	if count == 0 {
		panic(ErrMinimumReservationSize)
	}
	capacity := uint64(len(d.ring))
	if count > capacity {
		panic(fmt.Errorf("reserve count %d exceeds capacity %d", count, capacity))
	}
	lower, upper := d.sequencer.next(int64(count), d.upstream)
	return uint64(lower), uint64(upper)
}

// Commit publishes the reserved range to readers.
func (d *baseDisruptor[T]) Commit(lower, upper uint64) {
	d.sequencer.publish(int64(lower), int64(upper))
}

// Read starts the reader loop until Close is called.
func (d *baseDisruptor[T]) Read() {
	d.reader.Read()
}

// Close stops the reader loop.
func (d *baseDisruptor[T]) Close() error {
	return d.reader.Close()
}

const SpinMask = 1024*16 - 1

var (
	_ Producer[any] = (*SingleProducer[any])(nil)
	_ Producer[any] = (*MultiProducer[any])(nil)
)

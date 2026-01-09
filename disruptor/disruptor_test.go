package disruptor_test

import (
	"sort"
	"sync"
	"testing"

	"github.com/delaneyj/toolbelt/disruptor"
)

type collectingConsumer struct {
	mu     sync.Mutex
	values []uint64
}

func (c *collectingConsumer) Consume(events *disruptor.Events[uint64]) {
	c.mu.Lock()
	for seq := events.Lower(); seq <= events.Upper(); seq++ {
		c.values = append(c.values, *events.At(seq))
	}
	c.mu.Unlock()
}

func (c *collectingConsumer) ConsumeSlice(lower, upper uint64, ring []uint64, mask uint64) {
	c.mu.Lock()
	for seq := lower; seq <= upper; seq++ {
		c.values = append(c.values, ring[seq&mask])
	}
	c.mu.Unlock()
}

func (c *collectingConsumer) Values() []uint64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := make([]uint64, len(c.values))
	copy(out, c.values)
	return out
}

func TestMultiProducerSequencer(t *testing.T) {
	const (
		capacity    = 1024
		producers   = 8
		perProducer = 1000
	)

	consumer := &collectingConsumer{}
	d := disruptor.NewMultiProducer[uint64](
		disruptor.WithCapacity(capacity),
		disruptor.WithConsumerGroup[uint64](consumer),
	)

	var readerWG sync.WaitGroup
	readerWG.Add(1)
	go func() {
		defer readerWG.Done()
		d.Read()
	}()

	var producerWG sync.WaitGroup
	producerWG.Add(producers)
	for p := 0; p < producers; p++ {
		base := uint64(p * perProducer)
		go func(base uint64) {
			defer producerWG.Done()
			for n := uint64(0); n < perProducer; n++ {
				value := base + n
				d.Publish(func(slot *uint64) {
					*slot = value
				})
			}
		}(base)
	}

	producerWG.Wait()
	if err := d.Close(); err != nil {
		t.Fatalf("close disruptor: %v", err)
	}
	readerWG.Wait()

	values := consumer.Values()
	expected := producers * perProducer
	if len(values) != expected {
		t.Fatalf("expected %d values, got %d", expected, len(values))
	}

	sort.Slice(values, func(i, j int) bool { return values[i] < values[j] })
	for i, v := range values {
		if v != uint64(i) {
			t.Fatalf("expected value %d at index %d, got %d", i, i, v)
		}
	}
}

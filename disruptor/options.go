package disruptor

type baseConfig struct {
	capacity     uint64
	waitStrategy WaitStrategy
}

type config[T any] struct {
	baseConfig
	consumerGroups [][]Consumer[T]
}

func newConfig[T any]() *config[T] {
	return &config[T]{
		baseConfig: baseConfig{
			waitStrategy: DefaultWaitStrategy{},
		},
	}
}

type capacityOption struct {
	value uint64
}

// WithCapacity sets the ring buffer size. The value must be a power of two.
func WithCapacity(value uint64) capacityOption {
	return capacityOption{value: value}
}

type waitOption struct {
	value WaitStrategy
}

// WithWaitStrategy customises the reader wait strategy.
func WithWaitStrategy(value WaitStrategy) waitOption {
	return waitOption{value: value}
}

type consumerGroupOption[T any] struct {
	consumers []Consumer[T]
}

// WithConsumerGroup registers one or more consumers that will process events
// in parallel. Additional calls create downstream consumer stages.
func WithConsumerGroup[T any](consumers ...Consumer[T]) consumerGroupOption[T] {
	return consumerGroupOption[T]{consumers: consumers}
}

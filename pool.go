package toolbelt

import (
	"sync"
)

// A Pool is a generic wrapper around a sync.Pool.
type Pool[T any] struct {
	pool  sync.Pool
	newFn func() T
	reset func(T) T
}

// New creates a new Pool with the provided new function.
//
// The equivalent sync.Pool construct is "sync.Pool{New: fn}"
func New[T any](fn func() T, opts ...PoolOption[T]) Pool[T] {
	p := Pool[T]{
		pool:  sync.Pool{New: func() interface{} { return fn() }},
		newFn: fn,
	}
	for _, opt := range opts {
		if opt != nil {
			opt(&p)
		}
	}
	return p
}

// PoolOption configures pool behavior.
type PoolOption[T any] func(*Pool[T])

// WithReset configures a reset function applied by GetWithReset.
func WithReset[T any](reset func(T) T) PoolOption[T] {
	return func(p *Pool[T]) {
		p.reset = reset
	}
}

// Get is a generic wrapper around sync.Pool's Get method.
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// GetWithReset returns an item from the pool and applies the optional reset.
func (p *Pool[T]) GetWithReset() T {
	v := p.Get()
	if p.reset != nil {
		v = p.reset(v)
	}
	return v
}

// Clear resets the pool. If keepCapacity is true, the pool is left intact.
func (p *Pool[T]) Clear(keepCapacity bool) {
	if keepCapacity {
		return
	}
	if p.newFn == nil {
		p.pool = sync.Pool{}
		return
	}
	p.pool = sync.Pool{New: func() interface{} { return p.newFn() }}
}

// Put is a generic wrapper around sync.Pool's Put method.
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

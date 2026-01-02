package vecdb

import (
	"errors"
	"math/rand"
	"time"
)

// Metric defines how distances are computed.
type Metric int

const (
	MetricL2Squared Metric = iota
	MetricCosine
)

var (
	ErrIDExists    = errors.New("vecdb: id already exists")
	ErrDimMismatch = errors.New("vecdb: dimension mismatch")
	ErrEmptyVector = errors.New("vecdb: empty vector")
)

type config struct {
	metric         Metric
	m              int
	efConstruction int
	efSearch       int
	rng            *rand.Rand
}

func defaultConfig() config {
	return config{
		metric:         MetricL2Squared,
		m:              16,
		efConstruction: 200,
		efSearch:       50,
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Option configures an index at construction time.
type Option func(*config)

// WithMetric sets the distance metric.
func WithMetric(metric Metric) Option {
	return func(cfg *config) {
		cfg.metric = metric
	}
}

// WithM configures the maximum neighbors per layer for HNSW.
func WithM(m int) Option {
	return func(cfg *config) {
		if m > 0 {
			cfg.m = m
		}
	}
}

// WithEFConstruction sets the efConstruction parameter for HNSW insertions.
func WithEFConstruction(ef int) Option {
	return func(cfg *config) {
		if ef > 0 {
			cfg.efConstruction = ef
		}
	}
}

// WithEFSearch sets the default efSearch parameter for HNSW queries.
func WithEFSearch(ef int) Option {
	return func(cfg *config) {
		if ef > 0 {
			cfg.efSearch = ef
		}
	}
}

// WithSeed sets the random seed for HNSW level generation.
func WithSeed(seed int64) Option {
	return func(cfg *config) {
		cfg.rng = rand.New(rand.NewSource(seed))
	}
}

// WithRNG sets the random source used for HNSW level generation.
func WithRNG(rng *rand.Rand) Option {
	return func(cfg *config) {
		if rng != nil {
			cfg.rng = rng
		}
	}
}

// Result is a nearest-neighbor search result.
type Result[ID comparable] struct {
	ID    ID
	Score float32
}

// WeightedQuery is a query vector scaled by Weight. SearchWeighted normalizes
// weights by the sum of absolute weights. Negative weights are allowed.
type WeightedQuery struct {
	Weight float32
	Vector []float32
}

type searchOptions[ID comparable] struct {
	filter func(id ID) bool
	ef     int
}

// SearchOption configures search behavior.
type SearchOption[ID comparable] func(*searchOptions[ID])

// WithFilter filters candidates by ID.
func WithFilter[ID comparable](filter func(id ID) bool) SearchOption[ID] {
	return func(opts *searchOptions[ID]) {
		opts.filter = filter
	}
}

// WithEF overrides efSearch for a single HNSW query.
func WithEF[ID comparable](ef int) SearchOption[ID] {
	return func(opts *searchOptions[ID]) {
		if ef > 0 {
			opts.ef = ef
		}
	}
}

func applySearchOptions[ID comparable](opts []SearchOption[ID]) searchOptions[ID] {
	out := searchOptions[ID]{}
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return out
}

func copyVector(vector []float32) []float32 {
	out := make([]float32, len(vector))
	copy(out, vector)
	return out
}

package vecdb

import (
	"sort"
	"sync"

	"github.com/chewxy/math32"
	"github.com/viterin/vek/vek32"
)

// Flat is a brute-force in-memory vector index.
type Flat[ID comparable] struct {
	mu     sync.RWMutex
	dim    int
	metric Metric

	columnNames []string

	ids     []ID
	vectors [][]float32
	index   map[ID]int
}

// NewFlat creates a flat index. If dim is zero, the first insert sets the dimension.
func NewFlat[ID comparable](dim int, opts ...Option) *Flat[ID] {
	if dim < 0 {
		panic("vecdb: dim must be >= 0")
	}
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	return &Flat[ID]{
		dim:         dim,
		metric:      cfg.metric,
		columnNames: copyStrings(cfg.columnNames),
		index:       make(map[ID]int),
	}
}

// Len returns the number of stored vectors.
func (f *Flat[ID]) Len() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return len(f.ids)
}

// Dim returns the configured dimension. Zero means unset.
func (f *Flat[ID]) Dim() int {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.dim
}

// Metric returns the configured distance metric.
func (f *Flat[ID]) Metric() Metric {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.metric
}

// ColumnName returns the associated column name for the given dimension (0-based).
func (f *Flat[ID]) ColumnName(dim int) (string, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	if dim < 0 || dim >= f.dim {
		return "", false
	}
	if dim >= len(f.columnNames) {
		return "", false
	}
	name := f.columnNames[dim]
	if name == "" {
		return "", false
	}
	return name, true
}

// SetColumnName sets the associated column name for the given dimension (0-based).
func (f *Flat[ID]) SetColumnName(dim int, name string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if dim < 0 || dim >= f.dim {
		return ErrInvalidColumnIndex
	}
	if err := f.ensureColumnNamesLocked(); err != nil {
		return err
	}
	f.columnNames[dim] = name
	return nil
}

// Add inserts a new vector. Returns ErrIDExists if id already exists.
func (f *Flat[ID]) Add(id ID, vector ...float32) error {
	if len(vector) == 0 {
		return ErrEmptyVector
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if _, ok := f.index[id]; ok {
		return ErrIDExists
	}
	if err := f.ensureDimLocked(len(vector)); err != nil {
		return err
	}
	f.addLocked(id, vector)
	return nil
}

// Upsert inserts or updates a vector.
func (f *Flat[ID]) Upsert(id ID, vector ...float32) error {
	if len(vector) == 0 {
		return ErrEmptyVector
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	if err := f.ensureDimLocked(len(vector)); err != nil {
		return err
	}
	if idx, ok := f.index[id]; ok {
		f.vectors[idx] = copyVector(vector)
		return nil
	}
	f.addLocked(id, vector)
	return nil
}

// Delete removes a vector by id.
func (f *Flat[ID]) Delete(id ID) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	idx, ok := f.index[id]
	if !ok {
		return false
	}
	last := len(f.ids) - 1
	if idx != last {
		f.ids[idx] = f.ids[last]
		f.vectors[idx] = f.vectors[last]
		f.index[f.ids[idx]] = idx
	}
	f.ids = f.ids[:last]
	f.vectors = f.vectors[:last]
	delete(f.index, id)
	return true
}

// Clear removes all vectors from the index. If keepCapacity is true, backing
// storage is retained for reuse.
func (f *Flat[ID]) Clear(keepCapacity bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if keepCapacity {
		f.ids = f.ids[:0]
		f.vectors = f.vectors[:0]
		clear(f.index)
		return
	}
	f.ids = nil
	f.vectors = nil
	f.index = make(map[ID]int)
}

// Vector returns a copy of the vector for an id, if present.
func (f *Flat[ID]) Vector(id ID) ([]float32, bool) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	idx, ok := f.index[id]
	if !ok {
		return nil, false
	}
	return copyVector(f.vectors[idx]), true
}

// Search returns the k closest vectors to query.
func (f *Flat[ID]) Search(k int, query ...float32) []Result[ID] {
	return f.SearchWithOptions(k, query, nil)
}

// SearchWeighted returns the k closest vectors to the weighted query sum,
// normalizing weights by the sum of absolute weights.
func (f *Flat[ID]) SearchWeighted(k int, queries ...WeightedQuery) []Result[ID] {
	return f.SearchWeightedWithOptions(k, queries)
}

// SearchWithOptions returns the k closest vectors to query with options applied.
func (f *Flat[ID]) SearchWithOptions(k int, query []float32, opts ...SearchOption[ID]) []Result[ID] {
	if k <= 0 || len(query) == 0 {
		return nil
	}
	f.mu.RLock()
	defer f.mu.RUnlock()
	if f.dim != 0 && len(query) != f.dim {
		return nil
	}
	if len(f.ids) == 0 {
		return nil
	}
	searchOpts := applySearchOptions(opts)
	var queryNorm float32
	if f.metric == MetricCosine {
		queryNorm = vek32.Norm(query)
	}

	results := make([]Result[ID], 0, len(f.ids))
	for i, id := range f.ids {
		if searchOpts.filter != nil && !searchOpts.filter(id) {
			continue
		}
		vector := f.vectors[i]
		var dist float32
		if f.metric == MetricCosine {
			vectorNorm := vek32.Norm(vector)
			if queryNorm == 0 || vectorNorm == 0 {
				dist = 1
			} else {
				dot := vek32.Dot(query, vector)
				dist = 1 - (dot / (queryNorm * vectorNorm))
			}
		} else {
			d := vek32.Distance(query, vector)
			dist = d * d
		}
		results = append(results, Result[ID]{ID: id, Score: dist})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score < results[j].Score
	})
	if len(results) > k {
		results = results[:k]
	}
	return results
}

// SearchWeightedWithOptions returns the k closest vectors to the weighted query sum with options applied.
func (f *Flat[ID]) SearchWeightedWithOptions(k int, queries []WeightedQuery, opts ...SearchOption[ID]) []Result[ID] {
	if k <= 0 || len(queries) == 0 {
		return nil
	}
	queryDim := len(queries[0].Vector)
	if queryDim == 0 {
		return nil
	}
	if f.dim != 0 && queryDim != f.dim {
		return nil
	}
	combined := make([]float32, queryDim)
	var weightSum float32
	for _, q := range queries {
		if len(q.Vector) != queryDim {
			return nil
		}
		if q.Weight == 0 {
			continue
		}
		weightSum += math32.Abs(q.Weight)
		for i, v := range q.Vector {
			combined[i] += q.Weight * v
		}
	}
	if weightSum > 0 {
		inv := 1 / weightSum
		for i := range combined {
			combined[i] *= inv
		}
	}
	return f.SearchWithOptions(k, combined, opts...)
}

func (f *Flat[ID]) addLocked(id ID, vector []float32) {
	f.ids = append(f.ids, id)
	f.vectors = append(f.vectors, copyVector(vector))
	f.index[id] = len(f.ids) - 1
}

func (f *Flat[ID]) ensureDimLocked(dim int) error {
	if f.dim == 0 {
		f.dim = dim
	}
	if dim != f.dim {
		return ErrDimMismatch
	}
	return f.normalizeColumnNamesLocked()
}

func (f *Flat[ID]) normalizeColumnNamesLocked() error {
	if f.dim == 0 || len(f.columnNames) == 0 {
		return nil
	}
	if len(f.columnNames) > f.dim {
		return ErrColumnNamesMismatch
	}
	if len(f.columnNames) < f.dim {
		names := make([]string, f.dim)
		copy(names, f.columnNames)
		f.columnNames = names
	}
	return nil
}

func (f *Flat[ID]) ensureColumnNamesLocked() error {
	if f.dim == 0 {
		return ErrInvalidColumnIndex
	}
	if len(f.columnNames) == 0 {
		f.columnNames = make([]string, f.dim)
		return nil
	}
	if len(f.columnNames) > f.dim {
		return ErrColumnNamesMismatch
	}
	if len(f.columnNames) < f.dim {
		names := make([]string, f.dim)
		copy(names, f.columnNames)
		f.columnNames = names
	}
	return nil
}

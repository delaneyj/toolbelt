package vecdb

import (
	"container/heap"
	"math/rand"
	"sort"
	"sync"

	"github.com/chewxy/math32"
	tb "github.com/delaneyj/toolbelt"
	"github.com/viterin/vek/vek32"
)

// HNSW is an approximate in-memory vector index.
type HNSW[ID comparable] struct {
	mu sync.RWMutex

	dim    int
	metric Metric

	m              int
	efConstruction int
	efSearch       int
	rng            *rand.Rand
	columnNames    []string

	entry         int
	maxLevel      int
	nodes         []hnswNode[ID]
	index         map[ID]int
	candidatePool *tb.Pool[[]candidate]
	visitedPool   *tb.Pool[map[int]struct{}]
}

type hnswNode[ID comparable] struct {
	id      ID
	vector  []float32
	norm    float32
	level   int
	links   [][]int
	deleted bool
}

// NewHNSW creates a new HNSW index. If dim is zero, the first insert sets the dimension.
// Type parameters:
//   - ID: a comparable identifier used as the primary key for update/delete/lookup.
func NewHNSW[ID comparable](dim int, opts ...Option) *HNSW[ID] {
	if dim < 0 {
		panic("vecdb: dim must be >= 0")
	}
	cfg := defaultConfig()
	for _, opt := range opts {
		if opt != nil {
			opt(&cfg)
		}
	}
	if cfg.efConstruction < cfg.m {
		cfg.efConstruction = cfg.m
	}
	return &HNSW[ID]{
		dim:            dim,
		metric:         cfg.metric,
		m:              cfg.m,
		efConstruction: cfg.efConstruction,
		efSearch:       cfg.efSearch,
		rng:            cfg.rng,
		columnNames:    copyStrings(cfg.columnNames),
		entry:          -1,
		index:          make(map[ID]int),
		candidatePool:  newCandidatePool(cfg.efConstruction),
		visitedPool:    newVisitedPool(cfg.efConstruction),
	}
}

// Len returns the number of live vectors.
func (h *HNSW[ID]) Len() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.index)
}

// Dim returns the configured dimension. Zero means unset.
func (h *HNSW[ID]) Dim() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.dim
}

// Metric returns the configured distance metric.
func (h *HNSW[ID]) Metric() Metric {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.metric
}

// ColumnName returns the associated column name for the given dimension (0-based).
func (h *HNSW[ID]) ColumnName(dim int) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if dim < 0 || dim >= h.dim {
		return "", false
	}
	if dim >= len(h.columnNames) {
		return "", false
	}
	name := h.columnNames[dim]
	if name == "" {
		return "", false
	}
	return name, true
}

// ColumnNames returns a copy of the associated column names, indexed by dimension (0-based).
// Unset names are returned as empty strings.
func (h *HNSW[ID]) ColumnNames() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.dim == 0 {
		return copyStrings(h.columnNames)
	}
	if len(h.columnNames) == 0 {
		return make([]string, h.dim)
	}
	if len(h.columnNames) < h.dim {
		names := make([]string, h.dim)
		copy(names, h.columnNames)
		return names
	}
	if len(h.columnNames) > h.dim {
		names := make([]string, h.dim)
		copy(names, h.columnNames[:h.dim])
		return names
	}
	return copyStrings(h.columnNames)
}

// SetColumnNames replaces all associated column names, indexed by dimension (0-based).
func (h *HNSW[ID]) SetColumnNames(names ...string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.columnNames = copyStrings(names)
	if h.dim == 0 {
		return nil
	}
	return h.normalizeColumnNamesLocked()
}

// SetColumnName sets the associated column name for the given dimension (0-based).
func (h *HNSW[ID]) SetColumnName(dim int, name string) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if dim < 0 || dim >= h.dim {
		return ErrInvalidColumnIndex
	}
	if err := h.ensureColumnNamesLocked(); err != nil {
		return err
	}
	h.columnNames[dim] = name
	return nil
}

// Add inserts a new vector. Returns ErrIDExists if id already exists.
func (h *HNSW[ID]) Add(id ID, vector ...float32) error {
	if len(vector) == 0 {
		return ErrEmptyVector
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, ok := h.index[id]; ok {
		return ErrIDExists
	}
	if err := h.ensureDimLocked(len(vector)); err != nil {
		return err
	}
	h.addLocked(id, vector)
	return nil
}

// Upsert inserts or updates a vector. Updates are implemented as delete + add.
func (h *HNSW[ID]) Upsert(id ID, vector ...float32) error {
	if len(vector) == 0 {
		return ErrEmptyVector
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureDimLocked(len(vector)); err != nil {
		return err
	}
	if idx, ok := h.index[id]; ok {
		h.nodes[idx].deleted = true
		delete(h.index, id)
	}
	h.addLocked(id, vector)
	return nil
}

// BatchUpsert inserts or updates multiple vectors. Updates are implemented as delete + add.
func (h *HNSW[ID]) BatchUpsert(ids []ID, vectors [][]float32) error {
	if len(ids) != len(vectors) {
		return ErrBatchSizeMismatch
	}
	if len(ids) == 0 {
		return nil
	}
	for _, vector := range vectors {
		if len(vector) == 0 {
			return ErrEmptyVector
		}
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if err := h.ensureDimLocked(len(vectors[0])); err != nil {
		return err
	}
	for _, vector := range vectors {
		if len(vector) != h.dim {
			return ErrDimMismatch
		}
	}
	for i, id := range ids {
		vector := vectors[i]
		if idx, ok := h.index[id]; ok {
			h.nodes[idx].deleted = true
			delete(h.index, id)
		}
		h.addLocked(id, vector)
	}
	return nil
}

// Delete removes a vector by id.
func (h *HNSW[ID]) Delete(id ID) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	idx, ok := h.index[id]
	if !ok {
		return false
	}
	h.nodes[idx].deleted = true
	delete(h.index, id)
	return true
}

// Clear removes all vectors from the index. If keepCapacity is true, backing
// storage is retained for reuse.
func (h *HNSW[ID]) Clear(keepCapacity bool) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.entry = -1
	h.maxLevel = 0
	if keepCapacity {
		h.nodes = h.nodes[:0]
		clear(h.index)
		h.candidatePool.Clear(true)
		h.visitedPool.Clear(true)
	} else {
		h.nodes = nil
		h.index = make(map[ID]int)
		h.candidatePool = newCandidatePool(h.efConstruction)
		h.visitedPool = newVisitedPool(h.efConstruction)
	}
}

// Vector returns a copy of the vector for an id, if present.
func (h *HNSW[ID]) Vector(id ID) ([]float32, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	idx, ok := h.index[id]
	if !ok {
		return nil, false
	}
	return copyVector(h.nodes[idx].vector), true
}

// Search returns the k closest vectors to query.
func (h *HNSW[ID]) Search(k int, query ...float32) []Result[ID] {
	return h.SearchWithOptions(k, query, nil)
}

// SearchWeighted returns the k closest vectors to the weighted query sum,
// normalizing weights by the sum of absolute weights.
func (h *HNSW[ID]) SearchWeighted(k int, queries ...WeightedQuery) []Result[ID] {
	if len(queries) == 0 {
		return nil
	}
	return h.SearchWeightedWithOptions(k, queries)
}

// SearchWithOptions returns the k closest vectors to query with options applied.
func (h *HNSW[ID]) SearchWithOptions(k int, query []float32, opts ...SearchOption[ID]) []Result[ID] {
	if k <= 0 || len(query) == 0 {
		return nil
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.entry < 0 {
		return nil
	}
	if h.dim != 0 && len(query) != h.dim {
		return nil
	}
	searchOpts := applySearchOptions(opts)
	ef := searchOpts.ef
	if ef <= 0 {
		ef = h.efSearch
	}
	if ef < k {
		ef = k
	}

	var queryNorm float32
	if h.metric == MetricCosine {
		queryNorm = vek32.Norm(query)
	}

	entry := h.entry
	for level := h.maxLevel; level > 0; level-- {
		entry = h.greedySearchLayer(query, queryNorm, entry, level)
	}

	candidates := h.searchLayer(query, queryNorm, entry, ef, 0)
	if len(candidates) == 0 {
		return nil
	}

	results := make([]Result[ID], 0, len(candidates))
	for _, cand := range candidates {
		node := h.nodes[cand.idx]
		if node.deleted {
			continue
		}
		if searchOpts.filter != nil && !searchOpts.filter(node.id) {
			continue
		}
		results = append(results, Result[ID]{ID: node.id, Score: cand.dist})
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
func (h *HNSW[ID]) SearchWeightedWithOptions(k int, queries []WeightedQuery, opts ...SearchOption[ID]) []Result[ID] {
	if k <= 0 || len(queries) == 0 {
		return nil
	}
	queryDim := len(queries[0].Vector)
	if queryDim == 0 {
		return nil
	}
	if h.dim != 0 && queryDim != h.dim {
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
	return h.SearchWithOptions(k, combined, opts...)
}

func (h *HNSW[ID]) addLocked(id ID, vector []float32) {
	level := 0
	if h.m > 1 {
		multiplier := float32(1.0) / math32.Log(float32(h.m))
		r := h.rng.Float32()
		if r == 0 {
			r = math32.SmallestNonzeroFloat32
		}
		level = int(-math32.Log(r) * multiplier)
	}
	node := hnswNode[ID]{
		id:     id,
		vector: copyVector(vector),
		level:  level,
		links:  make([][]int, level+1),
	}
	if h.metric == MetricCosine {
		node.norm = vek32.Norm(vector)
	}
	index := len(h.nodes)
	h.nodes = append(h.nodes, node)
	h.index[id] = index

	if h.entry < 0 {
		h.entry = index
		h.maxLevel = level
		return
	}

	var queryNorm float32
	if h.metric == MetricCosine {
		queryNorm = vek32.Norm(vector)
	}

	entry := h.entry
	for l := h.maxLevel; l > level; l-- {
		entry = h.greedySearchLayer(vector, queryNorm, entry, l)
	}

	startLevel := level
	if h.maxLevel < startLevel {
		startLevel = h.maxLevel
	}
	for l := startLevel; l >= 0; l-- {
		candidates := h.searchLayer(vector, queryNorm, entry, h.efConstruction, l)
		maxNeighbors := h.maxNeighbors(l)
		var neighbors []int
		if len(candidates) > 0 && maxNeighbors > 0 {
			if len(candidates) > maxNeighbors {
				candidates = candidates[:maxNeighbors]
			}
			neighbors = make([]int, 0, len(candidates))
			seen := make(map[int]struct{}, len(candidates))
			for _, cand := range candidates {
				if _, ok := seen[cand.idx]; ok {
					continue
				}
				seen[cand.idx] = struct{}{}
				neighbors = append(neighbors, cand.idx)
			}
		}
		h.nodes[index].links[l] = neighbors
		for _, nb := range neighbors {
			links := h.nodes[nb].links[l]
			exists := false
			for _, existing := range links {
				if existing == index {
					exists = true
					break
				}
			}
			if exists {
				continue
			}
			links = append(links, index)
			if len(links) > maxNeighbors && maxNeighbors > 0 {
				pruneCandidates := make([]candidate, 0, len(links))
				base := &h.nodes[nb]
				baseNorm := base.norm
				if h.metric == MetricCosine && baseNorm == 0 {
					baseNorm = vek32.Norm(base.vector)
				}
				for _, neighbor := range links {
					if neighbor == nb {
						continue
					}
					dist := h.distance(base.vector, baseNorm, &h.nodes[neighbor])
					pruneCandidates = append(pruneCandidates, candidate{idx: neighbor, dist: dist})
				}
				sort.Slice(pruneCandidates, func(i, j int) bool {
					return pruneCandidates[i].dist < pruneCandidates[j].dist
				})
				if len(pruneCandidates) > maxNeighbors {
					pruneCandidates = pruneCandidates[:maxNeighbors]
				}
				pruned := make([]int, 0, len(pruneCandidates))
				for _, cand := range pruneCandidates {
					pruned = append(pruned, cand.idx)
				}
				links = pruned
			}
			h.nodes[nb].links[l] = links
		}
		if len(candidates) > 0 {
			entry = candidates[0].idx
		}
	}

	if level > h.maxLevel {
		h.entry = index
		h.maxLevel = level
	}
}

func (h *HNSW[ID]) ensureDimLocked(dim int) error {
	if h.dim == 0 {
		h.dim = dim
	}
	if dim != h.dim {
		return ErrDimMismatch
	}
	return h.normalizeColumnNamesLocked()
}

func (h *HNSW[ID]) normalizeColumnNamesLocked() error {
	if h.dim == 0 || len(h.columnNames) == 0 {
		return nil
	}
	if len(h.columnNames) > h.dim {
		return ErrColumnNamesMismatch
	}
	if len(h.columnNames) < h.dim {
		names := make([]string, h.dim)
		copy(names, h.columnNames)
		h.columnNames = names
	}
	return nil
}

func (h *HNSW[ID]) ensureColumnNamesLocked() error {
	if h.dim == 0 {
		return ErrInvalidColumnIndex
	}
	if len(h.columnNames) == 0 {
		h.columnNames = make([]string, h.dim)
		return nil
	}
	if len(h.columnNames) > h.dim {
		return ErrColumnNamesMismatch
	}
	if len(h.columnNames) < h.dim {
		names := make([]string, h.dim)
		copy(names, h.columnNames)
		h.columnNames = names
	}
	return nil
}

func (h *HNSW[ID]) maxNeighbors(level int) int {
	if level == 0 {
		return h.m * 2
	}
	return h.m
}

func (h *HNSW[ID]) distance(query []float32, queryNorm float32, node *hnswNode[ID]) float32 {
	switch h.metric {
	case MetricCosine:
		if queryNorm == 0 || node.norm == 0 {
			return 1
		}
		dot := vek32.Dot(query, node.vector)
		return 1 - (dot / (queryNorm * node.norm))
	default:
		d := vek32.Distance(query, node.vector)
		return d * d
	}
}

func (h *HNSW[ID]) greedySearchLayer(query []float32, queryNorm float32, entry int, level int) int {
	curr := entry
	currDist := h.distance(query, queryNorm, &h.nodes[curr])
	for {
		changed := false
		for _, nb := range h.nodes[curr].links[level] {
			if h.nodes[nb].deleted {
				continue
			}
			dist := h.distance(query, queryNorm, &h.nodes[nb])
			if dist < currDist {
				curr = nb
				currDist = dist
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	return curr
}

type candidate struct {
	idx  int
	dist float32
}

type minHeap []candidate

func (h minHeap) Len() int            { return len(h) }
func (h minHeap) Less(i, j int) bool  { return h[i].dist < h[j].dist }
func (h minHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *minHeap) Push(x interface{}) { *h = append(*h, x.(candidate)) }
func (h *minHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

type maxHeap []candidate

func (h maxHeap) Len() int            { return len(h) }
func (h maxHeap) Less(i, j int) bool  { return h[i].dist > h[j].dist }
func (h maxHeap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *maxHeap) Push(x interface{}) { *h = append(*h, x.(candidate)) }
func (h *maxHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[:n-1]
	return item
}

func (h maxHeap) worstDist() float32 {
	if len(h) == 0 {
		return math32.Inf(1)
	}
	return h[0].dist
}

func (h *HNSW[ID]) searchLayer(query []float32, queryNorm float32, entry int, ef int, level int) []candidate {
	visited := h.visitedPool.GetWithReset()
	candidateSlice := h.getCandidates()
	resultSlice := h.getCandidates()
	candidates := minHeap(candidateSlice)
	results := maxHeap(resultSlice)
	candidatesPtr := &candidates
	resultsPtr := &results

	entryDist := h.distance(query, queryNorm, &h.nodes[entry])
	heap.Push(candidatesPtr, candidate{idx: entry, dist: entryDist})
	if !h.nodes[entry].deleted {
		heap.Push(resultsPtr, candidate{idx: entry, dist: entryDist})
	}
	visited[entry] = struct{}{}

	for candidatesPtr.Len() > 0 {
		curr := heap.Pop(candidatesPtr).(candidate)
		if resultsPtr.Len() >= ef && curr.dist > results.worstDist() {
			break
		}
		for _, nb := range h.nodes[curr.idx].links[level] {
			if _, ok := visited[nb]; ok {
				continue
			}
			visited[nb] = struct{}{}
			dist := h.distance(query, queryNorm, &h.nodes[nb])
			if resultsPtr.Len() < ef || dist < results.worstDist() {
				heap.Push(candidatesPtr, candidate{idx: nb, dist: dist})
			}
			if h.nodes[nb].deleted {
				continue
			}
			if resultsPtr.Len() < ef || dist < results.worstDist() {
				heap.Push(resultsPtr, candidate{idx: nb, dist: dist})
				if resultsPtr.Len() > ef {
					heap.Pop(resultsPtr)
				}
			}
		}
	}

	out := make([]candidate, len(results))
	copy(out, results)
	sort.Slice(out, func(i, j int) bool { return out[i].dist < out[j].dist })
	h.putCandidates(candidates)
	h.putCandidates(results)
	if len(visited) > maxVisitedPoolEntries {
		h.visitedPool.Put(make(map[int]struct{}))
	} else {
		clear(visited)
		h.visitedPool.Put(visited)
	}
	return out
}

const (
	maxPoolCandidates     = 16384
	maxVisitedPoolEntries = 16384
)

func (h *HNSW[ID]) getCandidates() []candidate {
	return h.candidatePool.GetWithReset()
}

func newCandidatePool(capacity int) *tb.Pool[[]candidate] {
	return tb.New(
		func() []candidate { return make([]candidate, 0, capacity) },
		tb.WithReset(func(c []candidate) []candidate {
			if c == nil {
				return nil
			}
			return c[:0]
		}),
	)
}

func newVisitedPool(capacity int) *tb.Pool[map[int]struct{}] {
	return tb.New(
		func() map[int]struct{} { return make(map[int]struct{}, capacity) },
		tb.WithReset(func(v map[int]struct{}) map[int]struct{} {
			if v == nil {
				return make(map[int]struct{}, capacity)
			}
			clear(v)
			return v
		}),
	)
}

func (h *HNSW[ID]) putCandidates(candidates []candidate) {
	if cap(candidates) > maxPoolCandidates {
		candidates = nil
	} else {
		candidates = candidates[:0]
	}
	h.candidatePool.Put(candidates)
}

# vecdb
In-memory vector indices for Go with Flat (exact) and HNSW (approximate) search.

Features
- Generics over ID types.
- Support for L2 squared and cosine distance.
- Flat and HNSW indices with shared API.
- Tests, fuzzing, and benchmarks included.
- Vectors use `float32` and leverage SIMD via `github.com/viterin/vek/vek32` on supported CPUs.
- No cgo required; SIMD uses Go asm with a pure-Go fallback.

Flat vs HNSW
- Flat is exact search over all vectors (O(n) per query). Use it for small datasets, tight correctness requirements, or when build time must be minimal.
- HNSW is approximate search with faster queries on larger datasets, at the cost of more memory and slower inserts/builds. Use it when you can trade some recall for speed.

Usage
```go
package main

import "github.com/delaneyj/toolbelt/vecdb"

func main() {
	idx := vecdb.NewHNSW[string](2,
		vecdb.WithMetric(vecdb.MetricCosine),
		vecdb.WithEFConstruction(200),
		vecdb.WithEFSearch(50),
	)

	_ = idx.Add("a", 1, 0)
	_ = idx.Add("b", 0, 1)

	results := idx.Search(2, 1, 0)
	_ = results

	weighted := idx.SearchWeighted(2,
		vecdb.WeightedQuery{Weight: 1, Vector: []float32{1, 0}},
		vecdb.WeightedQuery{Weight: -0.25, Vector: []float32{0, 1}},
	)
	_ = weighted
}
```

API highlights
- `Add(id, vector...)` inserts a new vector (ErrIDExists if the id already exists).
- `Upsert(id, vector...)` inserts or updates.
- `Delete(id)` removes by id.
- `Clear(keepCapacity)` removes all vectors, optionally keeping backing storage.
- `Vector(id)` returns a copy of the vector.
- `Search(k, vector...)` returns the k closest neighbors.
- `SearchWithOptions(k, vectorSlice, ...SearchOption)` applies per-query options.
- `SearchWeighted(k, queries...)` searches with weighted query vectors, normalized by the sum of absolute weights (negative weights allowed).

Generics
`NewHNSW[ID](dim, ...Option)` uses a single type parameter:
- `ID`: a comparable identifier used as the primary key for update/delete and lookup.
Vector components are `float32` only.

Example:
```go
// IDs are strings, vectors are float32.
idx := vecdb.NewHNSW[string](384)
```

Options
- `WithMetric`: choose `MetricL2Squared` or `MetricCosine`.
- `WithM`, `WithEFConstruction`, `WithEFSearch`: HNSW tuning.
- `WithSeed` or `WithRNG`: HNSW level generation control.
- `WithFilter`: per-query filter on id.
- `WithEF`: per-query override for HNSW ef.

Notes
- This package is in-memory only (persistence can be layered later).
- Distances are returned as `Score`, lower is better.
- HNSW deletes are tombstones; memory is not compacted.

Benchmarks
Search: `go test ./vecdb -bench=Search -benchmem -benchtime=1s -count=3` on `AMD Ryzen 9 6900HX` (linux/amd64).
Build: `go test ./vecdb -bench=Build -benchmem -benchtime=1s -count=3` on the same machine.
Vector dim = 20, queries = 100, vectors = 20,000 for both indices.
Each benchmark op runs 80 searches for Flat and HNSW, which keeps the slowest op around 1s on the reference machine without SIMD.
Index size is an approximate heap delta after building the index in a fresh process (not per-op).
At dim 20, SIMD results are noisy with only 3 samples; recent runs show Flat roughly unchanged and HNSW faster with `vek32`.
Build time is measured by constructing a new index with 20,000 vectors (same dim/params as above).

| Benchmark | Vectors | time/op | B/op | allocs/op | index heap (overall) |
| --- | --- | --- | --- | --- | --- |
| FlatSearch | 20,000 | 870 ms | 25.01 MiB | 400 | 2.78 MiB |
| HNSWSearch | 20,000 | 59.9 ms | 752 KiB | 38.6k | 7.57 MiB |

| Build Benchmark | Vectors | time/op | B/op | allocs/op |
| --- | --- | --- | --- | --- |
| FlatBuild | 20,000 | 13.0 ms | 5.46 MiB | 20.2k |
| HNSWBuild | 20,000 | 41.3 s | 1.38 GiB | 30.4M |

Tasks
- `task test` run package tests
- `task fuzz` run fuzzers (override with `FUZZTIME=2m`)
- `task bench` run benchmarks

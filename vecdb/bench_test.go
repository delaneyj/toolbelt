package vecdb

import (
	"math/rand"
	"testing"
)

const (
	benchDim      = 20
	benchCount    = 20000
	benchQueries  = 100
	benchSearches = 80
)

func BenchmarkFlatSearch(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	vectors := randomVectors(rng, benchCount, benchDim)
	queries := randomVectors(rng, benchQueries, benchDim)
	idx := NewFlat[int](benchDim)
	for i, vec := range vectors {
		if err := idx.Add(i, vec...); err != nil {
			b.Fatalf("add: %v", err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := 0
		for j := 0; j < benchSearches; j++ {
			idx.Search(10, queries[q]...)
			q++
			if q == len(queries) {
				q = 0
			}
		}
	}
}

func BenchmarkFlatBuild(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	vectors := randomVectors(rng, benchCount, benchDim)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := NewFlat[int](benchDim)
		for j, vec := range vectors {
			if err := idx.Add(j, vec...); err != nil {
				b.Fatalf("add: %v", err)
			}
		}
	}
}

func BenchmarkHNSWSearch(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	vectors := randomVectors(rng, benchCount, benchDim)
	queries := randomVectors(rng, benchQueries, benchDim)
	idx := NewHNSW[int](benchDim,
		WithSeed(1),
		WithEFConstruction(200),
		WithEFSearch(50),
	)
	for i, vec := range vectors {
		if err := idx.Add(i, vec...); err != nil {
			b.Fatalf("add: %v", err)
		}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := 0
		for j := 0; j < benchSearches; j++ {
			idx.Search(10, queries[q]...)
			q++
			if q == len(queries) {
				q = 0
			}
		}
	}
}

func BenchmarkHNSWBuild(b *testing.B) {
	rng := rand.New(rand.NewSource(1))
	vectors := randomVectors(rng, benchCount, benchDim)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		idx := NewHNSW[int](benchDim,
			WithSeed(1),
			WithEFConstruction(200),
			WithEFSearch(50),
		)
		for j, vec := range vectors {
			if err := idx.Add(j, vec...); err != nil {
				b.Fatalf("add: %v", err)
			}
		}
	}
}

func randomVectors(rng *rand.Rand, n int, dim int) [][]float32 {
	out := make([][]float32, n)
	for i := 0; i < n; i++ {
		out[i] = randomVector(rng, dim)
	}
	return out
}

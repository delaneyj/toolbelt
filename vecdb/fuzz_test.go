package vecdb

import (
	"math/rand"
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzFlatSearchSorted(f *testing.F) {
	f.Add(int64(1), 16, 3)
	f.Add(int64(42), 32, 4)

	f.Fuzz(func(t *testing.T, seed int64, n int, dim int) {
		if n <= 0 || n > 64 {
			n = 16
		}
		if dim <= 0 || dim > 8 {
			dim = 3
		}
		rng := rand.New(rand.NewSource(seed))
		metric := MetricL2Squared
		if rng.Intn(2) == 0 {
			metric = MetricCosine
		}
		idx := NewFlat[int](dim, WithMetric(metric))
		for i := 0; i < n; i++ {
			vec := randomVector(rng, dim)
			require.NoError(t, idx.Add(i, vec...))
		}
		query := randomVector(rng, dim)
		k := rng.Intn(n) + 1
		results := idx.Search(k, query...)
		require.LessOrEqual(t, len(results), k)
		for i := 1; i < len(results); i++ {
			require.LessOrEqual(t, results[i-1].Score, results[i].Score)
		}
	})
}

func FuzzHNSWSearchSorted(f *testing.F) {
	f.Add(int64(7), 16, 3)
	f.Add(int64(99), 24, 5)

	f.Fuzz(func(t *testing.T, seed int64, n int, dim int) {
		if n <= 0 || n > 64 {
			n = 16
		}
		if dim <= 0 || dim > 8 {
			dim = 3
		}
		rng := rand.New(rand.NewSource(seed))
		metric := MetricL2Squared
		if rng.Intn(2) == 0 {
			metric = MetricCosine
		}
		idx := NewHNSW[int](dim,
			WithMetric(metric),
			WithSeed(seed),
			WithEFConstruction(64),
			WithEFSearch(64),
		)
		idSet := make(map[int]struct{}, n)
		for i := 0; i < n; i++ {
			vec := randomVector(rng, dim)
			require.NoError(t, idx.Add(i, vec...))
			idSet[i] = struct{}{}
		}
		query := randomVector(rng, dim)
		k := rng.Intn(n) + 1
		results := idx.Search(k, query...)
		require.LessOrEqual(t, len(results), k)
		for i := 1; i < len(results); i++ {
			require.LessOrEqual(t, results[i-1].Score, results[i].Score)
		}
		for _, res := range results {
			_, ok := idSet[res.ID]
			require.True(t, ok)
		}
	})
}

func randomVector(rng *rand.Rand, dim int) []float32 {
	out := make([]float32, dim)
	for i := range out {
		out[i] = rng.Float32()*2 - 1
	}
	return out
}

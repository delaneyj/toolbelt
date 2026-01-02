package vecdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlatBasic(t *testing.T) {
	idx := NewFlat[string](2)

	require.NoError(t, idx.Add("a", 0, 0))
	require.NoError(t, idx.Add("b", 1, 0))
	require.NoError(t, idx.Add("c", 2, 0))

	vec, ok := idx.Vector("c")
	require.True(t, ok)
	require.Equal(t, []float32{2, 0}, vec)

	results := idx.Search(2, 0.5, 0)
	require.Len(t, results, 2)
	require.Equal(t, "a", results[0].ID)
	require.Equal(t, "b", results[1].ID)

	require.True(t, idx.Delete("b"))
	require.False(t, idx.Delete("b"))
	_, ok = idx.Vector("b")
	require.False(t, ok)
}

func TestFlatSetAndFilter(t *testing.T) {
	idx := NewFlat[string](0, WithMetric(MetricCosine))

	require.NoError(t, idx.Upsert("a", 1, 0))
	require.NoError(t, idx.Upsert("b", 0, 1))
	require.NoError(t, idx.Upsert("a", 1, 1))

	results := idx.SearchWithOptions(2, []float32{1, 0}, WithFilter(func(id string) bool {
		return id == "a"
	}))
	require.Len(t, results, 1)
	require.Equal(t, "a", results[0].ID)
}

func TestFlatDimMismatch(t *testing.T) {
	idx := NewFlat[int](2)
	require.NoError(t, idx.Add(1, 0, 1))
	require.ErrorIs(t, idx.Add(2, 0, 1, 2), ErrDimMismatch)
}

func TestFlatSearchWeighted(t *testing.T) {
	idx := NewFlat[string](2)
	require.NoError(t, idx.Add("p", 1, 0))
	require.NoError(t, idx.Add("n", -1, 0))

	results := idx.SearchWeighted(1, WeightedQuery{Weight: -1, Vector: []float32{1, 0}})
	require.Len(t, results, 1)
	require.Equal(t, "n", results[0].ID)
}

func TestFlatSearchWeightedNormalization(t *testing.T) {
	idx := NewFlat[string](2)
	require.NoError(t, idx.Add("near", 1, 0))
	require.NoError(t, idx.Add("far", 3, 0))

	results := idx.SearchWeighted(1, WeightedQuery{Weight: 4, Vector: []float32{1, 0}})
	require.Len(t, results, 1)
	require.Equal(t, "near", results[0].ID)
}

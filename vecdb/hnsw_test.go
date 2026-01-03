package vecdb

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHNSWBasic(t *testing.T) {
	idx := NewHNSW[string](2,
		WithSeed(42),
		WithEFConstruction(64),
		WithEFSearch(64),
	)

	require.NoError(t, idx.Add("a", 0, 0))
	require.NoError(t, idx.Add("b", 1, 0))
	require.NoError(t, idx.Add("c", 2, 0))

	vec, ok := idx.Vector("c")
	require.True(t, ok)
	require.Equal(t, []float32{2, 0}, vec)

	results := idx.Search(2, 0.4, 0)
	require.Len(t, results, 2)
	require.Equal(t, "a", results[0].ID)
	require.Equal(t, "b", results[1].ID)

	require.True(t, idx.Delete("b"))
	_, ok = idx.Vector("b")
	require.False(t, ok)
}

func TestHNSWSet(t *testing.T) {
	idx := NewHNSW[int](0, WithSeed(7))

	require.NoError(t, idx.Upsert(1, 1, 0))
	require.NoError(t, idx.Upsert(2, 0, 1))
	require.NoError(t, idx.Upsert(1, 1, 1))
	require.Equal(t, 2, idx.Len())
}

func TestHNSWColumnName(t *testing.T) {
	idx := NewHNSW[string](2, WithSeed(7), WithColumnNames("embedding", "title"))
	name, ok := idx.ColumnName(0)
	require.True(t, ok)
	require.Equal(t, "embedding", name)
	name, ok = idx.ColumnName(1)
	require.True(t, ok)
	require.Equal(t, "title", name)

	require.NoError(t, idx.SetColumnName(1, "title_embedding"))
	name, ok = idx.ColumnName(1)
	require.True(t, ok)
	require.Equal(t, "title_embedding", name)
}

func TestHNSWDimMismatch(t *testing.T) {
	idx := NewHNSW[int](2)
	require.NoError(t, idx.Add(1, 0, 1))
	require.ErrorIs(t, idx.Add(2, 0, 1, 2), ErrDimMismatch)
}

func TestHNSWSearchWeighted(t *testing.T) {
	idx := NewHNSW[string](2, WithSeed(9), WithEFConstruction(32), WithEFSearch(32))
	require.NoError(t, idx.Add("p", 1, 0))
	require.NoError(t, idx.Add("n", -1, 0))

	results := idx.SearchWeighted(1, WeightedQuery{Weight: -1, Vector: []float32{1, 0}})
	require.Len(t, results, 1)
	require.Equal(t, "n", results[0].ID)
}

func TestHNSWSearchWeightedNormalization(t *testing.T) {
	idx := NewHNSW[string](2, WithSeed(11), WithEFConstruction(32), WithEFSearch(32))
	require.NoError(t, idx.Add("near", 1, 0))
	require.NoError(t, idx.Add("far", 3, 0))

	results := idx.SearchWeighted(1, WeightedQuery{Weight: 4, Vector: []float32{1, 0}})
	require.Len(t, results, 1)
	require.Equal(t, "near", results[0].ID)
}

func TestHNSWClearKeepCapacity(t *testing.T) {
	idx := NewHNSW[string](2, WithSeed(13), WithEFConstruction(32), WithEFSearch(32))
	require.NoError(t, idx.Add("a", 1, 0))
	require.NoError(t, idx.Add("b", 0, 1))

	nodesCap := cap(idx.nodes)

	idx.Clear(true)
	require.Equal(t, 0, idx.Len())
	require.Equal(t, 2, idx.Dim())
	require.Equal(t, nodesCap, cap(idx.nodes))
	require.Equal(t, -1, idx.entry)
	require.Equal(t, 0, idx.maxLevel)

	require.NoError(t, idx.Add("c", 1, 0))
	require.Equal(t, 1, idx.Len())
}

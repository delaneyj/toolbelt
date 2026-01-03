package vecdb

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFlatSaveLoad(t *testing.T) {
	idx := NewFlat[string](2)
	require.NoError(t, idx.Add("a", 0, 0))
	require.NoError(t, idx.Add("b", 1, 0))
	require.NoError(t, idx.Add("c", 0, 1))

	var buf bytes.Buffer
	require.NoError(t, idx.Save(&buf))

	var loaded Flat[string]
	require.NoError(t, loaded.Load(&buf))
	require.Equal(t, idx.Dim(), loaded.Dim())
	require.Equal(t, idx.Metric(), loaded.Metric())

	vec, ok := loaded.Vector("b")
	require.True(t, ok)
	require.Equal(t, []float32{1, 0}, vec)

	results := loaded.Search(1, 0.9, 0)
	require.Len(t, results, 1)
	require.Equal(t, "b", results[0].ID)
}

func TestHNSWSaveLoad(t *testing.T) {
	idx := NewHNSW[string](2, WithSeed(7), WithEFConstruction(32), WithEFSearch(32))
	require.NoError(t, idx.Add("a", 0, 0))
	require.NoError(t, idx.Add("b", 1, 0))
	require.NoError(t, idx.Add("c", 0, 1))
	require.True(t, idx.Delete("b"))

	var buf bytes.Buffer
	require.NoError(t, idx.Save(&buf))

	var loaded HNSW[string]
	require.NoError(t, loaded.Load(&buf))
	require.Equal(t, idx.Dim(), loaded.Dim())
	require.Equal(t, idx.Metric(), loaded.Metric())
	require.Equal(t, idx.Len(), loaded.Len())

	_, ok := loaded.Vector("b")
	require.False(t, ok)

	results := loaded.Search(1, 0.9, 0)
	require.Len(t, results, 1)
	require.Equal(t, "a", results[0].ID)
}

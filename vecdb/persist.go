package vecdb

import (
	"bufio"
	"errors"
	"io"
	"math/rand"
	"reflect"
	"time"

	tb "github.com/delaneyj/toolbelt"
)

const (
	persistVersion uint8 = 1
	persistMagic         = "VECDB"
)

const (
	persistKindFlat uint8 = iota + 1
	persistKindHNSW
)

type persistOptions[ID comparable] struct {
	codec IDCodec[ID]
}

// PersistOption configures Save/Load behavior.
type PersistOption[ID comparable] func(*persistOptions[ID])

// IDCodec handles serialization of ID values.
type IDCodec[ID comparable] interface {
	Encode(w io.Writer, id ID) error
	Decode(r io.Reader) (ID, error)
}

// WithIDCodec overrides ID encoding/decoding for persistence.
func WithIDCodec[ID comparable](codec IDCodec[ID]) PersistOption[ID] {
	return func(opts *persistOptions[ID]) {
		opts.codec = codec
	}
}

func applyPersistOptions[ID comparable](opts []PersistOption[ID]) persistOptions[ID] {
	out := persistOptions[ID]{codec: defaultIDCodec[ID]{}}
	for _, opt := range opts {
		if opt != nil {
			opt(&out)
		}
	}
	return out
}

type defaultIDCodec[ID comparable] struct{}

func (defaultIDCodec[ID]) Encode(w io.Writer, id ID) error {
	v := reflect.ValueOf(id)
	if !v.IsValid() {
		return ErrUnsupportedIDType
	}
	switch v.Kind() {
	case reflect.String:
		return tb.WriteString(w, v.String())
	case reflect.Bool:
		if v.Bool() {
			return tb.WriteUint8(w, 1)
		}
		return tb.WriteUint8(w, 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return tb.WriteInt64(w, v.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return tb.WriteUint64(w, v.Uint())
	case reflect.Float32:
		return tb.WriteFloat32(w, float32(v.Float()))
	default:
		return ErrUnsupportedIDType
	}
}

func (defaultIDCodec[ID]) Decode(r io.Reader) (ID, error) {
	var zero ID
	t := reflect.TypeOf(zero)
	if t == nil {
		return zero, ErrUnsupportedIDType
	}
	v := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.String:
		s, err := tb.ReadString(r)
		if err != nil {
			return zero, err
		}
		v.SetString(s)
	case reflect.Bool:
		b, err := tb.ReadUint8(r)
		if err != nil {
			return zero, err
		}
		v.SetBool(b != 0)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := tb.ReadInt64(r)
		if err != nil {
			return zero, err
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		n, err := tb.ReadUint64(r)
		if err != nil {
			return zero, err
		}
		v.SetUint(n)
	case reflect.Float32:
		f, err := tb.ReadFloat32(r)
		if err != nil {
			return zero, err
		}
		v.SetFloat(float64(f))
	default:
		return zero, ErrUnsupportedIDType
	}
	return v.Interface().(ID), nil
}

// Save writes the flat index to w.
func (f *Flat[ID]) Save(w io.Writer, opts ...PersistOption[ID]) error {
	if w == nil {
		return errors.New("vecdb: nil writer")
	}
	cfg := applyPersistOptions(opts)
	bw := bufio.NewWriter(w)
	f.mu.RLock()
	defer f.mu.RUnlock()
	if err := writeHeader(bw, persistKindFlat); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(f.dim)); err != nil {
		return err
	}
	if err := tb.WriteUint8(bw, uint8(f.metric)); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(len(f.ids))); err != nil {
		return err
	}
	for i, id := range f.ids {
		if err := cfg.codec.Encode(bw, id); err != nil {
			return err
		}
		if len(f.vectors[i]) != f.dim {
			return ErrInvalidFormat
		}
		if err := writeFloat32Slice(bw, f.vectors[i]); err != nil {
			return err
		}
	}
	return bw.Flush()
}

// Load replaces the flat index with data read from r.
func (f *Flat[ID]) Load(r io.Reader, opts ...PersistOption[ID]) error {
	if r == nil {
		return errors.New("vecdb: nil reader")
	}
	cfg := applyPersistOptions(opts)
	br := bufio.NewReader(r)
	if err := readHeader(br, persistKindFlat); err != nil {
		return err
	}
	dim32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	dim, err := checkedInt(dim32)
	if err != nil {
		return err
	}
	metricByte, err := tb.ReadUint8(br)
	if err != nil {
		return err
	}
	metric := Metric(metricByte)
	if metric != MetricL2Squared && metric != MetricCosine {
		return ErrInvalidFormat
	}
	count32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	count, err := checkedInt(count32)
	if err != nil {
		return err
	}
	ids := make([]ID, 0, count)
	vectors := make([][]float32, 0, count)
	index := make(map[ID]int, count)
	for i := 0; i < count; i++ {
		id, err := cfg.codec.Decode(br)
		if err != nil {
			return err
		}
		vector := make([]float32, dim)
		if err := readFloat32Slice(br, vector); err != nil {
			return err
		}
		ids = append(ids, id)
		vectors = append(vectors, vector)
		index[id] = i
	}
	f.mu.Lock()
	defer f.mu.Unlock()
	f.dim = dim
	f.metric = metric
	f.ids = ids
	f.vectors = vectors
	f.index = index
	return nil
}

// Save writes the HNSW index to w, preserving graph structure.
func (h *HNSW[ID]) Save(w io.Writer, opts ...PersistOption[ID]) error {
	if w == nil {
		return errors.New("vecdb: nil writer")
	}
	cfg := applyPersistOptions(opts)
	bw := bufio.NewWriter(w)
	h.mu.RLock()
	defer h.mu.RUnlock()
	if err := writeHeader(bw, persistKindHNSW); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(h.dim)); err != nil {
		return err
	}
	if err := tb.WriteUint8(bw, uint8(h.metric)); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(h.m)); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(h.efConstruction)); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(h.efSearch)); err != nil {
		return err
	}
	if err := tb.WriteInt32(bw, int32(h.entry)); err != nil {
		return err
	}
	if err := tb.WriteInt32(bw, int32(h.maxLevel)); err != nil {
		return err
	}
	if err := tb.WriteUint32(bw, uint32(len(h.nodes))); err != nil {
		return err
	}
	for _, node := range h.nodes {
		if err := cfg.codec.Encode(bw, node.id); err != nil {
			return err
		}
		if node.deleted {
			if err := tb.WriteUint8(bw, 1); err != nil {
				return err
			}
		} else {
			if err := tb.WriteUint8(bw, 0); err != nil {
				return err
			}
		}
		if err := tb.WriteInt32(bw, int32(node.level)); err != nil {
			return err
		}
		if err := tb.WriteFloat32(bw, node.norm); err != nil {
			return err
		}
		if len(node.vector) != h.dim {
			return ErrInvalidFormat
		}
		if err := writeFloat32Slice(bw, node.vector); err != nil {
			return err
		}
		if len(node.links) < node.level+1 {
			return ErrInvalidFormat
		}
		for level := 0; level <= node.level; level++ {
			links := node.links[level]
			if err := tb.WriteUint32(bw, uint32(len(links))); err != nil {
				return err
			}
			for _, nb := range links {
				if err := tb.WriteInt32(bw, int32(nb)); err != nil {
					return err
				}
			}
		}
	}
	return bw.Flush()
}

// Load replaces the HNSW index with data read from r.
func (h *HNSW[ID]) Load(r io.Reader, opts ...PersistOption[ID]) error {
	if r == nil {
		return errors.New("vecdb: nil reader")
	}
	cfg := applyPersistOptions(opts)
	br := bufio.NewReader(r)
	if err := readHeader(br, persistKindHNSW); err != nil {
		return err
	}
	dim32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	dim, err := checkedInt(dim32)
	if err != nil {
		return err
	}
	metricByte, err := tb.ReadUint8(br)
	if err != nil {
		return err
	}
	metric := Metric(metricByte)
	if metric != MetricL2Squared && metric != MetricCosine {
		return ErrInvalidFormat
	}
	m32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	efConstruction32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	efSearch32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	entry32, err := tb.ReadInt32(br)
	if err != nil {
		return err
	}
	maxLevel32, err := tb.ReadInt32(br)
	if err != nil {
		return err
	}
	nodeCount32, err := tb.ReadUint32(br)
	if err != nil {
		return err
	}
	nodeCount, err := checkedInt(nodeCount32)
	if err != nil {
		return err
	}
	nodes := make([]hnswNode[ID], nodeCount)
	index := make(map[ID]int, nodeCount)
	for i := 0; i < nodeCount; i++ {
		id, err := cfg.codec.Decode(br)
		if err != nil {
			return err
		}
		deletedByte, err := tb.ReadUint8(br)
		if err != nil {
			return err
		}
		level32, err := tb.ReadInt32(br)
		if err != nil {
			return err
		}
		level := int(level32)
		if level < 0 {
			return ErrInvalidFormat
		}
		norm, err := tb.ReadFloat32(br)
		if err != nil {
			return err
		}
		vector := make([]float32, dim)
		if err := readFloat32Slice(br, vector); err != nil {
			return err
		}
		links := make([][]int, level+1)
		for l := 0; l <= level; l++ {
			nn32, err := tb.ReadUint32(br)
			if err != nil {
				return err
			}
			nn, err := checkedInt(nn32)
			if err != nil {
				return err
			}
			if nn == 0 {
				continue
			}
			neighbors := make([]int, nn)
			for j := 0; j < nn; j++ {
				nb32, err := tb.ReadInt32(br)
				if err != nil {
					return err
				}
				if nb32 < 0 || nb32 >= int32(nodeCount) {
					return ErrInvalidFormat
				}
				neighbors[j] = int(nb32)
			}
			links[l] = neighbors
		}
		nodes[i] = hnswNode[ID]{
			id:      id,
			vector:  vector,
			norm:    norm,
			level:   level,
			links:   links,
			deleted: deletedByte != 0,
		}
		if deletedByte == 0 {
			index[id] = i
		}
	}
	entry := int(entry32)
	if entry < -1 || entry >= nodeCount {
		return ErrInvalidFormat
	}
	maxLevel := int(maxLevel32)
	if maxLevel < 0 {
		return ErrInvalidFormat
	}
	if nodeCount == 0 {
		entry = -1
		maxLevel = 0
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dim = dim
	h.metric = metric
	h.m = int(m32)
	h.efConstruction = int(efConstruction32)
	h.efSearch = int(efSearch32)
	h.entry = entry
	h.maxLevel = maxLevel
	h.nodes = nodes
	h.index = index
	if h.rng == nil {
		h.rng = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	h.candidatePool = newCandidatePool(h.efConstruction)
	h.visitedPool = newVisitedPool(h.efConstruction)
	return nil
}

func writeHeader(w io.Writer, kind uint8) error {
	if _, err := w.Write([]byte(persistMagic)); err != nil {
		return err
	}
	if err := tb.WriteUint8(w, persistVersion); err != nil {
		return err
	}
	return tb.WriteUint8(w, kind)
}

func readHeader(r io.Reader, expectedKind uint8) error {
	var magic [len(persistMagic)]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return err
	}
	if string(magic[:]) != persistMagic {
		return ErrInvalidFormat
	}
	version, err := tb.ReadUint8(r)
	if err != nil {
		return err
	}
	if version != persistVersion {
		return ErrUnsupportedVersion
	}
	kind, err := tb.ReadUint8(r)
	if err != nil {
		return err
	}
	if kind != expectedKind {
		return ErrInvalidFormat
	}
	return nil
}

func checkedInt(v uint32) (int, error) {
	maxInt := int(^uint(0) >> 1)
	if v > uint32(maxInt) {
		return 0, ErrInvalidFormat
	}
	return int(v), nil
}

func writeFloat32Slice(w io.Writer, v []float32) error {
	for _, f := range v {
		if err := tb.WriteFloat32(w, f); err != nil {
			return err
		}
	}
	return nil
}

func readFloat32Slice(r io.Reader, v []float32) error {
	for i := range v {
		f, err := tb.ReadFloat32(r)
		if err != nil {
			return err
		}
		v[i] = f
	}
	return nil
}

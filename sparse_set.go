package toolbelt

// Handle is a stable reference to an element in a SparseSet.
// Index and Generation are both uint32. A handle is valid if its
// generation matches the current generation for its slot in the set
// in which it was created.
type Handle struct {
    Index      uint32
    Generation uint32
}

// SparseSet is a growable, O(1) add/get/remove container that uses
// stable, generational handles to reference elements without exposing
// internal indices. Removals are constant time by swapping the last
// element into the removed slot in the dense array.
//
// This is a Go port inspired by Karl Zylinski's odin-handle-map
// (handle_map_growing), adapted to Go generics and idioms.
//
// The zero value is ready to use; use NewSparseSet to pre-reserve.
type SparseSet[T any] struct {
    // dense holds the contiguous elements for iteration/cache friendliness.
    dense []T
    // denseSlots maps dense index -> slot id (sparse index) for reverse updates on swap-remove.
    denseSlots []uint32

    // sparse maps slot id -> dense index; -1 indicates free/unoccupied slot.
    sparse []int
    // generations is per-slot generation counter used to invalidate stale handles.
    generations []uint32

    // free is a stack of reusable slot ids.
    free []uint32
}

// NewSparseSet constructs a SparseSet with an optional initial slot capacity.
// A capacity of 0 creates an empty set with no pre-allocated slots.
func NewSparseSet[T any](initialCapacity int) *SparseSet[T] {
    s := &SparseSet[T]{}
    if initialCapacity > 0 {
        s.reserveSlots(initialCapacity)
    }
    return s
}

// Len returns the number of live elements.
func (s *SparseSet[T]) Len() int { return len(s.dense) }

// Cap returns the number of total slots (live + free) that can be used
// without growing the sparse structures.
func (s *SparseSet[T]) Cap() int { return len(s.sparse) }

// Reserve ensures the set can accommodate at least n live elements
// without reallocating dense storage. Note: this does not change the
// number of available slots for handles. To increase the number of
// potential handles without insertions, call ReserveSlots.
func (s *SparseSet[T]) Reserve(n int) {
    if n > cap(s.dense) {
        // Grow dense slices' capacity
        newDense := make([]T, len(s.dense), n)
        copy(newDense, s.dense)
        s.dense = newDense

        newDenseSlots := make([]uint32, len(s.denseSlots), n)
        copy(newDenseSlots, s.denseSlots)
        s.denseSlots = newDenseSlots
    }
}

// maxSlotCount returns the maximum number of addressable slots for a uint32 index,
// clamped to the platform's max int to avoid overflow on 32-bit architectures.
func maxSlotCount() int {
    // If MaxInt < MaxUint32, clamp to MaxInt; otherwise return 2^32.
    if (uint64(^uint(0)) >> 1) < uint64(^uint32(0)) {
        return int(^uint(0) >> 1)
    }
    return int(^uint32(0)) + 1
}

// ReserveSlots ensures the set has at least n slots available to be
// populated by future Insert calls before growing the sparse structures.
func (s *SparseSet[T]) ReserveSlots(n int) { s.reserveSlots(n) }

func (s *SparseSet[T]) reserveSlots(n int) {
    if n <= len(s.sparse) {
        return
    }
    if n > maxSlotCount() {
        panic("requested slots exceed index type capacity")
    }
    // Extend sparse arrays; new slots start as unoccupied with generation 1.
    add := n - len(s.sparse)
    s.sparse = append(s.sparse, make([]int, add)...)
    s.generations = append(s.generations, make([]uint32, add)...)
    // Mark new slots free and initialize generation to 1 so zero is always invalid.
    for i := len(s.sparse) - add; i < len(s.sparse); i++ {
        s.sparse[i] = -1
        s.generations[i] = uint32(1)
        s.free = append(s.free, uint32(i))
    }
}

// Insert adds v to the set and returns a stable handle to it.
// Amortized O(1). May grow the internal storage.
func (s *SparseSet[T]) Insert(v T) Handle {
    var slot uint32
    if n := len(s.free); n > 0 {
        slot = s.free[n-1]
        s.free = s.free[:n-1]
    } else {
        // Grow by at least one slot when needed.
        i := len(s.sparse)
        if i >= maxSlotCount() {
            panic("exceeded index type capacity")
        }
        s.sparse = append(s.sparse, -1)
        s.generations = append(s.generations, uint32(1)) // start at 1; 0 means invalid
        slot = uint32(i)
    }

    denseIndex := len(s.dense)
    s.dense = append(s.dense, v)
    s.denseSlots = append(s.denseSlots, slot)
    s.sparse[int(slot)] = denseIndex

    return Handle{Index: slot, Generation: s.generations[int(slot)]}
}

// Get returns the value for h if it is valid and present.
func (s *SparseSet[T]) Get(h Handle) (T, bool) {
    var zero T
    idx, ok := s.indexOf(h)
    if !ok {
        return zero, false
    }
    return s.dense[idx], true
}

// GetRef returns a pointer to the value for h if valid.
// Note: Removing other elements may move the value due to swap-remove,
// invalidating previously taken pointers. Use only transiently.
func (s *SparseSet[T]) GetRef(h Handle) (*T, bool) {
    idx, ok := s.indexOf(h)
    if !ok {
        return nil, false
    }
    return &s.dense[idx], true
}

// Contains reports whether h refers to a live element in s.
func (s *SparseSet[T]) Contains(h Handle) bool {
    _, ok := s.indexOf(h)
    return ok
}

// Remove deletes the element referenced by h if present.
// Returns true if the element existed and was removed.
func (s *SparseSet[T]) Remove(h Handle) bool {
    idx, ok := s.indexOf(h)
    if !ok {
        return false
    }
    // Slot corresponding to the element to remove
    slot := s.denseSlots[idx]

    last := len(s.dense) - 1
    if idx != last {
        // Move last element into removed position
        s.dense[idx] = s.dense[last]
        s.denseSlots[idx] = s.denseSlots[last]
        movedSlot := s.denseSlots[idx]
        s.sparse[int(movedSlot)] = idx
    }
    // Truncate dense arrays
    s.dense = s.dense[:last]
    s.denseSlots = s.denseSlots[:last]

    // Invalidate handle and free slot
    s.sparse[int(slot)] = -1
    s.generations[int(slot)]++
    if s.generations[int(slot)] == 0 { // avoid zero being a valid generation
        s.generations[int(slot)] = uint32(1)
    }
    s.free = append(s.free, slot)
    return true
}

// Clear removes all elements. Existing handles become invalid.
func (s *SparseSet[T]) Clear() {
    // Invalidate all occupied slots and add them to free list.
    for i := range s.denseSlots {
        slot := s.denseSlots[i]
        s.sparse[int(slot)] = -1
        s.generations[int(slot)]++
        if s.generations[int(slot)] == 0 {
            s.generations[int(slot)] = uint32(1)
        }
        s.free = append(s.free, slot)
    }
    s.dense = s.dense[:0]
    s.denseSlots = s.denseSlots[:0]
}

// Range iterates over all live elements. The callback receives a current
// handle for each element and a pointer to its value. If f returns false,
// iteration stops early.
func (s *SparseSet[T]) Range(f func(h Handle, v *T) bool) {
    for i := 0; i < len(s.dense); i++ {
        slot := s.denseSlots[i]
        h := Handle{Index: slot, Generation: s.generations[int(slot)]}
        if !f(h, &s.dense[i]) {
            return
        }
    }
}

// Handles returns a snapshot of handles for all live elements.
func (s *SparseSet[T]) Handles() []Handle {
    out := make([]Handle, len(s.dense))
    for i, slot := range s.denseSlots {
        out[i] = Handle{Index: slot, Generation: s.generations[int(slot)]}
    }
    return out
}

// Values returns a shallow copy of all live elements in dense order.
func (s *SparseSet[T]) Values() []T {
    out := make([]T, len(s.dense))
    copy(out, s.dense)
    return out
}

// indexOf returns the dense index of handle h if valid.
func (s *SparseSet[T]) indexOf(h Handle) (int, bool) {
    idxI := int(h.Index)
    if idxI < 0 || idxI >= len(s.sparse) {
        return -1, false
    }
    di := s.sparse[idxI]
    if di < 0 || di >= len(s.dense) {
        return -1, false
    }
    if s.generations[idxI] != h.Generation {
        return -1, false
    }
    // Sanity: verify reverse map points back to same slot (should always hold)
    if s.denseSlots[di] != h.Index {
        return -1, false
    }
    return di, true
}

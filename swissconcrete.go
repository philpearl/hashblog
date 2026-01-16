package hashblog

import (
	"math/bits"
	"unsafe"
)

// SwissTable is a hash table implementation inspired by the SwissTable design
// used in the Go standard library.
//
// It uses open addressing with quadratic probing, grouping entries into groups
// of 8. Each group has a control byte per entry that indicates whether the slot
// is empty, and stores part of the hash of the key to speed up lookups.
//
// SwissTable improves on GroupTableCtrl by using bitwise operations to find
// matching control bytes and empty slots, reducing the number of operations
// needed to find these.
type SwissConcrete struct {
	groups [groupTableSize]concreteGroupWithCtrl
}

func NewSwissConcrete() *SwissConcrete {
	m := &SwissConcrete{}
	for i := range m.groups {
		m.groups[i].ctrl = concreteCtrl(0x8080_8080_8080_8080)
	}
	return m
}

func (m *SwissConcrete) Set(key string, value int) {
	h := concreteHash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	// Expand h1 to a 64-bit value where each byte is h1. This allows us to
	// compare against all control bytes in a group simultaneously.
	h1Expanded := uint64(h1) * 0x0101010101010101

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Find possible matches for this entry in the group. findMatches
		// returns a bitmask where each byte with a matching control byte has
		// its high bit set.
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := bits.TrailingZeros64(matches) / 8
			if e := &g.entries[i]; e.key == key {
				e.value = value
				return
			}
			// Clear the lowest set bit and continue
			matches &= matches - 1
		}
		// Check for empty slot in group. This returns a bitmask where each
		// byte that is empty has its high bit set.
		if empties := g.ctrl.findEmpty(); empties != 0 {
			i := bits.TrailingZeros64(empties) / 8
			// Empty slot - this means the key is not present in the table
			g.entries[i] = concreteEntry{key: key, value: value}
			g.ctrl.set(i, h1)
			// g.ctrl[i] = h1
			return
		}
	}
}

func (m *SwissConcrete) Get(key string) (v int, ok bool) {
	if m == nil {
		return v, false
	}
	h := concreteHash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	// Expand h1 to a 64-bit value where each byte is h1. This allows us to
	// compare against all control bytes in a group simultaneously.
	h1Expanded := uint64(h1) * 0x0101_0101_0101_0101

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Find possible matches for this entry in the group. findMatches
		// returns a bitmask where each byte with a matching control byte has
		// its high bit set.
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := bits.TrailingZeros64(matches) / 8
			if e := &g.entries[i]; e.key == key {
				return e.value, true
			}
			// Clear the lowest set bit and continue
			matches &= matches - 1
		}
		// Check for empty slot in the group. If there is an empty slot, the key
		// is not present in the table.
		empties := g.ctrl.findEmpty()
		if empties != 0 {
			return v, false
		}
	}
}

type concreteGroupWithCtrl struct {
	ctrl    concreteCtrl
	entries [groupSize]concreteEntry
}

type concreteCtrl uint64

func (gc concreteCtrl) findMatches(h1Expanded uint64) uint64 {
	matchesAreZero := (uint64(gc) ^ h1Expanded)
	return ((matchesAreZero - 0x0101_0101_0101_0101) &^ matchesAreZero) & 0x8080_8080_8080_8080
}

func (gc concreteCtrl) findEmpty() uint64 {
	return (uint64(gc) & 0x8080_8080_8080_8080)
}

func (gc *concreteCtrl) set(i int, v byte) {
	(*(*[8]byte)(unsafe.Pointer(gc)))[i] = v
}

type concreteEntry struct {
	key   string
	value int
}

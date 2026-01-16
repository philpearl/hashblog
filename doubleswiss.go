//go:build goexperiment.simd && amd64

package hashblog

import (
	"math/bits"
	"simd/archsimd"
)

const (
	doubleSwissGroupSize = 16
	doubleSwissTableSize = simpleTableSize / doubleSwissGroupSize
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
type DoubleSwiss struct {
	groups [doubleSwissTableSize]swissGroup
}

func NewDoubleSwiss() *DoubleSwiss {
	m := &DoubleSwiss{}
	for i := range m.groups {
		m.groups[i].ctrl = swissCtrl{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	}
	return m
}

func (m *DoubleSwiss) Set(key string, value int) {
	h := concreteHash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	// Expand h1 to a 64-bit value where each byte is h1. This allows us to
	// compare against all control bytes in a group simultaneously.
	h1Expanded := archsimd.BroadcastUint8x16(h1)

	for seq := makeProbeSeq(h2, hashValue(doubleSwissTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Find possible matches for this entry in the group. findMatches
		// returns a bitmask where each byte with a matching control byte has
		// its high bit set.
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := matches.first()
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
			i := empties.first()
			// Empty slot - this means the key is not present in the table
			g.entries[i] = concreteEntry{key: key, value: value}
			g.ctrl[i] = h1
			return
		}
	}
}

func (m *DoubleSwiss) Get(key string) (v int, ok bool) {
	if m == nil {
		return v, false
	}
	h := concreteHash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	// Expand h1 to a 64-bit value where each byte is h1. This allows us to
	// compare against all control bytes in a group simultaneously.
	h1Expanded := archsimd.BroadcastUint8x16(h1)

	for seq := makeProbeSeq(h2, hashValue(doubleSwissTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Find possible matches for this entry in the group. findMatches
		// returns a bitmask where each byte with a matching control byte has
		// its high bit set.
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := matches.first()
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

type swissGroup struct {
	ctrl    swissCtrl
	entries [doubleSwissGroupSize]concreteEntry
}

type swissCtrl [doubleSwissGroupSize]uint8

type matchType uint16

func (m matchType) first() int {
	return bits.TrailingZeros16(uint16(m))
}

func (gc *swissCtrl) findMatches(ctrlHash archsimd.Uint8x16) matchType {
	return matchType(archsimd.LoadUint8x16((*[doubleSwissGroupSize]uint8)(gc)).Equal(ctrlHash).ToBits())
}

var emptyMask = archsimd.BroadcastUint8x16(0x80)

func (gc *swissCtrl) findEmpty() matchType {
	return matchType(archsimd.LoadUint8x16((*[doubleSwissGroupSize]uint8)(gc)).Equal(emptyMask).ToBits())
}

package hashblog

import (
	"math/bits"
	"unsafe"
)

type SwissTable[K comparable, V any] struct {
	groups [groupTableSize]groupWithCtrl[K, V]
}

func NewSwissTable[K comparable, V any]() *SwissTable[K, V] {
	m := &SwissTable[K, V]{}
	for i := range m.groups {
		m.groups[i].ctrl = groupCtrl{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	}
	return m
}

func (m *SwissTable[K, V]) Set(key K, value V) {
	h := hash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	h1Expanded := uint64(h1) * 0x0101010101010101

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Is the key in this group?
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := bits.TrailingZeros64(matches) / 8
			if e := &g.entries[i]; e.key == key {
				e.value = value
				return
			}
			matches &= matches - 1
		}
		// Check for empty slot in group
		if empties := g.ctrl.findEmpty(); empties != 0 {
			i := bits.TrailingZeros64(empties) / 8
			// Empty slot - this means the key is not present in the table
			g.entries[i] = entry[K, V]{key: key, value: value}
			g.ctrl[i] = h1
			return
		}
	}
}

func (m *SwissTable[K, V]) Get(key K) (v V, ok bool) {
	h := hash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	h1Expanded := uint64(h1) * 0x0101010101010101

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Is the key in this group?
		matches := g.ctrl.findMatches(h1Expanded)
		for matches != 0 {
			i := bits.TrailingZeros64(matches) / 8
			matches &= matches - 1
			e := &g.entries[i]
			if e.key == key {
				return e.value, true
			}
		}
		// Check for empty slot in group
		empties := g.ctrl.findEmpty()
		if empties != 0 {
			return v, false
		}
	}
}

func (gc groupCtrl) findMatches(h1Expanded uint64) uint64 {
	// Find the entries where the control byte matches ctrlHash
	//
	// We expand the ctrlHash to a groupControl where each byte is ctrlHash,
	// then XOR that with the group control. Any byte that was equal will now be
	// zero. We then subtract 0x01 from each byte, so any byte that was zero
	// will now have its high bit set. Finally we AND with 0x80 to keep only the
	// high bits.
	//
	// Note this does give false positives!
	matchesAreZero := (gc.toBitmask() ^ h1Expanded)
	return ((matchesAreZero - 0x0101010101010101) &^ matchesAreZero) & 0x8080808080808080
}

func (gc groupCtrl) findEmpty() uint64 {
	return gc.toBitmask() & 0x8080808080808080
}

func (gc groupCtrl) toBitmask() uint64 {
	return *(*uint64)(unsafe.Pointer(&gc))
}

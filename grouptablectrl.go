package hashblog

// GroupTableCtrl is a hash table implementation using group-based storage with
// control bytes. The control bytes should improve lookup speed by reducing the
// number of key comparisons needed during probing. But they also add
// complexity.
//
// This is a stepping stone to a full swiss table implementation.
//
// Note for this implementation I've removed the "find" function and moved that
// code into the Set and Get methods. This is because we need to update the
// control bytes as well as the entries.
type GroupTableCtrl[K comparable, V any] struct {
	groups [groupTableSize]groupWithCtrl[K, V]
}

func NewGroupTableCtrl[K comparable, V any]() *GroupTableCtrl[K, V] {
	m := &GroupTableCtrl[K, V]{}
	for i := range m.groups {
		// There's a control byte per entry in the group. The top bit indicates
		// whether the slot is empty. It's set to 1 when empty. The rest of the
		// byte is used to store the bottom 7 bits of the hash of the key.
		m.groups[i].ctrl = groupCtrl{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	}
	return m
}

func (m *GroupTableCtrl[K, V]) Set(key K, value V) {
	h := hash(key)

	// h1 is the control byte value (bottom 7 bits of hash, top bit clear to
	// indicate the slot in the group is in use).
	//
	// h2 is used for probing at the group level. We don't use the bottom 7 bits
	// of the hash so that entries with the same h1 are more likely to be in
	// different groups.
	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Is the key in this group?
		for i := range g.ctrl {
			if g.ctrl[i] != h1 {
				continue
			}
			if e := &g.entries[i]; e.key == key {
				e.value = value
				return
			}
		}
		// Check for empty slot in group
		for i := range g.ctrl {
			if g.ctrl[i] != 0x80 {
				continue
			}
			// Empty slot - this means the key is not present in the table
			g.ctrl[i] = h1
			g.entries[i] = entry[K, V]{key: key, value: value}
			return
		}
	}
}

func (m *GroupTableCtrl[K, V]) Get(key K) (v V, ok bool) {
	h := hash(key)

	h1 := byte(h & 0x7F)
	h2 := (h >> 7)

	for seq := makeProbeSeq(h2, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Is the key in this group?
		for i := range g.ctrl {
			ctrl := g.ctrl[i]
			if ctrl == 0x80 {
				return v, false
			}
			if g.ctrl[i] == h1 {
				if e := &g.entries[i]; e.key == key {
					return e.value, true
				}
			}
		}
	}
}

type groupWithCtrl[K comparable, V any] struct {
	ctrl    groupCtrl
	entries [groupSize]entry[K, V]
}

type groupCtrl [groupSize]byte

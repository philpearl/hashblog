package hashblog

const groupTableSize = simpleTableSize / groupSize

// GroupTable is a hash table implementation using grouping without control
// bytes. Each group contains multiple entries, and probing is done at the group
// level.
//
// This potentially improves cache locality compared to probing each entry
// individually, but more likely is worse than a simpler implementation, and is
// only a stepping stone to a full swiss table.
type GroupTable[K comparable, V any] struct {
	groups [groupTableSize]group[K, V]
}

func NewGroupTable[K comparable, V any]() *GroupTable[K, V] {
	return &GroupTable[K, V]{}
}

func (m *GroupTable[K, V]) Set(key K, value V) {
	ent, _ := m.find(key)
	*ent = entry[K, V]{key: key, value: value}
}

func (m *GroupTable[K, V]) Get(key K) (v V, ok bool) {
	ent, ok := m.find(key)
	if ok {
		return ent.value, true
	}
	return v, false
}

func (m *GroupTable[K, V]) find(key K) (*entry[K, V], bool) {
	h := hash(key)

	var zero K
	for seq := makeProbeSeq(h, hashValue(groupTableSize-1)); ; seq = seq.next() {
		g := &m.groups[seq.offset]
		// Is the key in this group?
		for i := range g {
			e := &g[i]
			if e.key == key {
				return e, true
			}
			if e.key == zero {
				// Empty slot - this means the key is not present in the table
				return e, false
			}
		}
	}
}

const groupSize = 8

type group[K comparable, V any] [groupSize]entry[K, V]

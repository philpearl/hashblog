package hashblog

const groupTableSize = simpleTableSize / groupSize

type GroupTable[K comparable, V any] struct {
	groups [groupTableSize]Group[K, V]
}

func NewGroupTable[K comparable, V any]() *GroupTable[K, V] {
	return &GroupTable[K, V]{}
}

func (m *GroupTable[K, V]) Set(key K, value V) {
	ent, _ := m.find(key)
	*ent = entry[K, V]{key: key, value: value}
}

func (m *GroupTable[K, V]) Get(key K) (V, bool) {
	ent, ok := m.find(key)
	if ok {
		return ent.value, true
	}
	var zero V
	return zero, false
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
		}
		// Check for empty slot in group
		for i := range g {
			e := &g[i]
			if e.key == zero {
				// Empty slot - this means the key is not present in the table
				return e, false
			}
		}
	}
}

const groupSize = 8

type Group[K comparable, V any] [groupSize]entry[K, V]

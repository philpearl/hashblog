package hashblog

type GroupTableCtrl[K comparable, V any] struct {
	groups [groupTableSize]groupWithCtrl[K, V]
}

func NewGroupTableCtrl[K comparable, V any]() *GroupTableCtrl[K, V] {
	m := &GroupTableCtrl[K, V]{}
	for i := range m.groups {
		m.groups[i].ctrl = groupCtrl{0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80}
	}
	return m
}

func (m *GroupTableCtrl[K, V]) Set(key K, value V) {
	h := hash(key)

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

type groupLocation[K comparable, V any] struct {
	groups *[groupTableSize]groupWithCtrl[K, V]
	index  hashValue
	slot   int
	h1     byte
}

func (g groupLocation[K, V]) set(k K, v V) {
	g.groups[g.index].ctrl[g.slot] = g.h1
	g.groups[g.index].entries[g.slot] = entry[K, V]{key: k, value: v}
}

func (g groupLocation[K, V]) get() V {
	return g.groups[g.index].entries[g.slot].value
}

type groupWithCtrl[K comparable, V any] struct {
	ctrl    groupCtrl
	entries [groupSize]entry[K, V]
}

type groupCtrl [groupSize]byte

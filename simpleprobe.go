package hashblog

type SimpleTableProbe[K comparable, V any] struct {
	entries [simpleTableSize]entry[K, V]
}

func NewSimpleTableProbe[K comparable, V any]() *SimpleTableProbe[K, V] {
	return &SimpleTableProbe[K, V]{}
}

func (st *SimpleTableProbe[K, V]) Set(key K, value V) {
	ent, _ := st.find(key)
	*ent = entry[K, V]{key: key, value: value}
}

func (st *SimpleTableProbe[K, V]) Get(key K) (V, bool) {
	ent, ok := st.find(key)
	if ok {
		return ent.value, true
	}
	var zero V
	return zero, false
}

func (st *SimpleTableProbe[K, V]) find(key K) (*entry[K, V], bool) {
	h := hash(key)

	var zero K
	for seq := makeProbeSeq(h, hashValue(simpleTableSize-1)); ; seq = seq.next() {
		e := &st.entries[seq.offset]
		if e.key == key {
			return e, true
		}
		if e.key == zero {
			// Empty slot - this means the key is not present in the table
			return e, false
		}
	}
}

type probeSeq struct {
	mask   hashValue
	offset hashValue
	index  hashValue
}

func makeProbeSeq(hash, mask hashValue) probeSeq {
	return probeSeq{
		mask:   mask,
		offset: hash & mask,
		index:  0,
	}
}

func (s probeSeq) next() probeSeq {
	s.index++
	s.offset = (s.offset + s.index) & s.mask
	return s
}

package hashblog

const simpleTableSize = 4096 * 8

type SimpleTable[K comparable, V any] struct {
	entries [simpleTableSize]entry[K, V]
}

func NewSimpleTable[K comparable, V any]() *SimpleTable[K, V] {
	return &SimpleTable[K, V]{}
}

func (st *SimpleTable[K, V]) Set(key K, value V) {
	ent, _ := st.find(key)
	*ent = entry[K, V]{key: key, value: value}
}

func (st *SimpleTable[K, V]) Get(key K) (V, bool) {
	ent, ok := st.find(key)
	if ok {
		return ent.value, true
	}
	var zero V
	return zero, false
}

func (st *SimpleTable[K, V]) find(key K) (*entry[K, V], bool) {
	index := hash(key) % simpleTableSize

	var zero K
	for {
		e := &st.entries[index]
		if e.key == key {
			return e, true
		}
		if e.key == zero {
			// Empty slot - this means the key is not present in the table
			return e, false
		}
		index = (index + 1) % simpleTableSize
	}
}

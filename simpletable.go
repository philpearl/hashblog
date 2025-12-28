package hashblog

// We're creating a fixed-size table. Our table length is a power of two to
// speed up modulo arthimetic.
const simpleTableSize = 32768

// SimpleTable is a basic hash table implementation using open addressing with
// linear probing.
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

func (st *SimpleTable[K, V]) Get(key K) (v V, ok bool) {
	ent, ok := st.find(key)
	if ok {
		return ent.value, true
	}
	return v, false
}

func (st *SimpleTable[K, V]) find(key K) (*entry[K, V], bool) {
	index := hash(key) % simpleTableSize

	var zero K
	for {
		e := &st.entries[index]
		if e.key == key {
			// Found our entry
			return e, true
		}
		if e.key == zero {
			// Empty slot - this means the key is not present in the table
			return e, false
		}
		index = (index + 1) % simpleTableSize
	}
}

package hashblog

import (
	"hash/maphash"
	"unsafe"
)

var seed = maphash.MakeSeed()

type hashValue uint64

func hash[K comparable](key K) hashValue {
	return hashValue(maphash.Comparable(seed, key))
}

func concreteHash(key string) hashValue {
	// Going direct to runtime_memhash seems to save a nanosecond. 4.7ns vs 3.7ns
	// return hashValue(maphash.String(seed, key))
	// return hashValue(maphash.Comparable(seed, key))
	return hashValue(runtime_memhash(
		unsafe.Pointer(unsafe.StringData(key)),
		0,
		uintptr(len(key)),
	))
}

// We use the runtime's map hash function without the overhead of using
// hash/maphash
//
//go:linkname runtime_memhash runtime.memhash
//go:noescape
func runtime_memhash(p unsafe.Pointer, seed, s uintptr) uintptr

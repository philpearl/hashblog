package hashblog

import "hash/maphash"

var seed = maphash.MakeSeed()

type hashValue uint64

func hash[K comparable](key K) hashValue {
	return hashValue(maphash.Comparable(seed, key))
}

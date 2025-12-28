package hashblog

type entry[K comparable, V any] struct {
	key   K
	value V
}

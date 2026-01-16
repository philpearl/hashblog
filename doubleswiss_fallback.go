//go:build !goexperiment.simd || !amd64

package hashblog

// Just so tests will compile and run when we can't do SIMD
func NewDoubleSwiss() *SwissConcrete {
	return NewSwissConcrete()
}

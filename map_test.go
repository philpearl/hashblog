package hashblog_test

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/philpearl/hashblog"
)

type mapper interface {
	Set(key string, value int)
	Get(key string) (int, bool)
}

func TestGetMissing(t *testing.T) {
	for _, m := range []mapper{
		hashblog.NewSimpleTable[string, int](),
		hashblog.NewSimpleTableProbe[string, int](),
		hashblog.NewGroupTable[string, int](),
		hashblog.NewGroupTableCtrl[string, int](),
		hashblog.NewSwissTable[string, int](),
	} {
		t.Run(fmt.Sprintf("%T", m), func(t *testing.T) {
			if _, ok := m.Get("missing"); ok {
				t.Fatalf("expected missing key to return ok == false")
			}
			m.Set("present", 42)

			if _, ok := m.Get("missing"); ok {
				t.Fatalf("expected missing key to return ok == false")
			}
		})
	}
}

func TestGetPresent(t *testing.T) {
	for _, m := range []mapper{
		hashblog.NewSimpleTable[string, int](),
		hashblog.NewSimpleTableProbe[string, int](),
		hashblog.NewGroupTable[string, int](),
		hashblog.NewGroupTableCtrl[string, int](),
		hashblog.NewSwissTable[string, int](),
	} {
		t.Run(fmt.Sprintf("%T", m), func(t *testing.T) {
			m.Set("present", 42)

			val, ok := m.Get("present")
			if !ok {
				t.Fatalf("expected present key to return ok == true")
			}

			if val != 42 {
				t.Fatalf("expected value 42, got %d", val)
			}
		})
	}
}

func TestOverwrite(t *testing.T) {
	for _, m := range []mapper{
		hashblog.NewSimpleTable[string, int](),
		hashblog.NewSimpleTableProbe[string, int](),
		hashblog.NewGroupTable[string, int](),
		hashblog.NewGroupTableCtrl[string, int](),
		hashblog.NewSwissTable[string, int](),
	} {
		t.Run(fmt.Sprintf("%T", m), func(t *testing.T) {
			m.Set("key", 1)
			m.Set("key", 2)

			val, ok := m.Get("key")
			if !ok {
				t.Fatalf("expected key to be present")
			}

			if val != 2 {
				t.Fatalf("expected value 2, got %d", val)
			}
		})
	}
}

func BenchmarkSet(b *testing.B) {
	for _, size := range []int{10, 100, 1000, 2000, 4000, 8000, 16000, 24000, 32768} {
		keys := make([]string, size)
		for i := range keys {
			keys[i] = strconv.Itoa(i)
		}

		b.Run(strconv.Itoa(size), func(b *testing.B) {
			b.Run("SimpleTable", func(b *testing.B) {
				m := hashblog.NewSimpleTable[string, int]()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m.Set(key, i)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
			b.Run("SimpleTableProbe", func(b *testing.B) {
				m := hashblog.NewSimpleTableProbe[string, int]()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m.Set(key, i)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
			b.Run("GroupTable", func(b *testing.B) {
				m := hashblog.NewGroupTable[string, int]()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m.Set(key, i)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
			b.Run("GroupTableCtrl", func(b *testing.B) {
				m := hashblog.NewGroupTableCtrl[string, int]()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m.Set(key, i)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
			b.Run("Swiss", func(b *testing.B) {
				m := hashblog.NewSwissTable[string, int]()
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m.Set(key, i)
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
			b.Run("map", func(b *testing.B) {
				m := make(map[string]int, 32768)
				b.ReportAllocs()
				b.ResetTimer()
				for b.Loop() {
					for i, key := range keys {
						m[key] = i
					}
				}
				b.ReportMetric(float64(b.Elapsed().Nanoseconds())/float64(b.N)/float64(size), "ns/op")
			})
		})
	}
}

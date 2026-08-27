package mallocbench

import "testing"

type small80 struct {
	a, b, c, d, e, f, g, h, i, j int64
}

type small64 struct {
	a, b, c, d, e, f, g, h int64
}

var sink any

func BenchmarkHeapAlloc80B(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = new(small80)
	}
}

func BenchmarkHeapAlloc64B(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		sink = new(small64)
	}
}

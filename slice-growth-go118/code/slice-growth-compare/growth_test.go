package main

import "testing"

var sinkLen int
var sinkCap int

func BenchmarkAppendByteGrowth_NoPrealloc_4K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var s []byte
		for j := 0; j < 4096; j++ {
			s = append(s, byte(j))
		}
		sinkLen = len(s)
		sinkCap = cap(s)
	}
}

func BenchmarkAppendByteGrowth_Prealloc256_4K(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]byte, 0, 256)
		for j := 0; j < 4096; j++ {
			s = append(s, byte(j))
		}
		sinkLen = len(s)
		sinkCap = cap(s)
	}
}

func BenchmarkAppendByteGrowth_From1024To4096(b *testing.B) {
	for i := 0; i < b.N; i++ {
		s := make([]byte, 1024, 1024)
		for j := 0; j < 3072; j++ {
			s = append(s, byte(j))
		}
		sinkLen = len(s)
		sinkCap = cap(s)
	}
}

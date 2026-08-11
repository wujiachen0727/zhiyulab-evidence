package main

import (
	"testing"
)

// 全局 sink 防止 DCE
var sink any

const batch = 100

func BenchmarkAlloc32(b *testing.B) {
	b.ReportAllocs()
	var objs []S32
	for i := 0; i < b.N; i++ {
		objs = make([]S32, batch)
		objs[0].a = 1
	}
	sink = objs
}

func BenchmarkAlloc33(b *testing.B) {
	b.ReportAllocs()
	var objs []S33
	for i := 0; i < b.N; i++ {
		objs = make([]S33, batch)
		objs[0].a = 1
	}
	sink = objs
}

func BenchmarkAlloc64(b *testing.B) {
	b.ReportAllocs()
	var objs []S64
	for i := 0; i < b.N; i++ {
		objs = make([]S64, batch)
		objs[0].a = 1
	}
	sink = objs
}

func BenchmarkAlloc65(b *testing.B) {
	b.ReportAllocs()
	var objs []S65
	for i := 0; i < b.N; i++ {
		objs = make([]S65, batch)
		objs[0].a = 1
	}
	sink = objs
}

func BenchmarkAlloc128(b *testing.B) {
	b.ReportAllocs()
	var objs []S128
	for i := 0; i < b.N; i++ {
		objs = make([]S128, batch)
		objs[0].a = 1
	}
	sink = objs
}

func BenchmarkAlloc129(b *testing.B) {
	b.ReportAllocs()
	var objs []S129
	for i := 0; i < b.N; i++ {
		objs = make([]S129, batch)
		objs[0].a = 1
	}
	sink = objs
}

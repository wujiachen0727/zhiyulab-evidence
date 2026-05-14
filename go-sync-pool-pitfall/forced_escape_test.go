package main

import (
	"sync"
	"testing"
)

type Obj16 [16]byte

var pool16 = sync.Pool{New: func() interface{} { return new(Obj16) }}

// sink 防止编译器优化
var sink *Obj16

func BenchmarkPoolForced16(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool16.Get().(*Obj16)
			obj[0] = 1
			pool16.Put(obj)
		}
	})
}

//go:noinline
func allocObj16() *Obj16 {
	return new(Obj16)
}

func BenchmarkAllocForced16(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := allocObj16()
			obj[0] = 1
			sink = obj
		}
	})
}

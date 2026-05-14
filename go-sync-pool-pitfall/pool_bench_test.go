package main

import (
	"fmt"
	"runtime"
	"sync"
	"testing"
)

// 强制对象逃逸到堆：通过 interface{} 赋值确保 escape analysis 无法栈分配
//go:noinline
func escapeToHeap(x interface{}) interface{} {
	return x
}

// 不同大小的对象类型
type Obj16 [16]byte
type Obj64 [64]byte
type Obj128 [128]byte
type Obj256 [256]byte
type Obj512 [512]byte
type Obj1K [1024]byte
type Obj4K [4096]byte

// Pool 定义
var (
	pool16  = sync.Pool{New: func() interface{} { return new(Obj16) }}
	pool64  = sync.Pool{New: func() interface{} { return new(Obj64) }}
	pool128 = sync.Pool{New: func() interface{} { return new(Obj128) }}
	pool256 = sync.Pool{New: func() interface{} { return new(Obj256) }}
	pool512 = sync.Pool{New: func() interface{} { return new(Obj512) }}
	pool1K  = sync.Pool{New: func() interface{} { return new(Obj1K) }}
	pool4K  = sync.Pool{New: func() interface{} { return new(Obj4K) }}
)

// --- Pool 方式 ---

func BenchmarkPool16(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool16.Get().(*Obj16)
			obj[0] = 1 // 模拟使用
			pool16.Put(obj)
		}
	})
}

func BenchmarkPool64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool64.Get().(*Obj64)
			obj[0] = 1
			pool64.Put(obj)
		}
	})
}

func BenchmarkPool128(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool128.Get().(*Obj128)
			obj[0] = 1
			pool128.Put(obj)
		}
	})
}

func BenchmarkPool256(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool256.Get().(*Obj256)
			obj[0] = 1
			pool256.Put(obj)
		}
	})
}

func BenchmarkPool512(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool512.Get().(*Obj512)
			obj[0] = 1
			pool512.Put(obj)
		}
	})
}

func BenchmarkPool1K(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool1K.Get().(*Obj1K)
			obj[0] = 1
			pool1K.Put(obj)
		}
	})
}

func BenchmarkPool4K(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool4K.Get().(*Obj4K)
			obj[0] = 1
			pool4K.Put(obj)
		}
	})
}

// --- 直接分配方式 ---

func BenchmarkAlloc16(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj16)
			obj[0] = 1
			escapeToHeap(obj) // 确保不被优化掉
		}
	})
}

func BenchmarkAlloc64(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj64)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

func BenchmarkAlloc128(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj128)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

func BenchmarkAlloc256(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj256)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

func BenchmarkAlloc512(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj512)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

func BenchmarkAlloc1K(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj1K)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

func BenchmarkAlloc4K(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := new(Obj4K)
			obj[0] = 1
			escapeToHeap(obj)
		}
	})
}

// 用于控制并发度的基准测试
func benchmarkPoolWithProcs(b *testing.B, pool *sync.Pool, size int, procs int) {
	old := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(old)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			pool.Put(obj)
		}
	})
}

func benchmarkAllocWithProcs(b *testing.B, size int, procs int) {
	old := runtime.GOMAXPROCS(procs)
	defer runtime.GOMAXPROCS(old)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			buf := make([]byte, size)
			buf[0] = 1
			escapeToHeap(buf)
		}
	})
}

// 二维矩阵测试：不同 GOMAXPROCS
func BenchmarkMatrix(b *testing.B) {
	sizes := []int{16, 64, 128, 256, 512, 1024, 4096}
	procs := []int{1, 4, 8, 16, 32}

	for _, size := range sizes {
		pool := &sync.Pool{New: func() interface{} { return make([]byte, size) }}

		for _, p := range procs {
			name := fmt.Sprintf("Pool/size=%d/procs=%d", size, p)
			b.Run(name, func(b *testing.B) {
				benchmarkPoolWithProcs(b, pool, size, p)
			})

			name = fmt.Sprintf("Alloc/size=%d/procs=%d", size, p)
			b.Run(name, func(b *testing.B) {
				benchmarkAllocWithProcs(b, size, p)
			})
		}
	}
}

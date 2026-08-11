// sizeclass-boundary — 跨 size class 边界分配实测
//
// 目标：验证核心假设"对象大小跨过 size class 边界，分配开销有可测差异"
// 方法：构造 32B/33B、64B/65B、128B/129B 三组结构体，用标准 testing.B 框架
//       实测 ns/op + allocs/op + B/op。
//
// 运行：
//   go test -bench=. -benchmem -run=^$
package main

import (
	"fmt"
	"testing"
	"unsafe"
)

// 通过字段组合构造不同大小的结构体
type S32 struct {
	a, b, c, d int64 // 32
}

type S33 struct {
	a, b, c, d int64
	x          int8 // 33 字节（对齐后 40）
}

type S64 struct {
	a, b, c, d, e, f, g, h int64 // 64
}

type S65 struct {
	a, b, c, d, e, f, g, h int64
	x                     int8  // 65 字节（对齐后 72）
}

type S128 struct {
	a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p int64 // 128
}

type S129 struct {
	a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p int64
	x                                             int8 // 129 字节（对齐后 136）
}

func TestSizes(t *testing.T) {
	fmt.Printf("S32 = %d bytes\n", unsafe.Sizeof(S32{}))
	fmt.Printf("S33 = %d bytes (对齐后)\n", unsafe.Sizeof(S33{}))
	fmt.Printf("S64 = %d bytes\n", unsafe.Sizeof(S64{}))
	fmt.Printf("S65 = %d bytes (对齐后)\n", unsafe.Sizeof(S65{}))
	fmt.Printf("S128 = %d bytes\n", unsafe.Sizeof(S128{}))
	fmt.Printf("S129 = %d bytes (对齐后)\n", unsafe.Sizeof(S129{}))
}

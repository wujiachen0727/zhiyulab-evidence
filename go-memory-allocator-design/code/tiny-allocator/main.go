// tiny-allocator — Go tiny allocator（<16B 小对象合并分配）收益实测
//
// 目标：验证"大量小对象（<16B）走 tiny allocator 合并分配，减少分配次数"
// 方法：对比 4B/8B/16B/32B 对象各分配 100 个，观察 allocs/op 与 B/op。
//       tiny allocator 会把多个 <16B 对象合并到同一个 16B 槽位。
//
// 运行：go test -bench=. -benchmem -benchtime=1s -run=^$
package main

// 4B 对象（两个 int16）
type T4 struct {
	a, b int16
}

// 8B 对象
type T8 struct {
	a int64
}

// 16B 对象（tiny 边界——16B 恰好是 tiny 上限 maxTinySize）
type T16 struct {
	a, b int64
}

// 32B 对象（超过 tiny 上限，直接走 size class）
type T32 struct {
	a, b, c, d int64
}

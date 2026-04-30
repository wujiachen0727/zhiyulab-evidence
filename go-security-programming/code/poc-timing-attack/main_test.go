// Package main benchmarks == vs subtle.ConstantTimeCompare for secret comparison.
//
// PoC E7: "== 字符串比较短路——匹配到前几位就早退。
// 攻击者用海量请求测延迟，能逐位猜出 token。"
//
// 证伪检查：如果现代 CPU 和 Go 编译器把差异吃掉了，结论需要修改。
// 先跑 benchmark，看数据说话。
//
// Run: go test -bench=. -benchmem -count=5
package main

import (
	"crypto/subtle"
	"testing"
)

// 固定目标：32 字节的"密钥"
var target = []byte("0123456789abcdef0123456789abcdef")

// 候选输入 1：第一个字节就不匹配
var guessWrongAtFirst = []byte("X123456789abcdef0123456789abcdef")

// 候选输入 2：最后一个字节才不匹配（用 == 的话要比完全部字节）
var guessWrongAtLast = []byte("0123456789abcdef0123456789abcdeX")

// 候选输入 3：完全匹配
var guessExact = []byte("0123456789abcdef0123456789abcdef")

// sink 防止 DCE 优化（来自 go-network-programming 复盘）
var sinkBool bool

// --- == 比较 ---

// compareNaive 模拟业务代码里的 "if token == expected" 直接比较。
// Go 的 []byte 不能直接用 ==，这里用 string 转换模拟 string == 场景。
func compareNaive(a, b []byte) bool {
	return string(a) == string(b)
}

func BenchmarkNaive_WrongAtFirst(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = compareNaive(guessWrongAtFirst, target)
	}
}

func BenchmarkNaive_WrongAtLast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = compareNaive(guessWrongAtLast, target)
	}
}

func BenchmarkNaive_Exact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = compareNaive(guessExact, target)
	}
}

// --- subtle.ConstantTimeCompare ---

func BenchmarkConstantTime_WrongAtFirst(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = subtle.ConstantTimeCompare(guessWrongAtFirst, target) == 1
	}
}

func BenchmarkConstantTime_WrongAtLast(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = subtle.ConstantTimeCompare(guessWrongAtLast, target) == 1
	}
}

func BenchmarkConstantTime_Exact(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sinkBool = subtle.ConstantTimeCompare(guessExact, target) == 1
	}
}

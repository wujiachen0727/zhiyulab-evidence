package main

import (
	"context"
	"fmt"
	"runtime"
	"sort"
	"strings"
	"time"
)

// leakBySend 向无缓冲、且没有接收者的 channel 发送，发送方永久阻塞
func leakBySend() {
	ch := make(chan int)
	go func() { ch <- 1 }()
}

// leakByContext 派生出子 context 却从不调用 cancel，等待取消的协程永久存活
func leakByContext() {
	ctx, _ := context.WithCancel(context.Background())
	go func() { <-ctx.Done() }()
}

// fixedByBuffer 缓冲够用时，发送方不必等人接就可以退出
func fixedByBuffer() {
	ch := make(chan int, 1)
	go func() { ch <- 1 }()
	time.Sleep(time.Millisecond)
	<-ch
}

// fixedByCancel 显式调用 cancel，等待取消的协程被唤醒并退出
func fixedByCancel() {
	ctx, cancel := context.WithCancel(context.Background())
	go func() { <-ctx.Done() }()
	time.Sleep(time.Millisecond)
	cancel()
}

// measure 记录 f 执行前后协程数的变化
func measure(label string, f func()) {
	time.Sleep(150 * time.Millisecond)
	before := runtime.NumGoroutine()
	f()
	time.Sleep(150 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("  %-38s 协程数 %3d -> %3d（变化 %+d）\n", label, before, after, after-before)
}

func main() {
	fmt.Printf("起始协程数：%d\n\n", runtime.NumGoroutine())

	fmt.Println("== 泄漏形态 1：发送方没人接 ==")
	measure("leakBySend() 调用 100 次", func() {
		for i := 0; i < 100; i++ {
			leakBySend()
		}
	})

	fmt.Println()
	fmt.Println("== 泄漏形态 2：等人取消，但没人取消 ==")
	measure("leakByContext() 调用 50 次", func() {
		for i := 0; i < 50; i++ {
			leakByContext()
		}
	})

	fmt.Println()
	fmt.Println("== 修复版本 ==")
	measure("fixedByBuffer() 调用 100 次", func() {
		for i := 0; i < 100; i++ {
			fixedByBuffer()
		}
	})
	measure("fixedByCancel() 调用 100 次", func() {
		for i := 0; i < 100; i++ {
			fixedByCancel()
		}
	})

	fmt.Println()
	fmt.Println("== 线上怎么看到它：阻塞栈里反复出现同一行 ==")
	buf := make([]byte, 1<<20)
	n := runtime.Stack(buf, true)
	seen := map[string]int{}
	for _, line := range strings.Split(string(buf[:n]), "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "main.leak") {
			seen[trimmed]++
		}
	}
	frames := make([]string, 0, len(seen))
	for frame := range seen {
		frames = append(frames, frame)
	}
	sort.Strings(frames)
	for _, frame := range frames {
		fmt.Printf("  %-32s 出现 %d 次\n", frame, seen[frame])
	}
	fmt.Printf("  当前仍在运行的协程数：%d\n", runtime.NumGoroutine())
}

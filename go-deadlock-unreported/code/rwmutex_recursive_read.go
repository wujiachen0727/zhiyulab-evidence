// RWMutex 递归读阻塞演示
// 运行：go run rwmutex_recursive_read.go
//
// Go 的 RWMutex 是 writer-preference 设计：
// 当有 writer 在等待锁时，新的 reader 不能获取读锁。
// 这导致：ReaderA 持有 RLock → Writer 等待 Lock → ReaderB 尝试 RLock → 阻塞
//
// 关键点：
// - 这不是死锁（没有循环等待），Go runtime 不报错
// - 服务进程正常，但 goroutine 永久阻塞在 RWMutex.RLock()
// - goroutine dump 中可见大量 goroutine 卡在 semacquire

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

func main() {
	var mu sync.RWMutex

	// 1. ReaderA 获取读锁
	mu.RLock()
	fmt.Println("[ReaderA] 获取了读锁")

	// 2. Writer 等待写锁（此时新的 reader 会被阻塞）
	writerDone := make(chan struct{})
	go func() {
		fmt.Println("[Writer] 尝试获取写锁...")
		mu.Lock()
		fmt.Println("[Writer] 获取了写锁")
		time.Sleep(100 * time.Millisecond)
		mu.Unlock()
		fmt.Println("[Writer] 释放了写锁")
		close(writerDone)
	}()

	// 让 Writer 先开始
	time.Sleep(50 * time.Millisecond)

	// 3. ReaderB 尝试获取读锁 —— 被阻塞！
	readerBDone := make(chan struct{})
	go func() {
		fmt.Println("[ReaderB] 尝试获取读锁...")
		mu.RLock()
		fmt.Println("[ReaderB] 获取了读锁")
		mu.RUnlock()
		fmt.Println("[ReaderB] 释放了读锁")
		close(readerBDone)
	}()

	time.Sleep(100 * time.Millisecond)

	// 打印 goroutine dump 关键部分
	fmt.Println("\n=== goroutine dump 特征信号 ===")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	dump := string(buf[:n])

	// 搜索关键信号
	if contains(dump, "semacquire") {
		fmt.Println("✅ 发现 semacquire 信号（RWMutex 阻塞的特征）")
	}
	if contains(dump, "RLock") {
		fmt.Println("✅ 发现 goroutine 卡在 RLock()")
	}

	fmt.Println("\n--- 完整 dump（前 2000 字符）---")
	if len(dump) > 2000 {
		fmt.Println(dump[:2000])
	} else {
		fmt.Println(dump)
	}

	// 释放 ReaderA 的读锁，让 Writer 和 ReaderB 继续
	fmt.Println("\n[ReaderA] 释放读锁...")
	mu.RUnlock()

	<-writerDone
	<-readerBDone

	fmt.Println("\n✅ 所有 goroutine 正常退出")
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

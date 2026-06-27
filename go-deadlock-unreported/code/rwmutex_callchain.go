// RWMutex 复杂调用链——看起来合理的代码如何卡住
// 运行：go run rwmutex_callchain.go
//
// 模拟真实场景：一个带缓存的查询服务
// - GetData(key): 先 RLock 查缓存，查到直接返回
// - RefreshCache(): 全量刷新缓存，需要 Lock
// - 如果在 RefreshCache 等待期间，大量请求同时调用 GetData → 全部卡在 RLock()
//
// 这种场景在线上最常见的表现形式：
// 服务进程还活着，健康检查通过（简单的 /health 不需要 RLock），
// 但业务接口全部超时。

package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

type Cache struct {
	mu   sync.RWMutex
	data map[string]string
}

func NewCache() *Cache {
	return &Cache{
		data: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}
}

func (c *Cache) GetData(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.data[key]
	return val, ok
}

func (c *Cache) RefreshCache() {
	fmt.Println("[Refresh] 等待写锁...")
	c.mu.Lock()
	defer c.mu.Unlock()
	fmt.Println("[Refresh] 开始刷新缓存...")
	time.Sleep(500 * time.Millisecond) // 模拟耗时刷新
	c.data["key1"] = "new_value1"
	c.data["key3"] = "value3"
	fmt.Println("[Refresh] 缓存刷新完成")
}

func main() {
	cache := NewCache()

	// 1. 先让一个 goroutine 持有读锁
	readDone := make(chan struct{})
	go func() {
		cache.mu.RLock()
		fmt.Println("[Reader1] 持有读锁，模拟长查询...")
		time.Sleep(1 * time.Second)
		cache.mu.RUnlock()
		fmt.Println("[Reader1] 释放读锁")
		close(readDone)
	}()
	time.Sleep(50 * time.Millisecond)

	// 2. Refresh 尝试获取写锁 —— 等 Reader1 释放
	go func() {
		cache.RefreshCache()
	}()

	time.Sleep(50 * time.Millisecond)

	// 3. 多个并发请求尝试读取 —— 全部被阻塞
	concurrentReaders := 5
	for i := 0; i < concurrentReaders; i++ {
		i := i
		go func() {
			fmt.Printf("[Reader%d] 尝试读取 key1...\n", i+2)
			if val, ok := cache.GetData("key1"); ok {
				fmt.Printf("[Reader%d] 读取成功: %s\n", i+2, val)
			}
		}()
	}

	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n=== goroutine dump 特征信号 ===")
	buf := make([]byte, 1<<16)
	n := runtime.Stack(buf, true)
	dump := string(buf[:n])

	if contains(dump, "Lock") {
		fmt.Println("✅ 发现 Lock 信号（writer 在等锁）")
	}
	if contains(dump, "RLock") {
		fmt.Println("✅ 发现 RLock 信号（多个 reader 被阻塞）")
	}
	if contains(dump, "semacquire") {
		fmt.Println("✅ 发现 semacquire 信号（RWMutex 阻塞特征）")
	}

	fmt.Println("\n--- 关键 dump 片段（仅包含 RWMutex 相关 goroutine）---")
	lines := splitLines(dump)
	for _, line := range lines {
		if contains(line, "sync.RWMutex") || contains(line, "semacquire") || contains(line, "Lock(") || contains(line, "RLock(") {
			fmt.Println(line)
		}
	}

	<-readDone
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

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

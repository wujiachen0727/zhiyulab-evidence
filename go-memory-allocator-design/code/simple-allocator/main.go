// allocsim — 简化 allocator 三版本对比（证伪实验）
//
// 目标：验证核心假设"无缓存 vs 有缓存 vs 带大小类，分配延迟有显著差异"
// 三个版本：
//   v0: 裸全局链表（无缓存，每次分配全局锁 + 线性扫）
//   v1: 每线程局部缓存（每次分配从 P-local 缓存取，不够再向全局要）
//   v2: 大小类 + 每线程缓存（按大小分类，命中缓存不锁）
//
// 这只是教学模拟——真实 Go allocator 复杂得多，但三者的核心权衡可以复现：
//   锁竞争（无缓存最高）、缓存命中（v1/v2 高）、碎片（v2 有大小类更可控）
package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
)

const (
	objects    = 1_000_000 // 每 goroutine 分配数
	goroutines = 8
)

// ---------- v0: 无缓存（全局链表 + 全局锁） ----------
type blockV0 struct {
	next *blockV0
	_    [24]byte // 模拟对象负载
}

type allocV0 struct {
	mu    sync.Mutex
	freelist *blockV0
	total int64
}

func (a *allocV0) alloc() *blockV0 {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.total++
	if a.freelist != nil {
		b := a.freelist
		a.freelist = b.next
		return b
	}
	return &blockV0{}
}

func (a *allocV0) free(b *blockV0) {
	a.mu.Lock()
	defer a.mu.Unlock()
	b.next = a.freelist
	a.freelist = b
}

// ---------- v1: 每 P 本地缓存（无大小类，跨 P 竞争小） ----------
type blockV1 struct {
	next *blockV1
	_    [24]byte
}

type pCacheV1 struct {
	head *blockV1
	mu   sync.Mutex // 简化：cache miss 时才碰全局
}

type allocV1 struct {
	globalMu sync.Mutex
	global   *blockV1
	caches   [32]pCacheV1
	total    int64
}

func (a *allocV1) alloc(p int) *blockV1 {
	c := &a.caches[p]
	if c.head != nil {
		b := c.head
		c.head = b.next
		return b
	}
	// cache miss → 向全局要
	a.globalMu.Lock()
	if a.global != nil {
		b := a.global
		a.global = b.next
		a.globalMu.Unlock()
		return b
	}
	a.globalMu.Unlock()
	return &blockV1{}
}

func (a *allocV1) free(p int, b *blockV1) {
	c := &a.caches[p]
	c.mu.Lock()
	b.next = c.head
	c.head = b
	c.mu.Unlock()
}

// ---------- v2: 大小类 + 每 P 缓存 ----------
type blockV2 struct {
	next *blockV2
	_    [24]byte
}

type classCache struct {
	head *blockV2
	mu   sync.Mutex
}

type allocV2 struct {
	globalMu sync.Mutex
	global   [4]*blockV2 // 4 个大小类
	caches   [32][4]classCache
}

func (a *allocV2) alloc(p, cls int) *blockV2 {
	c := &a.caches[p][cls]
	if c.head != nil {
		b := c.head
		c.head = b.next
		return b
	}
	a.globalMu.Lock()
	if a.global[cls] != nil {
		b := a.global[cls]
		a.global[cls] = b.next
		a.globalMu.Unlock()
		return b
	}
	a.globalMu.Unlock()
	return &blockV2{}
}

func (a *allocV2) free(p, cls int, b *blockV2) {
	c := &a.caches[p][cls]
	c.mu.Lock()
	b.next = c.head
	c.head = b
	c.mu.Unlock()
}

// ---------- 基准 ----------
// 模式 A：混合分配+释放（模拟典型工作负载，缓存被复用）
// 模式 B：纯分配（模拟突发分配，锁竞争最激烈）
func runMode(mixed bool) (time.Duration, time.Duration, time.Duration) {
	runtime.GOMAXPROCS(goroutines)

	// v0
	var v0 allocV0
	start := time.Now()
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var b *blockV0
			for i := 0; i < objects; i++ {
				b = v0.alloc()
				if mixed && i%2 == 0 {
					v0.free(b)
				}
			}
			_ = b
		}()
	}
	wg.Wait()
	d0 := time.Since(start)

	// v1
	var v1 allocV1
	start = time.Now()
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			var b *blockV1
			for i := 0; i < objects; i++ {
				b = v1.alloc(p)
				if mixed && i%2 == 0 {
					v1.free(p, b)
				}
			}
			_ = b
		}(g % 32)
	}
	wg.Wait()
	d1 := time.Since(start)

	// v2
	var v2 allocV2
	start = time.Now()
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func(p int) {
			defer wg.Done()
			var b *blockV2
			for i := 0; i < objects; i++ {
				cls := i % 4
				b = v2.alloc(p, cls)
				if mixed && i%2 == 0 {
					v2.free(p, cls, b)
				}
			}
			_ = b
		}(g % 32)
	}
	wg.Wait()
	d2 := time.Since(start)

	return d0, d1, d2
}

func main() {
	for _, mixed := range []bool{true, false} {
		label := "混合分配+释放"
		if !mixed {
			label = "纯分配（锁竞争最激烈）"
		}
		d0, d1, d2 := runMode(mixed)
		fmt.Printf("=== %s（%d goroutines × %d 次分配）===\n", label, goroutines, objects)
		fmt.Printf("v0 无缓存(全局锁):   %10s\n", d0)
		fmt.Printf("v1 每P缓存(无分类):  %10s\n", d1)
		fmt.Printf("v2 大小类+每P缓存:  %10s\n", d2)
		fmt.Printf("\nv1/v0 加速比: %.1fx\n", float64(d0)/float64(d1))
		fmt.Printf("v2/v0 加速比: %.1fx\n", float64(d0)/float64(d2))
		fmt.Printf("v2/v1 加速比: %.1fx\n\n", float64(d1)/float64(d2))
	}
}

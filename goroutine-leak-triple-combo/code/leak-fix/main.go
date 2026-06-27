// leak-fix: 三重修复对比实验
// 逐步加入修复手段，观察 goroutine 泄漏是否被遏制
package main

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"
)

func startSlowServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(30 * time.Second)
		w.Write([]byte("ok"))
	})
	go http.ListenAndServe(":18081", mux)
	time.Sleep(100 * time.Millisecond)
}

// 实验 1：只加 client 超时
func fixClientTimeout() {
	client := &http.Client{
		Timeout: 3 * time.Second, // 修复点：加 client 级超时
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := client.Get("http://localhost:18081/slow")
			if err != nil {
				return // 超时后 goroutine 正常退出
			}
			defer resp.Body.Close()
		}()
	}
	wg.Wait()
}

// 实验 2：只加 context cancel（不加 client 超时）
func fixContextOnly() {
	client := &http.Client{} // 无 client 超时
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:18081/slow", nil)
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}()
	}
	wg.Wait()
}

// 实验 3：client 超时 + context cancel（双保险）
func fixBoth() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()
			req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost:18081/slow", nil)
			resp, err := client.Do(req)
			if err != nil {
				return
			}
			defer resp.Body.Close()
		}()
	}
	wg.Wait()
}

func main() {
	startSlowServer()

	fmt.Println("=== 三重修复对比实验 ===\n")

	// 实验 1：client 超时
	fmt.Println("--- 实验 1：只加 client.Timeout = 3s ---")
	before := runtime.NumGoroutine()
	fixClientTimeout()
	time.Sleep(500 * time.Millisecond)
	after := runtime.NumGoroutine()
	fmt.Printf("修复前 goroutine: %d → 修复后: %d（泄漏: %d）\n\n", before, after, after-before)

	// 实验 2：context cancel
	fmt.Println("--- 实验 2：只加 context.WithTimeout = 3s ---")
	before = runtime.NumGoroutine()
	fixContextOnly()
	time.Sleep(500 * time.Millisecond)
	after = runtime.NumGoroutine()
	fmt.Printf("修复前 goroutine: %d → 修复后: %d（泄漏: %d）\n\n", before, after, after-before)

	// 实验 3：双保险
	fmt.Println("--- 实验 3：client.Timeout + context.WithTimeout ---")
	before = runtime.NumGoroutine()
	fixBoth()
	time.Sleep(500 * time.Millisecond)
	after = runtime.NumGoroutine()
	fmt.Printf("修复前 goroutine: %d → 修复后: %d（泄漏: %d）\n\n", before, after, after-before)

	fmt.Println("=== 结论 ===")
	fmt.Println("✅ 三种修复都能遏制泄漏")
	fmt.Println("⚠️  但只有 client.Timeout + context 的组合最可靠：")
	fmt.Println("   - client.Timeout 是兜底（即使忘了加 context）")
	fmt.Println("   - context 是精确控制（可以按业务场景设不同超时）")
	fmt.Println("   - 两者不冲突，谁先到谁生效")
}

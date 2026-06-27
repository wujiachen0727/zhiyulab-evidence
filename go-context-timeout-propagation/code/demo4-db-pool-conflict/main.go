package main

import (
	"context"
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "github.com/mattn/go-sqlite3" // 用 SQLite 演示，无需外部 DB
)

// Demo 4: DB 连接池超时和 context 超时会打架
// 证明：当连接池满时，新请求等待连接的时间和 context 超时会产生竞争。
// 你以为 context 5s 超时 = 查询有 5s 执行时间，
// 实际上等连接池可能就花了 4s，查询只剩 1s。

func main() {
	fmt.Println("=== Demo: DB 连接池超时 vs context 超时 ===")
	fmt.Println()

	// 创建一个内存 SQLite DB（演示用）
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		// 如果没有 sqlite3 driver，用模拟方式演示
		fmt.Println("(SQLite driver 不可用，使用模拟演示)")
		simulateDemo()
		return
	}
	defer db.Close()

	// 关键设置：连接池只有 1 个连接
	db.SetMaxOpenConns(1)
	fmt.Println("连接池配置: MaxOpenConns = 1")
	fmt.Println()

	// 创建测试表
	db.Exec("CREATE TABLE test (id INTEGER PRIMARY KEY, data TEXT)")

	// 先占用唯一的连接（模拟一个慢查询）
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		tx, _ := db.Begin()
		fmt.Println("[goroutine-1] 占用连接，执行慢事务 (4s)...")
		time.Sleep(4 * time.Second)
		tx.Commit()
		fmt.Println("[goroutine-1] 慢事务完成，释放连接")
	}()

	// 等一下确保 goroutine-1 先拿到连接
	time.Sleep(100 * time.Millisecond)

	// 新请求带 5s context timeout 来查询
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("[goroutine-2] 带 5s timeout 发起查询...")
	fmt.Printf("[goroutine-2] 开始等待连接: %s\n", time.Now().Format("15:04:05.000"))

	start := time.Now()
	row := db.QueryRowContext(ctx, "SELECT 1")

	waitTime := time.Since(start)
	fmt.Printf("[goroutine-2] 拿到连接: %s (等了 %v)\n",
		time.Now().Format("15:04:05.000"), waitTime.Round(time.Millisecond))

	var result int
	err = row.Scan(&result)
	if err != nil {
		fmt.Printf("[goroutine-2] 查询失败: %v\n", err)
	} else {
		queryTime := time.Since(start) - waitTime
		fmt.Printf("[goroutine-2] 查询成功: result=%d\n", result)
		fmt.Printf("[goroutine-2] 实际查询耗时: %v\n", queryTime.Round(time.Millisecond))
	}

	remaining := time.Until(func() time.Time { d, _ := ctx.Deadline(); return d }())
	fmt.Printf("[goroutine-2] context 剩余时间: %v\n", remaining.Round(time.Millisecond))
	fmt.Println()
	fmt.Println("结论:")
	fmt.Println("  你以为 5s timeout = 查询有 5s 执行时间")
	fmt.Printf("  实际: 等连接池花了 ~%v，查询只剩 ~%v\n",
		waitTime.Round(time.Millisecond),
		(5*time.Second - waitTime).Round(time.Millisecond))

	wg.Wait()
}

// simulateDemo 在没有 SQLite driver 时用纯 Go 模拟竞争效果
func simulateDemo() {
	fmt.Println("\n--- 模拟演示 ---")
	fmt.Println("场景: MaxOpenConns=1, 一个 4s 慢查询占着连接")
	fmt.Println()

	// 模拟连接池：容量 1
	pool := make(chan struct{}, 1)
	pool <- struct{}{} // 放入一个连接

	var wg sync.WaitGroup
	wg.Add(1)

	// goroutine-1: 占用连接 4 秒
	go func() {
		defer wg.Done()
		conn := <-pool // 取走连接
		fmt.Println("[goroutine-1] 拿到连接，执行 4s 慢操作...")
		time.Sleep(4 * time.Second)
		pool <- conn // 归还连接
		fmt.Println("[goroutine-1] 操作完成，归还连接")
	}()

	time.Sleep(100 * time.Millisecond) // 确保 g1 先拿到

	// goroutine-2: 带 5s timeout 等连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println("[goroutine-2] 带 5s timeout，等待连接...")
	start := time.Now()

	select {
	case conn := <-pool:
		waitTime := time.Since(start)
		fmt.Printf("[goroutine-2] 等了 %v 拿到连接\n", waitTime.Round(time.Millisecond))

		deadline, _ := ctx.Deadline()
		remaining := time.Until(deadline)
		fmt.Printf("[goroutine-2] context 剩余: %v\n", remaining.Round(time.Millisecond))
		fmt.Printf("[goroutine-2] 你以为有 5s 做查询，实际只剩 %v\n", remaining.Round(time.Millisecond))
		pool <- conn
	case <-ctx.Done():
		fmt.Printf("[goroutine-2] context 超时，连接都没等到! (%v)\n", ctx.Err())
	}

	fmt.Println("\n结论: context timeout 包含了等连接池的时间。")
	fmt.Println("     连接池满时，你的 '5s 查询超时' 可能 4s 都在排队。")

	wg.Wait()
}

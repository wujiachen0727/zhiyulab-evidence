// 实验 E1：进程崩溃后 Redis 锁不释放，新客户端必须等到 TTL 过期
//
// 复现路径：
//   1. 客户端 A 用 SETNX 获锁（TTL=10s），但模拟"崩溃"——不调用 DEL
//   2. 客户端 B 不停尝试 SETNX，记录从开始到拿到锁的等待时间
//   3. 观察：B 等了多久？答案接近 TTL（10s）
//
// 运行前置：本地 6379 端口有 Redis 实例
//   docker run -d --name dl-redis -p 6379:6379 redis:7-alpine
//
// 运行：go run main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	lockKey  = "order:1001:lock"
	ttl      = 10 * time.Second
	pollGap  = 200 * time.Millisecond
	maxWait  = 30 * time.Second
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer rdb.Close()

	ctx := context.Background()

	// 清理之前可能残留的锁
	rdb.Del(ctx, lockKey)

	// ---- 客户端 A：获锁但"崩溃" ----
	t0 := time.Now()
	okA, err := rdb.SetNX(ctx, lockKey, "client_A", ttl).Result()
	if err != nil {
		fmt.Fprintln(os.Stderr, "SETNX A error:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%6s] client_A SETNX → %v   (TTL=%v)\n", elapsed(t0), okA, ttl)
	fmt.Printf("[T+%6s] client_A 模拟 kill -9，不调用 DEL/UNLINK\n", elapsed(t0))

	// ---- 客户端 B：开始尝试获锁，记录等待时间 ----
	fmt.Printf("[T+%6s] client_B 开始尝试 SETNX，每 %v 一次\n", elapsed(t0), pollGap)

	deadline := time.Now().Add(maxWait)
	attempts := 0
	for time.Now().Before(deadline) {
		attempts++
		okB, err := rdb.SetNX(ctx, lockKey, "client_B", ttl).Result()
		if err != nil {
			fmt.Fprintln(os.Stderr, "SETNX B error:", err)
			os.Exit(1)
		}
		if okB {
			waited := time.Since(t0)
			fmt.Printf("[T+%6s] client_B SETNX → true  ✅ 拿到锁（第 %d 次尝试）\n", elapsed(t0), attempts)
			fmt.Printf("\n=== 实验结果 ===\n")
			fmt.Printf("client_B 总等待时间：%v\n", waited.Round(time.Millisecond))
			fmt.Printf("client_B 尝试次数：  %d\n", attempts)
			fmt.Printf("锁 TTL：             %v\n", ttl)
			fmt.Printf("结论：client_A 崩溃后未释放锁，client_B 必须等到 TTL 过期才能继续\n")

			rdb.Del(ctx, lockKey)
			return
		}
		time.Sleep(pollGap)
	}

	fmt.Fprintf(os.Stderr, "client_B 在 %v 内未能获锁，超时退出\n", maxWait)
	os.Exit(2)
}

func elapsed(t0 time.Time) string {
	return fmt.Sprintf("%5.2fs", time.Since(t0).Seconds())
}

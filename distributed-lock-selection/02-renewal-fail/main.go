// 实验 E2：watchdog 续期被 STW 冻结，锁过期被抢，A 苏醒后双持
//
// 复现路径：
//   1. 客户端 A 用 SETNX 获锁（TTL=3s），启动 watchdog 每 1s EXPIRE 续期
//   2. A 业务跑 2s 后整个进程"卡死"5s（用 sleep 模拟 STW > TTL）——watchdog 也跟着停
//   3. 期间客户端 B 在另一个 goroutine 不断尝试 SETNX
//   4. TTL 过期，B 拿到锁
//   5. A 苏醒，GET key 检查锁 owner，发现已经不是自己了
//
// 这个实验展示的是 watchdog 不解决问题的边界条件：当 STW 时间 > TTL 时，
// watchdog 也无法续期（它和业务在同一个进程里），锁会被自动释放并被他人抢占。
//
// 运行前置：本地 6379 端口有 Redis 实例
//   docker run -d --name dl-redis -p 6379:6379 redis:7-alpine
package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	lockKey       = "order:1002:lock"
	ttl           = 3 * time.Second
	renewInterval = 1 * time.Second
	stwDuration   = 5 * time.Second // STW 时长 > TTL，触发锁过期
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer rdb.Close()
	ctx := context.Background()
	rdb.Del(ctx, lockKey)

	t0 := time.Now()
	var wg sync.WaitGroup

	// ---- 客户端 A：获锁 → 启 watchdog → 业务跑一会 → STW → 苏醒查锁 ----
	wg.Add(1)
	go func() {
		defer wg.Done()
		ok, err := rdb.SetNX(ctx, lockKey, "client_A", ttl).Result()
		if err != nil {
			fmt.Fprintln(os.Stderr, "A SETNX:", err)
			os.Exit(1)
		}
		fmt.Printf("[T+%s][A] SETNX client_A → %v (TTL=%v)\n", elapsed(t0), ok, ttl)

		// watchdog goroutine：每 renewInterval 续期一次
		stopWatchdog := make(chan struct{})
		go func() {
			tick := time.NewTicker(renewInterval)
			defer tick.Stop()
			for {
				select {
				case <-tick.C:
					n, _ := rdb.Expire(ctx, lockKey, ttl).Result()
					fmt.Printf("[T+%s][A.watchdog] EXPIRE → 续期成功？%v\n", elapsed(t0), n)
				case <-stopWatchdog:
					fmt.Printf("[T+%s][A.watchdog] 收到停止信号，退出\n", elapsed(t0))
					return
				}
			}
		}()

		// 业务跑 2 秒（watchdog 应续期 2 次）
		time.Sleep(2 * time.Second)

		// ---- 关键：模拟 STW（GC pause / OS 挂起）----
		// 在真实场景里这是 GC、调度延迟、虚拟机迁移等导致的"主线程卡死"。
		// watchdog 是同一进程的 goroutine，没法在 STW 期间运行。
		// 这里用一个不可中断的 sleep 模拟整个进程被冻结。
		fmt.Printf("[T+%s][A] ⚠️ 模拟 STW %v 开始（业务 + watchdog 都被冻结）\n", elapsed(t0), stwDuration)
		// 用一个无 select 的 time.Sleep 模拟：goroutine 不可被打断，阻塞 watchdog 续期
		// 注意：真实 STW 整个进程都停，这里仅业务 goroutine 停，watchdog 仍在跑——
		// 所以我们也要在 STW 开始前显式停掉 watchdog 来模拟"watchdog 也停了"
		close(stopWatchdog)
		time.Sleep(stwDuration)
		fmt.Printf("[T+%s][A] STW 结束，业务继续\n", elapsed(t0))

		// ---- 苏醒后检查锁是否还是自己的 ----
		val, err := rdb.Get(ctx, lockKey).Result()
		if err == redis.Nil {
			fmt.Printf("[T+%s][A] GET 结果：锁不存在！\n", elapsed(t0))
		} else if err != nil {
			fmt.Fprintln(os.Stderr, "A GET:", err)
		} else {
			fmt.Printf("[T+%s][A] GET 锁 owner = %q\n", elapsed(t0), val)
			if val == "client_A" {
				fmt.Printf("[T+%s][A] 自认为还持锁，准备扣库存（双持发生）\n", elapsed(t0))
			} else {
				fmt.Printf("[T+%s][A] ⚠️ 锁已易主——但仅在主动 GET 时才察觉\n", elapsed(t0))
				fmt.Printf("[T+%s][A] 如果代码没做 owner 校验就直接 DEL+扣库存，就是双持事故\n", elapsed(t0))
			}
		}
	}()

	// ---- 客户端 B：等到 A 持锁后开始抢锁 ----
	wg.Add(1)
	go func() {
		defer wg.Done()
		// 等 A 先把锁拿到
		time.Sleep(500 * time.Millisecond)

		gotAt := time.Time{}
		for i := 0; i < 100; i++ {
			ok, _ := rdb.SetNX(ctx, lockKey, "client_B", ttl).Result()
			if ok && gotAt.IsZero() {
				gotAt = time.Now()
				fmt.Printf("[T+%s][B] SETNX client_B → true  ✅ 抢到锁\n", elapsed(t0))
				// 拿到后保持持锁观察 A 的反应
				time.Sleep(4 * time.Second)
				return
			}
			time.Sleep(300 * time.Millisecond)
		}
	}()

	wg.Wait()

	fmt.Printf("\n=== 实验结果 ===\n")
	fmt.Printf("STW 时长 %v > TTL %v → watchdog 来不及续期\n", stwDuration, ttl)
	fmt.Printf("结论：A 苏醒后锁的 owner 已变成 client_B；如果业务代码没做 owner 校验，A 会以为自己还持锁，造成双持\n")
	fmt.Printf("修复方向：1) 加 owner 校验（GET+CAS DEL）；2) Redisson 的 fencing token；3) 改用 etcd lease（lease 在客户端断连时由 server 自动 revoke）\n")

	rdb.Del(ctx, lockKey)
}

func elapsed(t0 time.Time) string {
	return fmt.Sprintf("%5.2fs", time.Since(t0).Seconds())
}

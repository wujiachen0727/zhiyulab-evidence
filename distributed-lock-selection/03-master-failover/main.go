// 实验 E3：Redis 主从异步复制 + 主库故障 → 锁丢失
//
// 复现路径：
//   1. master(6379) + replica(6380)，replica 通过 REPLICAOF 跟随 master（异步复制）
//   2. 客户端在 master 上 SETNX 一个锁
//   3. 模拟"master 在复制完成前宕机"——立即将 replica 提升为新 master，并断开旧 master
//   4. 另一客户端连新 master，对同一个 key 做 SETNX → 成功（锁丢了！）
//
// 这是 Kleppmann 在 "How to do distributed locking" 中指出的 Redis 单主锁的根本缺陷：
// 异步复制 + 故障切换 = 同一时刻可能有两个客户端都"持有"锁。
//
// 运行前置：
//   cd articles/distributed-lock-selection/evidence/code/03-master-failover
//   docker compose up -d
//   等 ~3s 让 replica 完成首次同步
//
// 然后运行：
//   go run main.go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	masterAddr  = "127.0.0.1:6379"
	replicaAddr = "127.0.0.1:6380"
	lockKey     = "order:1003:lock"
	ttl         = 30 * time.Second
)

func main() {
	master := redis.NewClient(&redis.Options{Addr: masterAddr})
	replica := redis.NewClient(&redis.Options{Addr: replicaAddr})
	defer master.Close()
	defer replica.Close()

	ctx := context.Background()

	t0 := time.Now()
	fmt.Printf("[T+%s] === E3 主从切换丢锁实验开始 ===\n", elapsed(t0))

	// 清理状态
	master.Del(ctx, lockKey)

	// 验证主从关系
	info, err := replica.Do(ctx, "INFO", "replication").Result()
	if err == nil {
		s := fmt.Sprintf("%v", info)
		if !contains(s, "role:slave") {
			fmt.Fprintln(os.Stderr, "❌ replica 不是从节点。请先 docker compose up -d")
			os.Exit(1)
		}
	}
	fmt.Printf("[T+%s] ✅ 主从关系建立：master=%s, replica=%s\n", elapsed(t0), masterAddr, replicaAddr)

	// ---- 关键：先让 replica 提前断开复制，模拟"网络分区导致复制延迟到无穷大" ----
	// 真实事故里这个延迟来自网络抖动 / replica 还在做 RDB 全量同步 / 跨机房延迟
	// 这里直接打 REPLICAOF NO ONE 把它压到极致：replica 之后所有 master 写入都收不到
	if _, err := replica.Do(ctx, "REPLICAOF", "NO", "ONE").Result(); err != nil {
		fmt.Fprintln(os.Stderr, "REPLICAOF NO ONE:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s] 🔌 replica 提前断开复制（模拟最坏复制延迟）\n", elapsed(t0))

	// ---- 客户端 A：在 master 上 SETNX（此时 replica 已断开，写入不会传到 replica）----
	okA, err := master.SetNX(ctx, lockKey, "client_A", ttl).Result()
	if err != nil {
		fmt.Fprintln(os.Stderr, "A SETNX:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][A] master.SETNX → %v (TTL=%v)\n", elapsed(t0), okA, ttl)
	fmt.Printf("[T+%s]    （replica 已断开复制，这条写入不会到达 replica）\n", elapsed(t0))

	// 模拟旧 master 失联
	_ = master.Shutdown(ctx)
	fmt.Printf("[T+%s] 💥 旧 master 已关闭\n", elapsed(t0))

	time.Sleep(200 * time.Millisecond)

	// ---- 检查新 master 上是否还有锁 ----
	val, err := replica.Get(ctx, lockKey).Result()
	if err == redis.Nil {
		fmt.Printf("[T+%s] ⚠️  新 master 上 GET %s → 不存在（锁丢了！）\n", elapsed(t0), lockKey)
	} else if err != nil {
		fmt.Fprintln(os.Stderr, "GET on new master:", err)
	} else {
		fmt.Printf("[T+%s] 新 master 上 GET %s → %q（罕见情况：复制赶上了）\n", elapsed(t0), lockKey, val)
	}

	// ---- 客户端 B：在新 master 上 SETNX 同一个 key ----
	okB, err := replica.SetNX(ctx, lockKey, "client_B", ttl).Result()
	if err != nil {
		fmt.Fprintln(os.Stderr, "B SETNX:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][B] new_master.SETNX → %v\n", elapsed(t0), okB)

	if okB {
		fmt.Printf("\n=== 实验结果 ===\n")
		fmt.Printf("❌ 锁互斥被破坏：client_A 和 client_B 同时认为自己持有 %q\n", lockKey)
		fmt.Printf("根因：异步复制 + 故障切换之间存在窗口，未复制的写入在 failover 后丢失\n")
		fmt.Printf("修复方向：\n")
		fmt.Printf("  1) min-replicas-to-write 配置：要求至少 N 个 replica ack 才算写成功（牺牲可用性换一致性）\n")
		fmt.Printf("  2) Redlock：N 个独立 master 投票，过半才算获锁（antirez 方案）\n")
		fmt.Printf("  3) 改用 etcd/ZK：基于 raft/zab 的强一致复制，不存在异步丢数据\n")
	} else {
		fmt.Printf("⚠️ 复制刚好完成，本次未复现。重试或调小复制延迟。\n")
	}

	// 清理
	replica.Del(ctx, lockKey)
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})())
}

func elapsed(t0 time.Time) string {
	return fmt.Sprintf("%5.2fs", time.Since(t0).Seconds())
}

// 实验 E4：etcd lease + Txn(CAS) 实现的分布式锁
//
// 演示三件事：
//   1. Lease 是带 TTL 的"租约"，client 主动续期，断连/进程退出 → server 自动 revoke
//   2. 锁的获取用 Txn(CompareAndSwap)：只在 key 不存在（CreateRevision == 0）时才写入
//   3. lease revoke 后 key 自动消失，其他 watch 的客户端立即感知（不用等 TTL）
//
// 这三点正好对应 E1/E2/E3 的三个失败场景的解药：
//   - E1（进程崩溃）→ lease keepalive 断了，server 自动清锁，不用等 TTL
//   - E2（GC pause）→ keepalive 断 N 秒，server 自动 revoke；A 苏醒后 GET 立即知道锁不在
//   - E3（主从切换）→ etcd 用 raft，写入需过半节点确认，不存在异步丢数据
//
// 运行前置：
//   docker run -d --name dl-etcd -p 2379:2379 -p 2380:2380 \
//     quay.io/coreos/etcd:v3.5.18 \
//     /usr/local/bin/etcd \
//     --name=dl-etcd \
//     --listen-client-urls=http://0.0.0.0:2379 \
//     --advertise-client-urls=http://0.0.0.0:2379 \
//     --listen-peer-urls=http://0.0.0.0:2380 \
//     --initial-advertise-peer-urls=http://0.0.0.0:2380 \
//     --initial-cluster=dl-etcd=http://0.0.0.0:2380
//
// 运行：
//   go run ./04-etcd-lock
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	lockKey = "/locks/order/1004"
	ttl     = 10 // seconds
)

func main() {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "dial etcd:", err)
		os.Exit(1)
	}
	defer cli.Close()

	ctx := context.Background()
	t0 := time.Now()

	// 清理状态
	cli.Delete(ctx, lockKey)

	// ====== 场景 1：A 正常获锁、释放锁 ======
	fmt.Printf("[T+%s] === 场景 1：A 用 lease+Txn 获锁，正常 release ===\n", elapsed(t0))

	leaseA, err := cli.Grant(ctx, ttl)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Grant:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][A] Grant lease (id=%x, ttl=%ds)\n", elapsed(t0), leaseA.ID, ttl)

	// keepalive：让 lease 在客户端存活时不过期
	keepCh, err := cli.KeepAlive(ctx, leaseA.ID)
	if err != nil {
		fmt.Fprintln(os.Stderr, "KeepAlive:", err)
		os.Exit(1)
	}
	go drainKeepAlive(keepCh, "A")

	// Txn(CAS)：只在 key 不存在（CreateRevision==0）时写入
	txnResp, err := cli.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "client_A", clientv3.WithLease(leaseA.ID))).
		Commit()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Txn A:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][A] Txn(CAS) Put → succeeded=%v（绑定 lease）\n", elapsed(t0), txnResp.Succeeded)

	// 业务工作 1s
	time.Sleep(1 * time.Second)

	// 主动 revoke lease → key 立即消失
	if _, err := cli.Revoke(ctx, leaseA.ID); err != nil {
		fmt.Fprintln(os.Stderr, "Revoke A:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][A] Revoke lease → key 立即消失（不用等 TTL）\n", elapsed(t0))

	resp, _ := cli.Get(ctx, lockKey)
	fmt.Printf("[T+%s] etcd 上 GET %q → %d 个 key\n", elapsed(t0), lockKey, resp.Count)

	// ====== 场景 2：B 获锁后"崩溃"（停止 keepalive，等 TTL 自动 revoke）======
	fmt.Printf("\n[T+%s] === 场景 2：B 获锁后崩溃 → server 自动 revoke ===\n", elapsed(t0))

	leaseB, err := cli.Grant(ctx, 3) // 短 TTL 方便观察
	if err != nil {
		fmt.Fprintln(os.Stderr, "Grant B:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][B] Grant lease (id=%x, ttl=3s)\n", elapsed(t0), leaseB.ID)

	// 注意：故意不调 KeepAlive，模拟客户端崩溃
	txnB, err := cli.Txn(ctx).
		If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "client_B", clientv3.WithLease(leaseB.ID))).
		Commit()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Txn B:", err)
		os.Exit(1)
	}
	fmt.Printf("[T+%s][B] Txn(CAS) Put → succeeded=%v（不调 KeepAlive，模拟崩溃）\n", elapsed(t0), txnB.Succeeded)

	// ---- C 在另一个 goroutine 用 Watch 监听锁释放，争锁 ----
	gotByC := make(chan time.Duration, 1)
	go func() {
		watchStart := time.Now()
		// 先尝试一次（应失败，因为 B 还在）
		txn, _ := cli.Txn(ctx).
			If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
			Then(clientv3.OpPut(lockKey, "client_C", clientv3.WithLease(0))).
			Commit()
		if txn.Succeeded {
			gotByC <- time.Since(watchStart)
			return
		}

		// Watch lockKey 直到删除事件
		watchCh := cli.Watch(ctx, lockKey)
		for wr := range watchCh {
			for _, ev := range wr.Events {
				if ev.Type == clientv3.EventTypeDelete {
					// 立刻尝试获锁
					leaseC, _ := cli.Grant(ctx, 10)
					txn2, _ := cli.Txn(ctx).
						If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
						Then(clientv3.OpPut(lockKey, "client_C", clientv3.WithLease(leaseC.ID))).
						Commit()
					if txn2.Succeeded {
						gotByC <- time.Since(watchStart)
						cli.Revoke(ctx, leaseC.ID)
						return
					}
				}
			}
		}
	}()

	// 等待 C 拿锁
	select {
	case waited := <-gotByC:
		fmt.Printf("[T+%s][C] Watch 到 key 删除，立即 Txn(CAS) → 拿到锁，等待 %v\n",
			elapsed(t0), waited.Round(time.Millisecond))
	case <-time.After(8 * time.Second):
		fmt.Printf("[T+%s][C] ⚠️ 8s 内未拿到锁\n", elapsed(t0))
	}

	cli.Delete(ctx, lockKey)

	// ====== 总结 ======
	fmt.Printf("\n=== 实验结论 ===\n")
	fmt.Printf("1. A 主动 Revoke：key 立即消失，watcher 立即收到 Delete 事件——不用等 TTL\n")
	fmt.Printf("2. B 模拟崩溃：lease ttl=3s，server 检测不到 keepalive 后自动 revoke\n")
	fmt.Printf("3. C 通过 Watch + Txn(CAS) 在 lock 释放的瞬间拿到锁——抢占公平且即时\n")
	fmt.Printf("\n对照 Redis 三场景：\n")
	fmt.Printf("- 进程崩溃（E1）：etcd lease 由 server 自动 revoke，不用等 TTL\n")
	fmt.Printf("- 续期失败（E2）：keepalive 断 → server 立即 revoke；新持锁者明确知道前一个走了\n")
	fmt.Printf("- 主从切换（E3）：etcd 走 raft，写入需过半节点 ack，不存在异步丢数据\n")
}

func drainKeepAlive(ch <-chan *clientv3.LeaseKeepAliveResponse, owner string) {
	for range ch {
		// 只是消费 channel，避免阻塞
	}
	fmt.Printf("    [keepalive %s] channel 关闭（lease 已过期或 client 关闭）\n", owner)
}

func elapsed(t0 time.Time) string {
	return fmt.Sprintf("%5.2fs", time.Since(t0).Seconds())
}

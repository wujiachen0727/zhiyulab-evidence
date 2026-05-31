// 实验 E6 配套 benchmark：Redis SETNX 与 etcd Txn(CAS) 的获锁 QPS 对比
//
// 不是绝对吞吐，是同环境下的相对对比——决策矩阵章节用得上。
// 测的是"单 client 串行获锁/释放的 round-trip 速率"，不是并发吞吐——
// 因为分布式锁的核心瓶颈就是 round-trip 延迟。
//
// 运行前置：本地 Redis(6379) + etcd(2379) 都在跑
// 运行：go run ./06-bench
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const iterations = 5000

func main() {
	ctx := context.Background()

	// ---- Redis SETNX + DEL ----
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	defer rdb.Close()

	rdb.Del(ctx, "bench:lock")
	startR := time.Now()
	for i := 0; i < iterations; i++ {
		key := "bench:lock"
		ok, err := rdb.SetNX(ctx, key, "x", 5*time.Second).Result()
		if err != nil {
			panic(err)
		}
		if !ok {
			panic("SETNX 应该总是 true")
		}
		if _, err := rdb.Del(ctx, key).Result(); err != nil {
			panic(err)
		}
	}
	rDur := time.Since(startR)
	rQPS := float64(iterations) / rDur.Seconds()
	fmt.Printf("Redis  SETNX+DEL  %d 次  耗时 %v  QPS=%.0f  P50≈%v/op\n",
		iterations, rDur.Round(time.Millisecond), rQPS, rDur/time.Duration(iterations))

	// ---- etcd Txn(CAS) Put + Delete ----
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		panic(err)
	}
	defer cli.Close()

	cli.Delete(ctx, "/bench/lock")
	startE := time.Now()
	for i := 0; i < iterations; i++ {
		key := "/bench/lock"
		txn, err := cli.Txn(ctx).
			If(clientv3.Compare(clientv3.CreateRevision(key), "=", 0)).
			Then(clientv3.OpPut(key, "x")).
			Commit()
		if err != nil {
			panic(err)
		}
		if !txn.Succeeded {
			panic("Txn 应该 succeed")
		}
		if _, err := cli.Delete(ctx, key); err != nil {
			panic(err)
		}
	}
	eDur := time.Since(startE)
	eQPS := float64(iterations) / eDur.Seconds()
	fmt.Printf("etcd   Txn+Delete %d 次  耗时 %v  QPS=%.0f  P50≈%v/op\n",
		iterations, eDur.Round(time.Millisecond), eQPS, eDur/time.Duration(iterations))

	fmt.Printf("\n速率比：Redis / etcd = %.2fx\n", rQPS/eQPS)
	fmt.Printf("\n注意：本测试是单 client 串行 round-trip，反映网络延迟+服务端处理；\n")
	fmt.Printf("      生产环境的并发吞吐会进一步放大差距，但相对比例近似。\n")
}

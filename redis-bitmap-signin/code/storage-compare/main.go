// E6：Bitmap vs Set vs Hash 内存对比
// 实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	// 测试两种场景：连续 ID 和 稀疏 ID
	scenarios := []struct {
		name   string
		minID  int64
		maxID  int64
		step   int64 // step=1 连续, step=1000 稀疏
	}{
		{"连续 ID 1 万用户", 0, 10000, 1},
		{"连续 ID 10 万用户", 0, 100000, 1},
		{"连续 ID 100 万用户", 0, 1000000, 1},
		{"稀疏 ID 1 万用户（步长 1000）", 0, 10000000, 1000},
	}

	fmt.Println("=== E6：Bitmap vs Set vs Hash 内存对比 ===")
	fmt.Println("实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64")
	fmt.Println()

	for _, sc := range scenarios {
		fmt.Printf("--- %s ---\n", sc.name)
		count := (sc.maxID - sc.minID) / sc.step

		// Bitmap
		keyBm := fmt.Sprintf("sign:bitmap:%s", sc.name)
		for uid := sc.minID; uid < sc.maxID; uid += sc.step {
			_ = rdb.SetBit(ctx, keyBm, uid, 1).Err()
		}
		memBm, _ := rdb.MemoryUsage(ctx, keyBm).Result()

		// Set
		keySet := fmt.Sprintf("sign:set:%s", sc.name)
		members := make([]interface{}, 0, count)
		for uid := sc.minID; uid < sc.maxID; uid += sc.step {
			members = append(members, fmt.Sprintf("%d", uid))
		}
		_ = rdb.SAdd(ctx, keySet, members...).Err()
		memSet, _ := rdb.MemoryUsage(ctx, keySet).Result()

		// Hash
		keyHash := fmt.Sprintf("sign:hash:%s", sc.name)
		fieldValues := make([]interface{}, 0, count*2)
		for uid := sc.minID; uid < sc.maxID; uid += sc.step {
			fieldValues = append(fieldValues, fmt.Sprintf("%d", uid), "1")
		}
		_ = rdb.HSet(ctx, keyHash, fieldValues...).Err()
		memHash, _ := rdb.MemoryUsage(ctx, keyHash).Result()

		fmt.Printf("  Bitmap:  MEMORY = %-12d 字节 (%.2f MB)\n", memBm, float64(memBm)/1024/1024)
		fmt.Printf("  Set:     MEMORY = %-12d 字节 (%.2f MB)\n", memSet, float64(memSet)/1024/1024)
		fmt.Printf("  Hash:    MEMORY = %-12d 字节 (%.2f MB)\n", memHash, float64(memHash)/1024/1024)
		fmt.Printf("  Bitmap/Set = %.2fx  Bitmap/Hash = %.2fx\n", float64(memBm)/float64(memSet), float64(memBm)/float64(memHash))
		fmt.Println()
	}

	_ = os.Stdout.Sync()
}

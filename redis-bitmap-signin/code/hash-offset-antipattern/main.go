// E4：hash offset 反模式复现（v2）
// 核心假设：hash(userID) 作为 SETBIT offset 会导致内存爆炸
// 证伪实验：测试多组不同分布的字符串用户 ID，看 hash 后 offset 是否真的会爆炸
// 实测环境：Redis 8.8.0 / Go 1.26.4
package main

import (
	"context"
	"fmt"
	"hash/crc32"
	"hash/fnv"
	"os"

	"github.com/redis/go-redis/v9"
)

// 模拟腾讯云事故：用 hash(用户ID) 作为 SETBIT 的 offset
// 对比：同样数量的用户，连续 offset vs hash offset 的内存差异

func fnv1a64(s string) int64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return int64(h.Sum64() % (1 << 32)) // 对 2^32 取模，模拟工程常见做法
}

// crc32 是工程中常见的 hash 选择
func crc32hash(s string) int64 {
	return int64(crc32.ChecksumIEEE([]byte(s)))
}

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	testCases := []struct {
		name    string
		userIDs []string
	}{
		{"10 个用户", genUserIDs(10)},
		{"100 个用户", genUserIDs(100)},
		{"1000 个用户", genUserIDs(1000)},
		{"10000 个用户", genUserIDs(10000)},
	}

	fmt.Println("=== E4：hash offset 反模式复现 ===")
	fmt.Println("实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64")
	fmt.Println()
	fmt.Println("对比：相同用户数，连续 offset（0,1,2...）vs hash(offset) 的内存占用")
	fmt.Println()

	for _, tc := range testCases {
		fmt.Printf("--- %s ---\n", tc.name)

		// 方案 A：连续 offset（正确做法）
		keyA := fmt.Sprintf("sign:seq:%d", len(tc.userIDs))
		for i := range tc.userIDs {
			_ = rdb.SetBit(ctx, keyA, int64(i), 1).Err()
		}
		memA, _ := rdb.MemoryUsage(ctx, keyA).Result()
		strlenA, _ := rdb.StrLen(ctx, keyA).Result()

		// 方案 B：fnv1a64 hash offset（反模式）
		keyB := fmt.Sprintf("sign:fnv:%d", len(tc.userIDs))
		maxOffsetB := int64(0)
		for _, uid := range tc.userIDs {
			offset := fnv1a64(uid)
			if offset > maxOffsetB {
				maxOffsetB = offset
			}
			_ = rdb.SetBit(ctx, keyB, offset, 1).Err()
		}
		memB, _ := rdb.MemoryUsage(ctx, keyB).Result()
		strlenB, _ := rdb.StrLen(ctx, keyB).Result()

		// 方案 C：crc32 hash offset（反模式，另一种常见 hash）
		keyC := fmt.Sprintf("sign:crc:%d", len(tc.userIDs))
		maxOffsetC := int64(0)
		for _, uid := range tc.userIDs {
			offset := crc32hash(uid)
			if offset > maxOffsetC {
				maxOffsetC = offset
			}
			_ = rdb.SetBit(ctx, keyC, offset, 1).Err()
		}
		memC, _ := rdb.MemoryUsage(ctx, keyC).Result()
		strlenC, _ := rdb.StrLen(ctx, keyC).Result()

		fmt.Printf("  连续 offset（正确）:     MEMORY = %-12d  STRLEN = %-12d  (max offset=%d)\n", memA, strlenA, len(tc.userIDs)-1)
		fmt.Printf("  FNV1a hash（反模式）:    MEMORY = %-12d  STRLEN = %-12d  (max offset=%d)\n", memB, strlenB, maxOffsetB)
		fmt.Printf("  CRC32 hash（反模式）:    MEMORY = %-12d  STRLEN = %-12d  (max offset=%d)\n", memC, strlenC, maxOffsetC)
		fmt.Printf("  倍数差异 FNV/连续: %.1fx\n", float64(memB)/float64(memA))
		fmt.Printf("  倍数差异 CRC/连续: %.1fx\n", float64(memC)/float64(memA))
		fmt.Println()
	}

	_ = os.Stdout.Sync()
}

func genUserIDs(n int) []string {
	ids := make([]string, n)
	for i := 0; i < n; i++ {
		ids[i] = fmt.Sprintf("user_%05d", i+1)
	}
	return ids
}

// E2：BITCOUNT 字节偏移 vs 位偏移的实测对比
// 实测环境：Redis 8.8.0
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

	key := "sign:202606"

	// 设置第 0 位和第 8 位为 1
	// 期望：BITCOUNT 总数 = 2
	_ = rdb.SetBit(ctx, key, 0, 1).Err()
	_ = rdb.SetBit(ctx, key, 8, 1).Err()

	// STRLEN 查看字节数
	sl, _ := rdb.StrLen(ctx, key).Result()
	fmt.Println("STRLEN sign:202606 →", sl, "字节")
	fmt.Println("（第 0 位在第 0 字节，第 8 位在第 1 字节）")
	fmt.Println()

	// 错误用法：以为是位偏移
	// BITCOUNT key 0 10 → 统计第 0-10 字节（不是 0-10 位）
	// 因为只有 2 字节有数据，所以 0-10 字节 = 全部 = 2
	total, _ := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 10}).Result()
	fmt.Println("BITCOUNT sign:202606 0 10  →", total, "（以为是 0-10 位，实际是 0-10 字节）")

	// 正确用法：字节偏移
	// 0 0 → 统计第 0 字节 = 1（只有第 0 位被设置）
	b0, _ := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 0}).Result()
	fmt.Println("BITCOUNT sign:202606 0 0   →", b0, "（统计第 0 字节）")

	// 0 1 → 统计第 0-1 字节 = 2（第 0 位和第 8 位）
	b01, _ := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 1}).Result()
	fmt.Println("BITCOUNT sign:202606 0 1   →", b01, "（统计第 0-1 字节）")

	// 1 1 → 统计第 1 字节 = 1（只有第 8 位被设置）
	b11, _ := rdb.BitCount(ctx, key, &redis.BitCount{Start: 1, End: 1}).Result()
	fmt.Println("BITCOUNT sign:202606 1 1   →", b11, "（统计第 1 字节）")

	fmt.Println()
	fmt.Println("反直觉点：")
	fmt.Println("- BITCOUNT 的 start/end 是【字节偏移】不是【位偏移】")
	fmt.Println("- 想统计第 0-7 位，应该用 BITCOUNT key 0 0（第 0 字节）")
	fmt.Println("- 想统计第 8-15 位，应该用 BITCOUNT key 1 1（第 1 字节）")

	_ = os.Stdout.Sync()
}

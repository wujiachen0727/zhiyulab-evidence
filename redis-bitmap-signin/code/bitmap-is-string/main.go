// E1：Bitmap 底层是 String 的演示
// 实测环境：Redis 8.8.0 / darwin/arm64
// 运行：go run main.go
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
	_ = rdb.FlushDB(ctx).Err() // 清空 DB，保证干净环境

	key := "sign:demo"

	// 1. 设置第 7 位为 1
	if err := rdb.SetBit(ctx, key, 7, 1).Err(); err != nil {
		fmt.Println("SetBit error:", err)
		os.Exit(1)
	}

	// 2. 查看类型
	t, _ := rdb.Type(ctx, key).Result()
	fmt.Println("TYPE sign:demo        →", t)

	// 3. 查看编码
	enc, _ := rdb.ObjectEncoding(ctx, key).Result()
	fmt.Println("OBJECT ENCODING sign:demo →", enc)

	// 4. 查看长度
	sl, _ := rdb.StrLen(ctx, key).Result()
	fmt.Println("STRLEN sign:demo      →", sl, "字节")

	// 5. 再设置第 100 位
	_ = rdb.SetBit(ctx, key, 100, 1).Err()
	enc2, _ := rdb.ObjectEncoding(ctx, key).Result()
	sl2, _ := rdb.StrLen(ctx, key).Result()
	fmt.Println()
	fmt.Println("设置第 100 位后：")
	fmt.Println("OBJECT ENCODING sign:demo →", enc2)
	fmt.Println("STRLEN sign:demo      →", sl2, "字节（13 字节 = ceil(101/8)）")

	// 6. 用 GET 查看底层字节（证明是 String）
	raw, _ := rdb.Get(ctx, key).Result()
	fmt.Println()
	fmt.Println("GET sign:demo         →", []byte(raw))
	fmt.Println("（底层就是 String 的字节序列）")
}

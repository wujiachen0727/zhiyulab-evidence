// E7：签到场景完整实现
// 含：每日签到、查询某天是否签到、连续签到天数、本月签到次数
// 实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// SigninService 签到服务
// key 设计：sign:{uid}:{yyyyMM}，每月一个 Bitmap
type SigninService struct {
	rdb *redis.Client
}

// NewSigninService 创建签到服务
func NewSigninService(rdb *redis.Client) *SigninService {
	return &SigninService{rdb: rdb}
}

// Signin 用户签到（指定日期）
// uid: 用户 ID（必须是连续整数，不能是 hash）
// date: 签到日期
func (s *SigninService) Signin(ctx context.Context, uid int64, date time.Time) error {
	key := s.key(uid, date)
	day := int64(date.Day() - 1) // 第 1 天对应 offset 0
	return s.rdb.SetBit(ctx, key, day, 1).Err()
}

// IsSigned 查询某天是否签到
func (s *SigninService) IsSigned(ctx context.Context, uid int64, date time.Time) (bool, error) {
	key := s.key(uid, date)
	day := int64(date.Day() - 1)
	bit, err := s.rdb.GetBit(ctx, key, day).Result()
	if err != nil {
		return false, err
	}
	return bit == 1, nil
}

// MonthlyCount 本月签到次数
// 用 BITCOUNT 统计整个月 Bitmap 的置位位数
func (s *SigninService) MonthlyCount(ctx context.Context, uid int64, date time.Time) (int64, error) {
	key := s.key(uid, date)
	return s.rdb.BitCount(ctx, key, nil).Result()
}

// ContinuousDays 连续签到天数（从指定日期往前数）
// 用 BITFIELD 读取本月 Bitmap 的位段，逐位判断
// 注意：跨月查询需要调用方按月分段查询，本方法只处理单月内
func (s *SigninService) ContinuousDays(ctx context.Context, uid int64, date time.Time) (int64, error) {
	key := s.key(uid, date)
	day := int64(date.Day()) // 从第 day 位往前数（offset 0 = 第 1 天）

	// 用 BITFIELD GET 读取从 offset 0 开始的 day 位
	// BITFIELD key GET u{day} 0
	// 返回一个无符号整数
	// 注意：BITFIELD 位序是 offset 0 = 最高位（大端序）
	// 所以 offset (day-1) = 最低位（bit 0）
	res, err := s.rdb.BitField(ctx, key, "GET", fmt.Sprintf("u%d", day), 0).Result()
	if err != nil {
		return 0, err
	}
	if len(res) == 0 {
		return 0, nil
	}
	bits := uint64(res[0])

	// 从最低位（当天，offset=day-1）往前数连续的 1
	// 最低位 = 当天，次低位 = 昨天，依此类推
	count := int64(0)
	for i := uint(0); i < uint(day); i++ {
		if (bits>>i)&1 == 1 {
			count++
		} else {
			break // 遇到 0 就中断
		}
	}
	return count, nil
}

// key 生成签到 key
// sign:{uid}:{yyyyMM}
func (s *SigninService) key(uid int64, date time.Time) string {
	return fmt.Sprintf("sign:%d:%s", uid, date.Format("200601"))
}

func main() {
	rdb := redis.NewClient(&redis.Options{Addr: "127.0.0.1:6379"})
	ctx := context.Background()
	_ = rdb.FlushDB(ctx).Err()

	svc := NewSigninService(rdb)
	uid := int64(1001)

	// 模拟 2026-06 月的签到
	loc, _ := time.LoadLocation("Asia/Shanghai")
	// 假设今天是 2026-06-28，用户从 6-25 开始连续签到到 6-28
	signinDays := []int{25, 26, 27, 28}
	for _, day := range signinDays {
		date := time.Date(2026, 6, day, 0, 0, 0, 0, loc)
		if err := svc.Signin(ctx, uid, date); err != nil {
			fmt.Println("Signin error:", err)
			os.Exit(1)
		}
	}

	today := time.Date(2026, 6, 28, 0, 0, 0, 0, loc)

	// 1. 查询今天是否签到
	signed, _ := svc.IsSigned(ctx, uid, today)
	fmt.Printf("用户 %d 今天(%s)是否签到: %v\n", uid, today.Format("2006-01-02"), signed)

	// 2. 查询本月签到次数
	count, _ := svc.MonthlyCount(ctx, uid, today)
	fmt.Printf("用户 %d 本月签到次数: %d\n", uid, count)

	// 3. 查询连续签到天数
	continuous, _ := svc.ContinuousDays(ctx, uid, today)
	fmt.Printf("用户 %d 连续签到天数: %d\n", uid, continuous)

	// 4. 查询某天是否签到（6-26）
	day26 := time.Date(2026, 6, 26, 0, 0, 0, 0, loc)
	signed26, _ := svc.IsSigned(ctx, uid, day26)
	fmt.Printf("用户 %d 在 2026-06-26 是否签到: %v\n", uid, signed26)

	// 5. 查询没签到的天
	day24 := time.Date(2026, 6, 24, 0, 0, 0, 0, loc)
	signed24, _ := svc.IsSigned(ctx, uid, day24)
	fmt.Printf("用户 %d 在 2026-06-24 是否签到: %v\n", uid, signed24)

	// 6. 查看 key 的底层信息
	key := svc.key(uid, today)
	t, _ := rdb.Type(ctx, key).Result()
	strlen, _ := rdb.StrLen(ctx, key).Result()
	mem, _ := rdb.MemoryUsage(ctx, key).Result()
	fmt.Println()
	fmt.Printf("key: %s\n", key)
	fmt.Printf("TYPE: %s, STRLEN: %d 字节, MEMORY: %d 字节\n", t, strlen, mem)

	_ = os.Stdout.Sync()
}

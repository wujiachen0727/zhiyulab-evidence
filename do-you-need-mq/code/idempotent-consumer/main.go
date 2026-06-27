package main

// 幂等消费者示例：订单支付成功通知
// 展示"写对一个幂等消费者"需要考虑的 5 个边界条件

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type PaymentEvent struct {
	OrderID   string
	EventID   string // 消息唯一 ID（由生产者生成）
	Amount    int64
	CreatedAt time.Time
}

type IdempotentConsumer struct {
	redis *redis.Client
	// db    *sql.DB  // 实际项目中还需要数据库
}

func (c *IdempotentConsumer) HandlePayment(ctx context.Context, event PaymentEvent) error {
	// ===== 边界条件 1：消息 ID 为空 =====
	// 生产者可能忘记设置 EventID（尤其是老代码迁移时）
	if event.EventID == "" {
		return fmt.Errorf("event_id is empty, cannot guarantee idempotency")
	}

	// ===== 边界条件 2：Redis 去重窗口设多长？ =====
	// 设太短：消息重试间隔超过窗口 → 重复执行
	// 设太长：Redis 内存无限增长
	// 经验值：消息最大重试间隔 × 3（如重试最多 5 分钟，窗口设 15 分钟）
	deduplicationKey := fmt.Sprintf("idempotent:%s", event.EventID)
	deduplicationWindow := 15 * time.Minute

	// ===== 边界条件 3：SET NX 和业务逻辑不是原子的 =====
	// 如果 SET NX 成功但业务逻辑失败 → 消息被"吞掉"了
	// 正确做法：先执行业务逻辑，成功后再 SET（但这又有窗口期问题）
	// 折中：SET NX 先占位，业务失败则 DEL
	ok, err := c.redis.SetNX(ctx, deduplicationKey, "processing", deduplicationWindow).Result()
	if err != nil {
		// ===== 边界条件 4：Redis 不可用时怎么办？ =====
		// 选项 A：拒绝消费（消息堆积）← 保守但安全
		// 选项 B：放行消费（可能重复）← 激进但不堆积
		// 大多数团队选 B 然后忘了加监控告警...
		return fmt.Errorf("redis unavailable: %w", err)
	}

	if !ok {
		// 已处理过，跳过
		return nil
	}

	// 执行业务逻辑
	if err := c.processPayment(ctx, event); err != nil {
		// 业务失败，删除占位键，允许重试
		c.redis.Del(ctx, deduplicationKey)
		return err
	}

	// ===== 边界条件 5：业务成功但更新 Redis 状态失败 =====
	// 下次重试时 SET NX 会成功 → 业务重复执行
	// 真正的幂等性需要业务逻辑本身是幂等的（如 INSERT ... ON CONFLICT DO NOTHING）
	// Redis 去重只是"尽力而为"的第一道防线
	c.redis.Set(ctx, deduplicationKey, "done", deduplicationWindow)

	return nil
}

func (c *IdempotentConsumer) processPayment(ctx context.Context, event PaymentEvent) error {
	// 实际业务：更新订单状态、发送通知等
	fmt.Printf("Processing payment for order %s, amount: %d\n", event.OrderID, event.Amount)
	return nil
}

// 总结：一个"正确"的幂等消费者需要考虑：
// 1. 消息 ID 为空的防御
// 2. 去重窗口的合理设置
// 3. 占位 vs 业务逻辑的原子性
// 4. Redis 不可用时的降级策略
// 5. 业务成功但状态更新失败的兜底
//
// 代码行数：~60 行核心逻辑
// 额外依赖：Redis
// 额外约束：业务逻辑本身也要做幂等（数据库层面）

package main

// Transactional Outbox 模式 vs 直接发 MQ 的代码对比
// 展示：同样实现"订单创建后发通知"，两种方案的差异

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// ============================================================
// 方案 A：Transactional Outbox（~40 行核心逻辑，0 外部 MQ 依赖）
// ============================================================

type OutboxEntry struct {
	ID        int64
	EventType string
	Payload   json.RawMessage
	CreatedAt time.Time
	Sent      bool
}

func CreateOrderWithOutbox(ctx context.Context, db *sql.DB, order Order) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 写入订单
	_, err = tx.ExecContext(ctx,
		"INSERT INTO orders (id, user_id, amount, status) VALUES (?, ?, ?, ?)",
		order.ID, order.UserID, order.Amount, "created")
	if err != nil {
		return err
	}

	// 2. 同一个事务写入 outbox（原子性保证：要么都成功，要么都失败）
	payload, _ := json.Marshal(map[string]interface{}{
		"order_id": order.ID,
		"user_id":  order.UserID,
		"amount":   order.Amount,
	})
	_, err = tx.ExecContext(ctx,
		"INSERT INTO outbox (event_type, payload, created_at, sent) VALUES (?, ?, ?, ?)",
		"order.created", payload, time.Now(), false)
	if err != nil {
		return err
	}

	return tx.Commit()
	// 就这么多。不需要 MQ 连接、不需要 MQ 集群、不需要处理 MQ 发送失败。
	// 一个独立的 goroutine 定期轮询 outbox 表，发送未发送的消息。
}

// Outbox Processor（独立运行，轮询 outbox 表）
func ProcessOutbox(ctx context.Context, db *sql.DB, notify func(payload []byte) error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			rows, _ := db.QueryContext(ctx,
				"SELECT id, payload FROM outbox WHERE sent = false ORDER BY created_at LIMIT 100")
			for rows.Next() {
				var id int64
				var payload []byte
				rows.Scan(&id, &payload)
				if err := notify(payload); err == nil {
					db.ExecContext(ctx, "UPDATE outbox SET sent = true WHERE id = ?", id)
				}
			}
			rows.Close()
		}
	}
}

// ============================================================
// 方案 B：直接发 RabbitMQ（~25 行核心逻辑，但需要 MQ 集群）
// ============================================================

/*
import amqp "github.com/rabbitmq/amqp091-go"

func CreateOrderWithMQ(ctx context.Context, db *sql.DB, ch *amqp.Channel, order Order) error {
    // 1. 写入订单
    _, err := db.ExecContext(ctx,
        "INSERT INTO orders (id, user_id, amount, status) VALUES (?, ?, ?, ?)",
        order.ID, order.UserID, order.Amount, "created")
    if err != nil {
        return err
    }

    // 2. 发送 MQ 消息（⚠️ 双写问题：DB 成功但 MQ 失败怎么办？）
    payload, _ := json.Marshal(map[string]interface{}{
        "order_id": order.ID,
        "user_id":  order.UserID,
        "amount":   order.Amount,
    })
    return ch.PublishWithContext(ctx,
        "orders",         // exchange
        "order.created",  // routing key
        false, false,
        amqp.Publishing{
            ContentType: "application/json",
            Body:        payload,
        })
    // ⚠️ 如果这里失败了：
    // - 订单已入库但通知没发出
    // - 需要补偿机制（定时任务扫描未通知的订单）
    // - 这个补偿机制...和 outbox 本质上是一回事
}
*/

// ============================================================
// 对比总结
// ============================================================
//
// |          | Outbox 模式         | 直接发 MQ           |
// |----------|--------------------|--------------------|
// | 核心代码 | ~40 行              | ~25 行              |
// | 外部依赖 | 无（用现有数据库）    | MQ 集群 + 客户端库   |
// | 一致性   | 强（同事务）         | 弱（双写问题）       |
// | 运维成本 | 低（数据库已有）     | 高（MQ 集群高可用）  |
// | 吞吐量   | 受 DB 限制           | 高（MQ 专门优化）    |
// | 适用场景 | 日均消息 < 10 万条   | 日均消息 > 10 万条   |

type Order struct {
	ID     string
	UserID string
	Amount int64
}

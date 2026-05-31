# E1：进程崩溃锁不释放

## 复现什么
单 Redis + 单进程模拟「客户端 A 拿锁后崩溃，没有调用 DEL/UNLINK」的场景。
观察客户端 B 多久才能拿到这把锁——答案是必须等到 Redis 自动让 TTL 过期。

## 前置环境
- 本地 6379 端口有 Redis：
  ```bash
  docker run -d --name dl-redis -p 6379:6379 redis:7-alpine
  ```
- Go 1.21+

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code
go run ./01-process-crash
```

## 关键参数（main.go 顶部常量）
- `ttl = 10 * time.Second`：A 设的锁 TTL
- `pollGap = 200 * time.Millisecond`：B 重试间隔
- `maxWait = 30 * time.Second`：B 最长等待

## 预期输出（关键行）
- A 在 `T+0s` 拿到锁
- B 在 `T+~10s` 拿到锁（接近 TTL）
- 总等待时间 ≈ ttl

## 关键证据
**这个等待时间是 SETNX 锁的"成本下限"——只要 A 异常退出且没人手动清锁，B 就必须等满 TTL。**
缩小 TTL 可以缩短等待，但 TTL 太短会让正常持锁者被误踢；TTL 太长会让崩溃恢复变慢。这是 SETNX 锁的固有张力。

## 没有验证什么
- Redis 7 是否有 client tracking 之类的机制能自动清锁？
  → 实验观察到 B 等了 ~10s，说明默认配置下没有这种自动清锁，必须等 TTL。如果你的环境配了 client-side caching 也不解决这个问题。

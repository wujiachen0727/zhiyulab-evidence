# E3 实验：不同淘汰策略的行为对比

## 目的
验证 Redis 不同 maxmemory-policy 的淘汰行为差异

## 环境
- Redis 7.0 (Docker `redis:7-alpine`)
- redis-cli 8.8.0
- macOS ARM64 (darwin/arm64)

## 实验方法
分别设置 noeviction / allkeys-lru / volatile-ttl / volatile-lru 策略，观察不同策略下的淘汰行为

## 关键发现

### 实验 1：noeviction
- 第 79 个 key 写入时触发 OOM 报错——验证 noeviction 真的不淘汰，满了就报错

### 实验 2：allkeys-lru
- 设置 128KB maxmemory 时，所有 50 个 key（每个 ~2KB）都被淘汰
- 原因：128KB 太小，淘汰器需要释放足够空间

### 实验 3：volatile-ttl
- 所有 3 个带不同 TTL 的 key 都被淘汰（64KB 限制太紧）

### 实验 4：volatile-lru vs allkeys-lru 的关键区别 ✅
- **volatile-lru**：有 TTL 的 key 全部被淘汰（0/10），**无 TTL 的 key 全部保留（10/10）**
- 这验证了 volatile-* 策略只会在设置了 TTL 的 key 中做淘汰选择
- 无 TTL 的 key 在 volatile-* 策略下永远不会被淘汰

## 数据
详见 `result.txt`

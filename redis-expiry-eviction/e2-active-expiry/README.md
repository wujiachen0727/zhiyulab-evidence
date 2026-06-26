# E2 实验：定期删除（Active Expiry Cycle）

## 目的
验证 Redis 后台定期删除机制——activeExpireCycle 每 100ms 运行一次，每次扫描最多 20 个 key。

## 环境
- Redis 7.0 (Docker `redis:7-alpine`)
- redis-cli 8.8.0
- macOS ARM64 (darwin/arm64)

## 实验方法
1. 批量写入 1000/10000/50000 个带相同 TTL 的 key
2. 等待超过过期时间
3. 用 INFO stats 的 expired_keys 指标监控删除进度

## 关键发现
1. **1000 个 key**：8 秒内全部清理（含 10 个惰性删除 + 990 个定期删除）
2. **10000 个 key**：10 秒内全部清理
3. **50000 个 key**（极端情况）：15 秒内全部清理
4. **定期删除效率很高**：即使 5 万个 key 同时过期，Redis 也能在秒级完成清理
5. 定期删除的自适应机制在起作用——过期比例高时会持续扫描

## 数据
详见 `result.txt`

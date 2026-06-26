# E1 实验：惰性删除（Lazy Expiry）触发时机

## 目的
验证 Redis 惰性删除的工作机制——设置了 TTL 的 key 过期后，不会立即被删除，而是在被访问时才触发惰性删除。

## 环境
- Redis 7.0 (Docker `redis:7-alpine`)
- redis-cli 8.8.0
- macOS ARM64 (darwin/arm64)

## 实验方法
1. 设置带 TTL 的 key
2. 等待超过过期时间
3. 用 GET 访问触发惰性删除
4. 用 INFO memory 观察内存变化

## 关键发现
1. **EXISTS 也会触发惰性删除**：不仅是 GET，检查 key 是否存在的命令（EXISTS）同样会触发 expireIfNeeded 检查
2. **过期 key 占用内存直到被访问**：10KB 的数据过期后，内存并未立即释放（used_memory_dataset 在过期后仍为 ~235KB），直到访问后才被惰性删除
3. **惰性删除是"按需"的**：key 过期后，如果一直不被访问，它会一直占用内存

## 数据
详见 `result.txt`

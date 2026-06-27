# 三实例共享 Redis 限流键的取舍三角

> 场景模拟（E3.2）| 基于工程常识的合理推演

## 场景

3 个网关实例（GW-1, GW-2, GW-3）共享同一个 Redis 限流键，总额度 100 QPS。

## 三种实现方案

### 方案 A：各自维护本地令牌桶

```
GW-1: 本地桶 33 QPS
GW-2: 本地桶 33 QPS
GW-3: 本地桶 34 QPS
```

**优点**：零网络开销、零延迟
**缺点**：
- 流量不均时（如 GW-1 承载 80% 流量）→ GW-1 打满 33 拒绝，GW-2/3 空闲
- 实际总通过量在 34~100 之间浮动，无法保证精确 100
- 实例扩缩容需要重新计算分片

**适用场景**：精度要求不高、流量分布较均匀

---

### 方案 B：中心化 Redis WATCH + MULTI

```
1. WATCH token_key
2. GET token_key → current_tokens
3. 本地计算 new_tokens
4. MULTI
5.   SET token_key new_tokens
6. EXEC（如果 key 被修改过则失败）
7. 失败则重试
```

**优点**：不需要 Lua 脚本，Redis 原生支持
**缺点**：
- 高并发下 WATCH 冲突率极高 → 重试风暴
- 3 实例 × 100 QPS = 300 请求/秒竞争同一个 key
- 估算冲突率：假设 WATCH→EXEC 耗时 1ms，300 QPS 下同一毫秒窗口平均 0.3 个并发
- 当 QPS 升到 3000 时，同毫秒并发约 3 → 冲突率急剧上升
- P99 延迟可能因重试飙升到 10-50ms

**适用场景**：低 QPS（< 100 QPS）、对延迟不敏感

---

### 方案 C：Lua 原子脚本

```lua
-- KEYS[1] = 令牌桶 key
-- ARGV[1] = rate, ARGV[2] = capacity, ARGV[3] = now
local tokens = tonumber(redis.call('GET', KEYS[1]) or capacity)
local last = tonumber(redis.call('GET', KEYS[1]..':ts') or now)
local elapsed = now - last
tokens = math.min(capacity, tokens + elapsed * rate)
if tokens >= 1 then
  tokens = tokens - 1
  redis.call('SET', KEYS[1], tokens)
  redis.call('SET', KEYS[1]..':ts', now)
  return 1
end
return 0
```

**优点**：原子性保证（Redis 单线程执行 Lua）、无重试
**缺点**：
- Redis 变成单点瓶颈（所有限流判定都经过它）
- Lua 脚本阻塞 Redis 主线程（虽然单次极快，但高并发下累积）
- Redis 故障 = 限流全部失效（需要降级策略）
- 跨数据中心的 RTT 叠加

**适用场景**：大多数生产环境的首选（正确性 > 性能）

---

## 取舍三角

```
        精确性
       /      \
      /        \
   方案C ---- 方案B
    /              \
延迟              运维简单
    \              /
     \            /
      方案A ----
```

| 维度 | 方案A（本地） | 方案B（WATCH） | 方案C（Lua） |
|------|:-----------:|:-------------:|:-----------:|
| 精确性 | ★★☆ | ★★★ | ★★★ |
| 延迟 | ★★★ | ★☆☆ | ★★☆ |
| 运维复杂度 | ★★★ | ★★☆ | ★★☆ |
| 高并发表现 | ★★★ | ★☆☆ | ★★☆ |
| 故障影响 | 限于单实例 | Redis 单点 | Redis 单点 |

## 决策建议

- **QPS < 100，精度要求不高** → 方案 A（本地桶 + 静态分片）
- **QPS < 1000，需要精确** → 方案 C（Lua 脚本，最通用）
- **QPS > 10000** → 方案 A + 定期同步（混合）或分层限流
- **方案 B 几乎不推荐** → 高并发下重试风暴的代价太高，不如直接用 Lua

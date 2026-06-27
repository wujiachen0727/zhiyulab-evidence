# E4: 缓存模块 TDD vs Go 惯用 对照实验

[场景模拟] 2026-04-14

## 场景设定

需求：实现一个简单的内存缓存模块，支持 Get/Set/Delete 操作，有过期时间。

对比两条开发路径：
- **TDD 路线**：先写测试（红）→ 最小实现（绿）→ 重构
- **Go 惯用路线**：先写实现 → 写测试验证 → 迭代

这是模拟场景，用于展示两种开发路径的体验差异。

## TDD 路线

```
第1步：写测试
  func TestCache_Get_Miss(t *testing.T) {
      c := NewCache()
      _, ok := c.Get("notexist")
      if ok {
          t.Error("expected cache miss")
      }
  }
  → 红色：NewCache 未定义

第2步：最小实现
  type Cache struct{}
  func NewCache() *Cache { return &Cache{} }
  func (c *Cache) Get(key string) (string, bool) { return "", false }
  → 绿色

第3步：写下一个测试
  func TestCache_SetAndGet(t *testing.T) {
      c := NewCache()
      c.Set("hello", "world", 0)
      val, ok := c.Get("hello")
      if !ok || val != "world" {
          t.Error("expected cache hit")
      }
  }
  → 红色：Set 未定义，Get 总是返回 miss

第4步：扩展实现
  type Cache struct { m map[string]entry }
  type entry struct { val string; exp time.Time }
  func (c *Cache) Set(key, val string, ttl time.Duration) { ... }
  func (c *Cache) Get(key string) (string, bool) { ... check expiry ... }
  → 绿色

第5步：重构
  → 提取 checkExpiry 方法
  → 添加 Delete

循环继续...最终约 8 个测试 + 1 个 Benchmark
```

**TDD 路线体验**：
- 每一步都有明确的"红→绿"信号
- 测试先于实现，接口设计由测试驱动
- 但：if+Errorf 的断言写法增加了每次"红→绿"的摩擦
- 总共需要 8 次红绿切换

## Go 惯用路线

```
第1步：写实现
  type Cache struct { mu sync.RWMutex; m map[string]entry }
  type entry struct { val string; exp time.Time }
  func NewCache() *Cache { return &Cache{m: make(map[string]entry)} }
  func (c *Cache) Set(key, val string, ttl time.Duration) {
      c.mu.Lock()
      defer c.mu.Unlock()
      c.m[key] = entry{val: val, exp: time.Now().Add(ttl)}
  }
  func (c *Cache) Get(key string) (string, bool) {
      c.mu.RLock()
      defer c.mu.RUnlock()
      e, ok := c.m[key]
      if !ok { return "", false }
      if !e.exp.IsZero() && time.Now().After(e.exp) {
          return "", false
      }
      return e.val, true
  }
  func (c *Cache) Delete(key string) { ... }
  → 直接得到一个可用的实现

第2步：写测试验证
  func TestCache_Operations(t *testing.T) {
      c := NewCache()
      // 测试 miss
      if _, ok := c.Get("notexist"); ok { t.Error("miss") }
      // 测试 set+get
      c.Set("hello", "world", time.Minute)
      if val, ok := c.Get("hello"); !ok || val != "world" { t.Error("hit") }
      // 测试过期
      c.Set("short", "lived", time.Nanosecond)
      time.Sleep(time.Millisecond)
      if _, ok := c.Get("short"); ok { t.Error("expired") }
      // 测试 delete
      c.Set("del", "me", 0)
      c.Delete("del")
      if _, ok := c.Get("del"); ok { t.Error("deleted") }
  }

第3步：加 Benchmark
  func BenchmarkCache_Get(b *testing.B) {
      c := NewCache()
      c.Set("key", "val", 0)
      b.ResetTimer()
      for range b.N {
          c.Get("key")
      }
  }
```

**Go 惯用路线体验**：
- 先有一个完整的心理模型，然后一次性实现
- 测试是"验证"而非"驱动"
- 可以在实现阶段做更多设计决策（比如直接加了 RWMutex）
- Benchmark 自然地从实现中生长出来——因为实现时就会考虑性能
- 总共 1 个测试函数 + 1 个 Benchmark

## 体验差异总结

| 维度 | TDD 路线 | Go 惯用路线 |
|------|---------|------------|
| 开发节奏 | 8 次红绿切换 | 1 次实现 + 1 次验证 |
| 接口设计 | 测试驱动 | 经验/设计驱动 |
| 并发考量 | 后期补 RWMutex | 一开始就加 |
| Benchmark | 后期补 | 实现时自然生长 |
| 代码量 | 更多测试代码，更小步迭代 | 更少测试函数，更大步推进 |
| 心理模型 | 逐步构建 | 先完整后验证 |

**核心差异**：TDD 的价值在于"用测试做设计探索"——当你不确定接口长什么样时，先写测试帮你理清思路。Go 惯用路线的价值在于"先想清楚再动手"——当你已经有清晰的设计时，先写测试是给自己加摩擦。

Go 的 testing 包设计隐含了后者的偏好：没有快速断言→红绿循环有摩擦→不是每次循环都很快→不如少循环几次，每次多想一点。

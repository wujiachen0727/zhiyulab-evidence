# E2：BITCOUNT 字节偏移 vs 位偏移的实测对比

## 测试方法

构造一个 Bitmap，设置第 0 位和第 8 位为 1，然后分别用不同 start/end 调用 BITCOUNT，对比结果。

## 实测环境

Redis 8.8.0 / Go 1.26.4 / darwin/arm64

## 数据

```
SETBIT sign:202606 0 1
SETBIT sign:202606 8 1

STRLEN sign:202606 → 2 字节（第 0 位在第 0 字节，第 8 位在第 1 字节）

BITCOUNT sign:202606 0 10  → 2 （"0-10"是字节偏移，不是位偏移）
BITCOUNT sign:202606 0 0   → 1 （统计第 0 字节）
BITCOUNT sign:202606 0 1   → 2 （统计第 0-1 字节）
BITCOUNT sign:202606 1 1   → 1 （统计第 1 字节）
```

## 关键发现

1. BITCOUNT 的 start/end 参数是**字节偏移**，不是位偏移
2. 反直觉点：想统计第 0-7 位，应该用 `BITCOUNT key 0 0`（第 0 字节），不是 `BITCOUNT key 0 7`
3. 想统计第 8-15 位，应该用 `BITCOUNT key 1 1`（第 1 字节）

## 常见误用

```go
// 错误：以为是位偏移
bitcount := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 7})  // 实际统计第 0-7 字节

// 正确：字节偏移
bitcount := rdb.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 0})  // 统计第 0 字节（第 0-7 位）
```

## 结论

BITCOUNT 的 start/end 是字节偏移是反直觉设计——大多数开发者第一次用会以为是位偏移。这个坑在 Redis 官方文档里有说明，但中文教程很少提及。

## 引用依据

Redis 官方文档：BITCOUNT 的 start 和 end 参数是 byte offsets，不是 bit offsets。
（https://redis.io/commands/bitcount/）

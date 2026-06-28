# E3：SETBIT offset 与内存分配的量化关系

## 测试方法

用不同 offset 设置位，用 MEMORY USAGE 命令查看 Redis 分配的内存。

## 实测环境

Redis 8.8.0 / darwin/arm64

## 数据

| SETBIT offset | STRLEN（字节）| MEMORY USAGE（字节）| 等价大小 |
|:-------------:|:-------------:|:------------------:|:--------:|
| 7 | 1 | 35 | - |
| 1,000 | 126 | 208 | - |
| 10,000 | 1,251 | 1,328 | ~1.3 KB |
| 1,000,000 | 125,001 | 131,120 | ~128 KB |
| 10,000,000 | 1,250,001 | 1,261,616 | ~1.2 MB |
| 100,000,000 | 12,500,001 | 12,501,040 | ~12 MB |

## 关键发现

1. **STRLEN = ceil((offset + 1) / 8)**：SETBIT offset=7 → 1 字节，offset=100 → 13 字节，offset=10^8 → 12.5MB
2. **MEMORY USAGE 略大于 STRLEN**：多出的是 Redis 内部元数据（redisObject 头、SDS 头等，约 16-64 字节）
3. **offset 决定内存**：即使只设置了 1 个位，offset 多大，内存就分配多大
4. **线性关系**：offset 每增加 8，STRLEN 增加 1 字节

## 结论

SETBIT 的 offset 直接决定内存分配。这是 Bitmap 的核心特性，也是所有踩坑的根源：
- 连续 offset → 内存紧凑（10 个用户只要 2 字节）
- 稀疏 offset → 内存爆炸（10 个用户但 max offset = 10^8 → 12MB）
- hash offset → 最坏情况（hash 值分散在 0-2^32，max offset 接近 2^32 → 512MB）

## 内存计算公式

```
STRLEN = ceil((max_offset + 1) / 8) 字节
MEMORY USAGE ≈ STRLEN + 64（元数据开销）
```

## 数据使用说明

- 正文引用时标注 `[实测 Redis 8.8.0]`
- 这个数据是 E4（hash offset 反模式）和 E6（稀疏 ID 场景）的原理基础

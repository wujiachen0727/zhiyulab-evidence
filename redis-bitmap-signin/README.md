# Evidence 总索引

> 文章：Redis Bitmap 签到实现：从命令到字节级原理
> 实测环境：Redis 8.8.0 / Go 1.26.4 / darwin/arm64 (M1)

## 论据清单

| ID | 类型 | 描述 | 产出路径 | 优先级 |
|----|------|------|---------|:------:|
| E1 | 实验验证 | Bitmap 底层是 String 的演示 | `code/bitmap-is-string/` + `output/bitmap-is-string/` + `data/bitmap-is-string.md` | P1 |
| E2 | 实验验证 | BITCOUNT 字节偏移 vs 位偏移对比 | `code/bitcount-byte-vs-bit/` + `data/bitcount-byte-vs-bit.md` | P1 |
| E3 | 数据实测 | SETBIT offset 与内存分配量化关系 | `data/offset-memory-relationship.md` | P1 |
| E4 | 实验验证 | hash offset 反模式复现（核心差异化）| `code/hash-offset-antipattern/` + `output/hash-offset-antipattern/` | P1 |
| E5 | 数据实测 | BITCOUNT 大 key 耗时实测 | `code/bitcount-latency/` + `data/bitcount-latency.md` | P2 |
| E6 | 数据实测 | Bitmap vs Set vs Hash 内存对比 | `code/storage-compare/` + `data/bitmap-vs-set-vs-hash.md` | P2 |
| E7 | 实验验证 | 签到场景完整实现代码（核心交付物）| `code/signin-implementation/` + `output/signin-implementation/` | P1 |

## 证伪实验结果

| ID | 核心假设 | 证伪结果 | 结论 |
|----|---------|---------|------|
| E4 | hash offset 会导致内存爆炸 | 假设成立 | hash 函数设计目标是均匀分布，必然产生大 offset |
| E5 | BITCOUNT 大 key 会阻塞 | **部分证伪** | 512MB Bitmap P99=8.83ms（本地），不严格算阻塞。论点需调整为"性能关注点" |
| E6 | Bitmap 比 Set/Hash 省内存 | **部分证伪** | 仅连续 ID 场景成立，稀疏 ID 场景 Bitmap 反而更费 |

## 自造比例

- 独立论据：7 项（E1-E7）
- 表达手法：2 项（M1-M2，不计入）
- 外部引用：3 处（Redis 官方文档 × 2 + 腾讯云事故）
- **自造比例：7/7 = 100%**

## 外部引用清单

| # | 引用内容 | 来源 | 使用位置 |
|---|---------|------|---------|
| 1 | Bitmaps not an actual data type | https://redis.io/docs/latest/develop/data-types/bitmaps/ | E1 佐证 |
| 2 | 腾讯云 60GB 内存事故 | https://cloud.tencent.com/developer/article/2421953 | E4 佐证 |
| 3 | BITCOUNT 复杂度 O(N) | https://redis.io/commands/bitcount/ | E5 佐证 |

## 重要发现（用于正文）

1. **E4：10 个用户 + hash offset = 300-470MB 内存**（FNV1a/CRC32 测试），是连续 offset 的 600-900 万倍
2. **E5：512MB Bitmap 的 BITCOUNT P99=8.83ms**（本地），论点需从"阻塞"调整为"性能关注点"
3. **E6：稀疏 ID 场景 Bitmap 反而比 Set/Hash 大 2.8 倍**，证伪"Bitmap 永远省内存"
4. **E2：BITCOUNT start/end 是字节偏移**，反直觉设计，中文教程很少提及
5. **E7：连续签到天数用 BITFIELD GET 一次读取整月**，避免 N 次 GETBIT 网络往返

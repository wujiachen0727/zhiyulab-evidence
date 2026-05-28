# Slice 扩容实验摘要

> 数据标注：除特别说明外，以下均为 `[实测 Go 1.17.13 / Go 1.26.2 darwin/arm64]`。

## E1：Go 1.17 非单调增长复现

- Go 1.17 扫描 oldcap=900..1400，append 1 个 byte 后，发现下降点 1 个。
- 关键下降：oldcap 1023 → newcap 2048；oldcap 1024 → newcap 1280。也就是 oldcap 增加 1，结果 newcap 从 2048 掉到 1280。
- Go 1.26.2 同范围下降点 0 个。

### 1024 附近窗口

| oldcap | Go 1.17 newcap | Go 1.26 newcap | 说明 |
|---:|---:|---:|---|
| 1018 | 2048 | 1536 |  |
| 1019 | 2048 | 1536 |  |
| 1020 | 2048 | 1536 |  |
| 1021 | 2048 | 1536 |  |
| 1022 | 2048 | 1536 |  |
| 1023 | 2048 | 1536 | 旧策略 1024 以下仍翻倍到 2048 |
| 1024 | 1280 | 1536 | 旧策略跨过阈值后掉到 1280；新策略保持 1536 |
| 1025 | 1408 | 1536 |  |
| 1026 | 1408 | 1536 |  |
| 1027 | 1408 | 1536 |  |
| 1028 | 1408 | 1536 |  |
| 1029 | 1408 | 1536 |  |
| 1030 | 1408 | 1536 |  |

## E2：新旧容量序列对比

| oldcap | Go 1.17 append 后 cap | Go 1.26 append 后 cap | 差异 |
|---:|---:|---:|---:|
| 900 | 2048 | 1408 | -640 |
| 960 | 2048 | 1408 | -640 |
| 1000 | 2048 | 1536 | -512 |
| 1023 | 2048 | 1536 | -512 |
| 1024 | 1280 | 1536 | 256 |
| 1025 | 1408 | 1536 | 128 |
| 1100 | 1408 | 1792 | 384 |
| 1200 | 1536 | 1792 | 256 |
| 1300 | 1792 | 2048 | 256 |
| 1400 | 1792 | 2048 | 256 |

## E3：benchmark 量化

> 解读规则：先看 allocs/op 与 B/op，再看 ns/op。

| benchmark | Go 1.17 ns/op | Go 1.26 ns/op | Go 1.17 B/op | Go 1.26 B/op | Go 1.17 allocs/op | Go 1.26 allocs/op |
|---|---:|---:|---:|---:|---:|---:|
| BenchmarkAppendByteGrowth_From1024To4096 | 1857 | 1514 | 12544 | 11008 | 5 | 4 |
| BenchmarkAppendByteGrowth_NoPrealloc_4K | 2454 | 2160 | 14584 | 12536 | 13 | 12 |
| BenchmarkAppendByteGrowth_Prealloc256_4K | 2397 | 1972 | 14080 | 12032 | 7 | 6 |

### benchmark 关键结论

- BenchmarkAppendByteGrowth_From1024To4096: B/op 下降约 12.2%，allocs/op 少 1 次。
- BenchmarkAppendByteGrowth_NoPrealloc_4K: B/op 下降约 14.0%，allocs/op 少 1 次。
- BenchmarkAppendByteGrowth_Prealloc256_4K: B/op 下降约 14.5%，allocs/op 少 1 次。

## E4：size class / roundupsize 根因推导

前提：`[]byte` 的元素大小为 1，因此 cap 基本等于申请字节数；实际 cap 会被 runtime 的 `roundupsize` 按 size class 向上取整。

三步推导：

1. Go 1.17 旧策略在 `oldcap < 1024` 时直接翻倍，所以 oldcap=1023 时理想 cap=2046，roundup 后得到 2048。
2. 但 oldcap=1024 一跨过阈值，旧策略进入 1.25x 分支，理想 cap=1280，roundup 后仍是 1280。
3. 因此 oldcap 从 1023 增加到 1024，append 后的新 cap 反而从 2048 掉到 1280。问题不是单独的 roundupsize，也不是单独的 1.25x，而是阈值硬切换 + size class 对齐共同制造了非单调。

Go 1.26.2 中，oldcap=1023 和 oldcap=1024 append 后都得到 1536，说明这个断崖被平滑策略消掉了。

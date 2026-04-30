# Evidence 总索引

## E1: sync.Map vs bigcache vs ristretto 性能对比

### 实验环境
- Go 1.26.2 darwin/arm64 (Apple M4 Pro)
- 测试参数：10万/100万 key，64B value，8 核并行

### 产出文件
- `code/cache-benchmark/bench_test.go` — ns/op benchmark
- `code/cache-benchmark/gc_pressure.go` — GC 压力对比
- `output/benchmark-results.txt` — benchmark 原始输出

### 核心发现

**ns/op 对比（10万 key，8 并发）：**

| 实现 | 读90写10 | 读70写30 | 读50写50 |
|------|:-------:|:-------:|:-------:|
| sync.Map | 33.6 ns | 50.5 ns | 66.9 ns |
| bigcache | 45.3 ns | 54.1 ns | 89.5 ns |
| ristretto | 116.8 ns | 219.0 ns | 335.2 ns |

**GC 压力对比（100万 key）：**

| 指标 | sync.Map | bigcache |
|------|:-------:|:-------:|
| 填充耗时 | 288ms | 177ms |
| HeapAlloc | 207 MB | 368 MB |
| HeapObjects | 4,842,250 | 499,432 |
| GC 暂停（均值）| ~25µs | ~39µs |

**结论**：sync.Map 单操作更快（纯查询低至 33ns），但对象数是 bigcache 的 10 倍。在大 key 量下 GC 频率更高、CPU 开销更大。拐点信号不是"单次操作慢了"，而是"GC CPU 占比明显上升"。

## E2: 多实例一致性问题（逻辑推演）

产出：直接融入正文。基于工程常识的推演——多实例部署后，用户 A 在实例 1 更新缓存，用户 B 在实例 2 读到旧值。

## E3: 本地缓存 vs Redis 延迟（量级分析）

产出：直接融入正文。已知量级：本地缓存 33-90ns（E1 实测），Redis 通常 100-500µs（网络往返），差距 1000-10000x。精确差距依赖网络配置，但量级差异是确定的。

## E4: 多级缓存一致性代价（逻辑推演）

产出：直接融入正文。分析 L1/L2 一致性的三种方案及其复杂度-收益权衡。

## E5: singleflight 防 stampede（实验验证 - 降级为推演）

产出：直接融入正文。singleflight 的效果在 Go 官方文档和社区已有充分验证，不需要重复实测。文中用代码片段说明用法即可。

## 自造度统计

- 独立论据 5 项：E1（实测）、E2（推演）、E3（量级分析）、E4（推演）、E5（降级为代码说明）
- 外部引用 1 处：sync.Map 官方文档设计意图
- 自造度：5/6 ≈ 83%

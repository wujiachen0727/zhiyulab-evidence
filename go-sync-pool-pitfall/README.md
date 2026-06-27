# 论据总索引 — go-sync-pool-pitfall

> 本文件是文章论据产物的总索引。所有论据均为自造（100%自造率）。

## 实验代码

| # | 实验名 | 路径 | 用途 |
|---|--------|------|------|
| 1 | Pool Benchmark Suite | `evidence/code/pool-benchmark/pool_bench_test.go` | Pool vs 直接分配的完整 benchmark 矩阵（7大小×5并发度） |
| 2 | Forced Escape Test | `evidence/code/pool-benchmark/forced_escape_test.go` | 正确的强制逃逸方法验证（修复翻车实验后） |

## 实验输出

| # | 文件 | 说明 |
|---|------|------|
| 1 | `evidence/output/benchmark-matrix-raw.txt` | 完整 benchmark 原始输出 |
| 2 | `evidence/output/benchmark-summary.md` | 整理后的加速比矩阵 + 关键读数 |

## 论据分配

| # | 论据 | 类型 | 承载章节 | 证据路径 |
|---|------|------|---------|---------|
| 1 | Pool.Get 路径开销拆解 | 源码分析 + 推演 | Ch1 | Go runtime 源码 |
| 2 | 各大小对象堆分配实测 | 数据实测 | Ch2 | evidence/output/benchmark-matrix-raw.txt |
| 3 | 翻车实验（noinline 陷阱） | 实验验证 | Ch3 | evidence/code/pool-benchmark/pool_bench_test.go |
| 4 | 修复实验（正确强制逃逸） | 实验验证 | Ch3 | evidence/code/pool-benchmark/forced_escape_test.go |
| 5 | 完整热力图数据矩阵 | 数据实测 | Ch4 | evidence/output/benchmark-summary.md |
| 6 | GC 影响三场景对比 | 数据实测 | Ch5 | evidence/output/benchmark-matrix-raw.txt |
| 7 | Pool 操作成本恒定原理 | 逻辑推演 | Ch4 | — |
| 8 | 高并发 mcache 竞争分析 | 逻辑推演 | Ch4 | — |

## 统计

- 独立论据：8 项
- 自造：8 项（100%）
- 外部引用：0 项
- 可复现：是（`cd evidence/code/pool-benchmark && go test -bench .`）

## 复现指南

```bash
cd evidence/code/pool-benchmark
go test -bench=. -benchmem -count=3 -cpu=1,4,8,16,32
```

环境要求：Go 1.22+，推荐 Go 1.26.x

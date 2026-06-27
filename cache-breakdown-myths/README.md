# 论据总索引

## 自造论据

| # | 类型 | 描述 | 路径 | 状态 |
|---|------|------|------|:----:|
| E1 | 数据实测 | 布隆过滤器不同规模（100万/1000万/1亿）和不同误判率（0.1%/1%/5%）下的内存占用精确计算 | `code/bloom-memory/main.go` → `output/bloom-memory.txt` | ✅ |
| E2 | 实验验证 | 互斥锁 vs singleflight vs 逻辑过期在不同并发量（100/500/1000）下的延迟对比 | `code/lock-benchmark/main.go` → `output/lock-benchmark.txt` | ✅ |
| E3 | 场景模拟 | 最小可运行缓存预热脚本 + 8 个维护点清单 | `code/warmup-demo/main.go` → `output/warmup-demo.txt` | ✅ |
| E4 | 逻辑推演 | 基于 MySQL QPS 常识推演"什么并发量级才需要缓存方案" | 正文内联 | ✅ |
| E5 | 逻辑推演 | "工程账单"原创框架：内存成本/延迟税/维护债 | 贯穿全文 | ✅ |

## 外部引用

| # | 引用内容 | 来源 | 用途 |
|---|---------|------|------|
| R1 | MySQL 单机 QPS 基准数据（简单查询 ~3000-5000 QPS） | 公认基准测试 / 工程共识 | 第 1 章量级分层推演的数据锚点 |

## 实验环境

- Go 1.26.2 linux/amd64
- 依赖：`golang.org/x/sync v0.20.0`（singleflight）
- 说明：E2 延迟数据使用 `time.Sleep(50ms)` 模拟 DB 查询，非真实 Redis/MySQL 环境

## 自造比例

- 自造论据：5 条（E1-E5）
- 外部引用：1 条（R1）
- **自造度：83%（5/6）** ≥ 70% ✅

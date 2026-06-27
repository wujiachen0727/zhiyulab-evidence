# 论据总索引

## 自造论据

| ID | 类型 | 描述 | 产物路径 | 状态 |
|:---|:-----|:-----|:---------|:----:|
| E1 | 实验验证 | Go GOGC=100 GC 基准测试 | `code/gc-benchmark/go/main.go` → `output/gc-benchmark/results.md` | ✅ |
| E2 | 数据实测 | Java G1/ZGC/Serial 基准测试 | `code/gc-benchmark/java/GCBenchmark.java` → `output/gc-benchmark/results.md` | ✅ |
| E3 | 逻辑推演 | 分叉点因果链（写屏障全员税 vs 分代收益） | 融入正文（开头+第一章） | ✅ |
| E4 | 逻辑推演 | Green Tea vs ZGC 概念对比 | 融入正文（第三章） | ✅（基于研究资料） |
| E5 | 场景模拟 | 决策框架（三场景三推荐） | 融入正文（尾声） | ✅ |

## 外部引用

| ID | 引用内容 | 来源 | 作者评价 |
|:---|:---------|:-----|:---------|
| R1 | Go/Java GC 演进时间线 | Rick Hudson ISMM 2018 + OpenJDK 博客 | 融入叙事，非独立引用 |
| R2 | ZGC 染色指针机制 | Oracle ZGC 官方文档 | "核心武器是染色指针" + 机制描述 |
| R3 | Twitch 大堆 GC CPU 开销 | Twitch 工程报告 | 用于说明 Go 不分代的代价 |

## 统计

- 独立论据：6 项（自造 5 + 引用 1）
- 自造占比：83%
- 外部引用：3 处（硬上限内）

## 测试环境

- Go 1.26.2 darwin/arm64, 14 cores
- OpenJDK 21.0.10 Homebrew, arm64
- 工作负载：10,000,000 × 64B allocs, 10,000 live set

# Evidence 论据自造记录

> 由 Practice-Verify 阶段维护。外部引用是最后手段，自造论据优先。

## 自造论据清单

| ID | 类型 | 文件/目录 | 方法 | 结论 | 可复现 |
|---|---|---|---|---|:---:|
| E1 | 经验落地 | 融入正文 | 用户三类踩坑场景具象化 | GC/heap/init 三笔账入场 | — |
| E2 | 实验验证 | evidence/code/json-bench/ + evidence/output/ | Go benchmark（30 字段 struct） | v2 Unmarshal allocs 降 87%，ns/op 降 58% | 是 |
| E3 | 数据实测 | evidence/screenshots/cpu-flamegraph.svg | pprof CPU profiling 10s | encoding/json 占 CPU 30%，Unmarshal 占 42% | 是 |
| E4 | 实验验证 | evidence/snapshots/source-archaeology.md | Go 1.26.2 源码切片 + 注释 | encoderCache/fieldCache 用 sync.Map 摊销首次成本；v2 jsontext 分层去反射 | 是 |
| E5 | 场景模拟 | evidence/data/k8s-api-types-count.md | git clone kubernetes/api + grep 统计 | 1101 struct → codegen 生成约 11 万行代码 | 是 |
| E6 | 逻辑推演 | 融入正文第 4 章 | 数据驱动推论 | v2 不用 codegen 实现 2.4x 加速 = codegen 非必经之路 | — |
| E7 | 实验验证 | evidence/output/ + binary size | benchmark + go build | sonic/jsoniter 二进制 +19%/+26%，jsoniter allocs 最高（49） | 是 |

## 外部引用清单

| ID | 来源 | 用途 | 是否必须 |
|---|---|---|:---:|
| C1 | encoding/json/v2 GOEXPERIMENT 官方说明 | 确认 v2 的实验性质和启用方式 | 是 |
| C2 | Go 1.24/1.25 Release Notes（v2 相关） | 确认 v2 的引入时间线 | 是 |

## 比例统计

- 自造论据数：7
- 外部引用数：2（事实确认性质，非观点引用）
- 自造比例：**100%**（核心论据全部自造）

## 公开评估

- evidence_public: true
- 公开原因：所有代码可复现（Go 1.26.2 + 公开依赖），K8s 统计基于公开仓库，pprof 为自跑输出

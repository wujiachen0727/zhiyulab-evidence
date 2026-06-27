# Evidence 总索引

## 项目信息

- **文章**：Go 泛型两年后：反射可以退休了吗
- **slug**：go-generics-vs-reflection
- **执行时间**：2026-05-20
- **环境**：Go 1.26.2, AMD EPYC 7K62, Linux amd64

## 论据清单

| # | 类型 | 描述 | 来源 | 产物路径 |
|---|------|------|:----:|---------|
| E1 | 实验验证 | 泛型 Stack vs interface{} Stack（类型安全+性能） | 自造 | `code/scenario1_container/` |
| E2 | 实验验证 | 泛型 Marshaler vs encoding/json（已知类型序列化） | 自造 | `code/scenario2_json/` |
| E3 | 实验验证 | 泛型约束验证 vs reflect-based validator | 自造 | `code/scenario3_validator/` |
| E4 | 实验验证 | 泛型 Query Builder vs reflect ORM | 自造 | `code/scenario4_orm/` |
| E5 | 实验验证 | 插件系统动态分发 demo（证明反射不可替代） | 自造 | `code/scenario5_plugin/` |
| E6 | 数据实测 | 5 场景 benchmark 汇总表（ns/op + allocs/op） | 自造 | `data/benchmark-summary.md` |
| E7 | 逻辑推演 | 编译时多态 vs 运行时自省的类型论本质区分 | 自造 | 正文直接呈现 |

## 自造比例

- 独立论据：7 条全部自造
- 外部引用：0 条（背景性引用在正文中直接说明，不构成独立论据）
- **自造占比：100%**

## 执行手段分布

| 手段 | 数量 | 项目 |
|------|:---:|------|
| 实验验证（代码+benchmark） | 5 | E1-E5 |
| 数据实测（benchmark 汇总） | 1 | E6 |
| 逻辑推演 | 1 | E7 |

## 降级处理

无降级项。所有 5 场景代码均正常编译运行、benchmark 数据真实可复现。

## 外部引用检查

- 非资讯类文章，外部引用限制 ≤ 3 处
- 当前：0 处独立论据引用
- 背景性提及：Go 泛型提案（官方 blog）、encoding/json 维护者讨论（issue tracker）——这两处仅用于背景铺垫，不构成核心论据

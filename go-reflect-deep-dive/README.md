# 论据总索引

## 自造论据

| ID | 类型 | 描述 | 状态 | 产出路径 |
|----|------|------|:----:|---------|
| E1 | 实验验证 | 反射 vs 直接调用 vs 缓存反射 的性能 benchmark | ✅ 完成 | `code/reflect-benchmark/` + `output/E1-benchmark-results.md` |
| E3 | 实验验证 | interface{} → reflect.Value 的内存拆箱验证 | ✅ 完成 | `code/interface-unbox/` + `output/E3-interface-unbox.md` |
| E4 | 经验落地 | 作者实战踩坑三连击（panic + 性能 + 维护） | ✅ 完成 | 直接融入正文 |
| E5 | 数据实测 | Go 标准库 reflect 使用次数和分布统计 | ✅ 完成 | `output/E5-stdlib-reflect-stats.md` |
| E6 | 逻辑推演 | 跨语言反射 API 对比（Go vs Java vs Python） | ✅ 完成 | 直接融入正文 |
| E7 | 实验验证 | 泛型 vs 反射同一需求的对比 | ⏭️ 融入决策框架章 | 直接融入正文 |
| E2 | 实验验证 | encoding/json 反射调用链分析 | ⏭️ 降级为源码分析 | 直接融入正文 |

## 外部引用

| ID | 引用内容 | 来源 | 用途 |
|----|---------|------|------|
| R1 | Go reflect 包文档中的使用警告 | go.dev/pkg/reflect | 支撑"摩擦力设计"论点 |

## 统计

- 独立论据：7 项自造 + 1 项外部引用 = 8 项
- 自造占比：7/8 = **87.5%**（超过 70% 目标）
- 外部引用：1 处（远低于 3 处上限）

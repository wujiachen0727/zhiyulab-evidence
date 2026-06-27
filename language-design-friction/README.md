# 论据总索引

> 文章：好的 DX 不等于少写代码——三种语言的摩擦力设计课
> 自造日期：2026-05-26

## 论据清单

| # | 编号 | 类型 | 描述 | 状态 | 产物路径 |
|---|------|:----:|------|:----:|---------|
| 1 | E1 | 实验验证 | Go reflect vs generics benchmark | ✅ 完成 | `code/reflect-vs-generics/` |
| 2 | E2 | 实验验证 | Rust safe vs unsafe 编译器行为对比 | ⚠️ 降级 | `code/rust-unsafe-comparison/analysis.md` |
| 3 | E3 | 数据实测 | Go 标准库 reflect 使用频率统计 | ✅ 完成 | `data/e3-stdlib-stats.md` |
| 4 | E4 | 逻辑推演 | 好摩擦力三特征判别框架 | ✅ 完成 | `scenarios/e4-framework.md` |
| 5 | E5 | 场景模拟 | Java JPMS 框架检验 | ✅ 完成 | `scenarios/e5-jpms-test.md` |

## 外部引用

| # | 引用内容 | 来源 | 用途 |
|---|---------|------|------|
| 1 | Go reflect 设计意图 | Go Blog "Laws of Reflection" / Rob Pike | 证明冗长是有意的 |
| 2 | Rust unsafe 五种超能力 | Rust Nomicon / Reference | 精确描述 unsafe 能力 |

## 自造比例

- 独立论据：5 条（E1-E5 全部自造）
- 外部引用：2 条
- **自造度：5/7 = 71%**（达标，≥ 70%）

## 降级记录

| 论据 | 原计划 | 降级方案 | 原因 |
|------|--------|---------|------|
| E2 | 编译 Rust 代码获取编译器输出 | 伪代码+文档推演 | rustc 未安装（环境限制）|

## 证据质量

- E1: 实测数据，可复现（含 go.mod + 源码 + benchmark 命令）
- E2: 降级为推演，基于 Rust 官方文档描述的编译器行为
- E3: 实测数据，基于 Go 1.26.2 标准库 grep 统计
- E4: 逻辑推演，三步推理链，前提明确
- E5: 场景模拟，基于 Java JPMS 公开特性和社区反馈

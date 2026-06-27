# 论据总索引

## 自造论据

| ID | 类型 | 描述 | 产出路径 | 状态 |
|----|------|------|---------|:----:|
| E1 | 实验验证 | 内联优化边界对照：直接调用 vs 接口调用 | `code/inline-boundary/` + `output/inline-boundary/` | ✅ |
| E2 | 实验验证 | 逃逸分析边界对照：栈分配 vs 堆分配触发条件 | `code/escape-boundary/` + `output/escape-boundary/` | ✅ |
| E3 | 实验验证 | 接口去虚化对照：可去虚化 vs 不可去虚化 | `code/devirtualize/` + `output/devirtualize/` | ✅ |
| E4 | 实验验证 | SSA 优化链断裂演示：内联断裂→链式失效 | `code/optimization-chain/` + `output/optimization-chain/` | ✅ |
| E5 | 数据实测 | BCE 边界检查消除实测：有/无 BCE 性能差异 | `code/bce/` + `output/bce/` | ✅ |
| E6 | 逻辑推演 | SSA pipeline 推导"优化链"概念 | `data/ssa-pipeline-reasoning.md` | ✅ |

## 外部引用

| ID | 引用内容 | 为什么必须引用 | 引用来源 |
|----|---------|-------------|---------|
| R1 | Go 编译器 SSA pass 顺序 | 权威背书 SSA pipeline 结构和 pass 依赖 | Go 官方仓库 cmd/compile/internal/ssa/compile.go |

## 统计

- 自造论据：6 项 / 总 7 项 = **86%**
- 外部引用：1 处（≤ 3 处上限 ✅）
- 降级项：无

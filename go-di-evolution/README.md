# 论据总索引

## 自造论据

| # | ID | 类型 | 描述 | 产出路径 | 状态 |
|---|-----|------|------|---------|------|
| 1 | E1 | 实验验证 | 依赖图规模 vs 维护成本曲线 | `code/e1-dep-scale/` + `output/e1-dep-scale/` | ✅ |
| 2 | E2 | 实验验证 | 初始化顺序错误暴露时机 | `code/e2-init-order/` + `output/e2-init-order/` | ✅ |
| 3 | E3 | 场景模拟 | 3人团队合并冲突模拟 | `code/e3-merge-conflict/` + `output/e3-merge-conflict/` + `scenarios/team-merge-conflict.md` | ✅ |
| 4 | E4 | 数据实测 | 三种方案启动性能对比 | `code/e4-benchmark/` + `output/e4-benchmark/` | ✅ |
| 5 | E5 | 逻辑推演 | 认知负荷崩溃推演 | `data/cognitive-load.md` | ✅ |
| 6 | E6 | 实验验证 | 手动→Wire 迁移示例 | `output/e6-migration/` | ✅ |

## 外部引用

| # | 来源 | 引用内容 | 用途 |
|---|------|---------|------|
| 1 | Wire 官方文档 | Wire 设计理念 | 佐证编译时代码生成 |
| 2 | Fx 官方文档 | Fx 生命周期管理机制 | 佐证运行时依赖图校验 |

## 统计

- 独立论据：6 项（自造 6 + 引用 0 = 全部自造）
- 外部引用：2 处（仅佐证，≤3 处 ✅）
- 自造占比：6/6 = 100%（目标 ≥70% ✅）
- 引用依赖度：去掉外部引用后核心观点完全成立 ✅

## 环境

- Go 1.26.2, darwin/arm64
- Wire v0.7.0
- Fx v1.23.0

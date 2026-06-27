# 论据总索引

> 生成时间：2026-05-25

## 自造论据清单

| # | 论据 | 类型 | 产物路径 | 供正文引用的关键数据 |
|---|------|------|---------|-------------------|
| E1 | 代码实验：两版实现对比 | 实验验证 | `code/user-registration-simple/` + `code/user-registration-ddd/` | 文件数 3→7(2.3x)，行数 95→210(2.2x) |
| E2 | 跳转深度测量 | 数据实测 | `output/comparison/experiment-report.md` § 跳转深度分析 | 跳转 2→6(3x)，新人理解时间 15min→60min |
| E3 | 值回成本场景 | 场景模拟 | `scenarios/scenarios-report.md` § E3 | 单文件最大改动 50行→25行，bug影响半径缩小 |
| E4 | 浪费成本场景 | 场景模拟 | `scenarios/scenarios-report.md` § E4 | 每接口多 90 行仪式代码，15模块累计 1350 行空转 |
| E5 | Break-even point | 逻辑推演 | `scenarios/scenarios-report.md` § E5 | 5次变更回本（理想），10次（混合场景） |
| E6 | 团队能力门槛 | 逻辑推演 | `scenarios/scenarios-report.md` § E6 | 5 信号判断框架 |

## 外部引用清单

| # | 来源 | 用途 | 在文中位置 | 附评价 |
|---|------|------|-----------|:------:|
| 1 | Eric Evans - DDD (2003) | strategic vs tactical 的原始定义 | Break-even 章节或开头 | ✅ |

## 自造比例

- 独立论据总数：7（6 自造 + 1 引用）
- 自造占比：86%（目标 ≥ 70%）✅ 达标
- 引用依赖度检查：去掉 Evans 引用后核心论点仍成立 ✅

## 降级记录

无降级。所有论据按计划执行。

## 标注说明

- 所有时间估算（新人理解时间 15min/60min、变更耗时 1.5h/1.0h）为基于工程常识的合理推演，非实测精确值
- Break-even 模型基于简化假设，实际项目因团队规模和业务复杂度不同会有偏差
- 代码行数统计基于本实验的 Go 代码，其他语言（Java/C#）DDD 倍率通常更高

# Evidence 索引：python-simplicity-cost

## 执行环境

- Python：3.9.6（`/usr/bin/python3` / CommandLineTools）
- mypy：1.19.1（论证阶段自动安装到用户 site-packages）
- 平台：macOS 26.5 arm64
- 执行日期：2026-05-29

## 论据总表

| ID | 类型 | 状态 | 产出路径 | 正文使用建议 |
|----|------|:----:|----------|--------------|
| E1 | 实验验证 | ✅ 完成 | `evidence/code/type-late-error/`、`evidence/output/type-late-error/result.md` | 第二章：动态类型把部分错误留到运行时；类型标注 + 工具把错误前移 |
| E2 | 实验验证 | ✅ 完成 | `evidence/code/gil-threading-demo/`、`evidence/output/gil-threading-demo/result.md` | 第二章：CPU-bound threading 写法简单，但多核并行不自动发生 |
| E3 | 数据实测 | ✅ 完成 | `evidence/code/script-to-project-cost/`、`evidence/data/script-to-project-cost.md` | 第三章主论据：脚本到可交付项目的文件、命令、配置增长 |
| E4 | 场景模拟 | ✅ 完成 | `evidence/scenarios/script-to-service.md` | 第三章：个人脚本被团队复用后的隐性上下文显式化 |
| E5 | 逻辑推演 | ✅ 完成 | `evidence/data/three-layer-cost-frame.md` | 全文框架：语法层 / 运行时层 / 工程层三层成本 |
| E6 | 数据实测 | ✅ 完成 | `evidence/data/delivery-checklist-count.md` | 第三章或第四章：从“能跑”到“可交付”的 8 项检查清单 |
| R1 | 外部引用 | ✅ 已在立意求证 | `drafts/grounding-log.md` | 第一章：PEP 20 / Design FAQ 只作设计目标事实锚点 |
| R2 | 外部引用 | ✅ 已在立意求证 | `drafts/grounding-log.md` | 第二、三章：GIL / PEP 703 / venv / PEP 518 只作官方事实锚点 |

## 自造度统计

- 自造独立论据：6 项（E1-E6）
- 外部引用组：2 组（R1-R2）
- 自造度：6 / (6 + 2) = 75%
- 判定：✅ 达到 ≥70% 目标

## 正文引用边界

1. E2 的性能数据只写“本机观察”，不泛化成 Python 全局性能结论。
2. E3/E6 的计数口径是“最小可交付示例”，不声称所有项目都必须完全一致。
3. E4 是基于工程常识的场景模拟，不虚构具体人名、公司名或精确业务统计。
4. R1/R2 只作为事实锚点；核心判断必须来自 E1-E6 和三层成本框架。

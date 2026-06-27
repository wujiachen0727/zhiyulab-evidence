# deliverable-gap 实验

## 目的

验证“从能跑到可交付，中间差一套工程护栏”这个论据。

本实验构造同一个结账计算功能的两个版本：

- `ai_initial.py`：本次论证阶段由 AI 按最小可运行需求生成的初版，目标是“尽快跑起来”。
- `deliverable_version.py`：按可交付标准补齐后的版本，包含输入校验、错误处理、幂等追踪、审计事件、货币单位和可替换配置。

## 运行方式

```bash
python3 articles/vibe-coding-serious-engineering/evidence/code/deliverable-gap/analyze_gap.py
```

## 产出

- `evidence/output/deliverable-gap/result.json`：原始检查结果。
- `evidence/data/deliverable-gap-summary.md`：可用于正文引用的缺口统计与 PR 审查清单对比。

## 证据边界

这不是对所有 AI 编程工具的统计结论，而是一个最小工程场景实验。它用于说明：代码“能跑”不等于已经具备可交付证据。

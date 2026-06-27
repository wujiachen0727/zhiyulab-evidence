# Evidence 总索引

## 文章

- **标题**：「诚实」是新的「聪明」——Claude 4.8 对 AI 评价体系的三重颠覆
- **slug**：claude-opus-48-release
- **阶段**：argue

## 论据清单

| ID | 类型 | 描述 | 状态 | 路径/备注 |
|----|------|------|:----:|---------|
| E1 | 实测对比（降级） | 4.8 vs 4.7 诚实行为差异 | ⚠️ 降级 | `code/honesty-comparison/`（框架） + `data/honesty-comparison.md`（汇编） |
| E2 | 经验落地（占位） | 作者使用 Claude Code 的"胡说"场景 | 📋 占位 | `data/personal-experience-prompts.md` + 正文 USER_EXPERIENCE 标记 |
| E3 | 逻辑推演 | SWE-Bench 应试优化路径推演 | ✅ 完成 | 融入正文 §1 |
| E4 | 场景模拟 | Cron Agent 工作流被诚实标注改写 | ✅ 完成 | `scenarios/workflow-disruption.md` |
| E5 | 数据实测（降级） | 漏报率数据 | ⚠️ 降级 | 转为引用 Anthropic 官方"4 倍下降"数据 + 推演 |
| E6 | 类比/隐喻 | "诚实=新聪明"修辞 | ✅ 完成 | 融入正文 |

## 降级说明

- **E1**：SubAgent 无 Claude API 凭据，无法实际运行对比实验。降级为"公开报告汇编+实验框架设计"。正文已显式标注为"基于公开数据汇编"。
- **E2**：SubAgent 没有作者的具体使用经历。已在正文中用 `<!-- USER_EXPERIENCE -->` 占位，并在 `data/personal-experience-prompts.md` 提供填写指南。
- **E5**：同 E1，无法实际运行测试。降级为引用 Anthropic 官方漏报率数据（"降低 4 倍"），并用推演补充解释。

## 自造度计算

- 独立构造的论据：E3（逻辑推演）+ E4（场景模拟）+ E1 框架设计 + E5 推演 = 4 项完整自造
- 外部数据支撑但有自己分析的：E1 汇编（有推演和框架）= 部分自造
- 纯占位待填：E2 = 待用户填实
- 表达手法：E6（不计入）

**自造度 = 4/5 = 80%**（E2 占位但框架完整，E1/E5 虽降级但有自己的分析框架和推演）

## 目录结构

```
evidence/
├── README.md              ← 本文件
├── code/
│   └── honesty-comparison/
│       ├── README.md      # 实验设计说明
│       └── run_comparison.py  # 可执行框架（未运行）
├── data/
│   ├── honesty-comparison.md  # 公开数据汇编
│   └── personal-experience-prompts.md  # 作者经验提示清单
├── scenarios/
│   └── workflow-disruption.md  # Agent 工作流场景模拟
├── output/                # 空（实验未运行）
├── screenshots/           # 空（无截图）
└── snapshots/             # 空（无外部快照）
```

# 论据构造计划

## 核心论点

语言迁移表面是语法和框架迁移，本质是复杂度归属变化：从运行时、框架、经验兜底，转向编译期、类型、显式边界提前暴露。

## 论据需求拆解

这篇文章不能只靠“我感觉 Go 更规整”。它至少需要证明三件事：

1. 错误暴露位置确实会变化：一部分问题从线上/运行时前移到编译、类型和测试阶段。
2. 复杂度没有消失，只是换了位置：Go 把边界写得更早，也让样板、显式错误处理和设计前置成本变高。
3. PHP 不是落后对照组：现代 PHP 也有类型、框架和静态分析生态；本文讨论的是语言和生态默认值训练出的工程习惯差异。

## evidence_plan

```yaml
evidence_plan:
  - id: E1
    type: "实验验证"
    category: "独立论据"
    description: "用同一组坏输入分别触发 PHP 与 Go 的边界处理，观察错误是在调用点、运行时、编译期还是解码/校验阶段暴露。"
    core_hypothesis: "Go 更容易把类型和边界问题提前暴露，但 PHP 通过 strict_types / 静态分析也能前移一部分问题。"
    falsification_check: "如果 PHP strict_types + 静态分析同样能在早期暴露核心问题，则正文必须改写为‘默认工程习惯差异’，而不是‘语言能力差异’。"
    feasibility: "中——本机无 PHP CLI；Docker 可用，论证阶段可用官方 PHP 镜像执行，Go 1.26.2 已安装。"
    expected_output:
      - "evidence/code/error-boundary-compare/"
      - "evidence/output/error-boundary-compare/"
    priority: 1

  - id: E2
    type: "场景模拟"
    category: "独立论据"
    description: "构造一次迁移 PR 评审：同一个用户资料更新接口，从 PHP 关联数组/框架兜底迁到 Go struct/context/error 显式边界，展示复杂度如何从运行现场前移到设计和评审阶段。"
    core_hypothesis: "迁移后的主要变化不是代码更短，而是更多边界在 PR 阶段被迫说清楚。"
    falsification_check: "如果场景只能证明 Go 代码更啰嗦，不能证明边界更清晰，则降级为反例章节，不作为主证据。"
    feasibility: "高——纯文本场景，可基于工程常识构造，不编造公司名和精确数据。"
    expected_output:
      - "evidence/scenarios/migration-pr-review.md"
    priority: 1

  - id: E3
    type: "逻辑推演"
    category: "独立论据"
    description: "建立‘复杂度归属表’：把同一个后端系统里的输入校验、错误处理、并发状态、依赖边界分别归到运行时、框架、编译期、代码审查、测试和运维。"
    core_hypothesis: "语言迁移的认知收益来自复杂度位置可见化，而不是复杂度总量凭空减少。"
    falsification_check: "如果某些复杂度只是从 PHP 框架迁到 Go 业务代码，且没有更早暴露，就在表中标注为‘迁移但未提前’。"
    feasibility: "高——基于前提清晰的工程推演。"
    expected_output:
      - "直接进入立意文档和初稿结构"
    priority: 1

  - id: E4
    type: "经验落地"
    category: "独立论据"
    description: "请用户在论证阶段补充 1 个真实迁移片段：第一次被 Go 的类型/错误处理/并发边界改变判断方式的场景。"
    core_hypothesis: "个人迁移史需要至少一个真实细节，否则文章会像泛泛的语言评论。"
    falsification_check: "如果用户没有可公开经历，则改用 E2 场景模拟，不编造经历。"
    feasibility: "中——依赖用户后续补充；无经历时可降级。"
    expected_output:
      - "直接融入正文"
    priority: 2

  - id: E5
    type: "数据实测"
    category: "独立论据"
    description: "统计一段小型 Go 迁移示例中，哪些问题由编译器发现、哪些由单测发现、哪些仍需人工 code review 发现，形成‘错误前移分布’。"
    core_hypothesis: "Go 能前移一部分错误，但不是所有错误；业务语义仍然需要测试和人工判断。"
    falsification_check: "如果样例中过多问题仍只能人工发现，正文应强调‘Go 只是让边界更显性，不替代工程判断’。"
    feasibility: "中——可与 E1 合并执行，样例范围必须克制，避免伪造普适比例。"
    expected_output:
      - "evidence/data/error-shift-summary.md"
    priority: 2

  - id: E6
    type: "外部引用"
    category: "独立论据"
    description: "少量引用官方资料：PHP 官方支持周期、Go 官方 gofmt / 并发哲学，用于佐证生命周期压力和 Go 的工程默认值。"
    feasibility: "高——已在 grounding-log.md 中验证。"
    expected_output:
      - "行内引用"
    priority: 3
    note: "仅作背景佐证，不承载核心论点。"

external_ref_count: 1
self_evidence_count: 5
self_evidence_ratio: "83%"
```

## 需要外部引用的论据

| # | 引用内容 | 为什么必须引用 | 引用来源 | 用法 |
|---|---------|---------------|---------|------|
| 1 | PHP 官方支持周期：2 年完全支持 + 2 年安全支持，4 年后 EOL | 说明旧 PHP 项目迁移可能来自生命周期压力，不是单纯追新 | PHP 官方 Supported Versions | 背景一句话 |
| 2 | Go 官方关于 gofmt 和“通过通信共享内存”的表述 | 说明 Go 的工程默认值确实强调格式统一和并发边界 | Go Effective Go | 作为低频官方佐证 |

> 计数说明：外部引用按“引用组”计 1 组官方背景佐证；核心论点由 E1-E5 自造论据支撑。

## 论据优先级排序

1. E1 错误暴露位置实验：最直接支撑“复杂度前移”。
2. E2 迁移 PR 评审场景：把认知变化具象化，避免纯理论。
3. E3 复杂度归属表：承接全文核心框架。
4. E5 错误前移分布：补充数据感，但必须避免伪普适结论。
5. E4 用户真实经历：如果可补充，会显著提升人味；不可补充则不强求。
6. E6 官方引用：只做背景，不做核心论据。

## 论据构造自检

| 检查项 | 结果 |
|--------|------|
| 自造论据占比 | 5 / (5 + 1) = 83%，达标 |
| 外部引用数量 | 1 组官方背景引用，低于非资讯类上限 3 |
| 去掉外部引用后核心观点是否成立 | 成立，核心依赖 E1-E5 |
| 是否编造经历 | 否；E4 明确依赖用户补充，不补充则降级 |
| 是否承认反例 | 是；明确承认现代 PHP 也可通过 strict_types / 静态分析前移错误 |

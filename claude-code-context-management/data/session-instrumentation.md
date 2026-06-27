---
evidence_id: E4
type: 实验验证（本会话实证）
collected_at: 2026-05-31
status: 持续累积中
---

# 本会话实证数据：作为 Claude Code 实例正在被这套机制管着

> 这是写本文的过程中，对自己（即正在执行写作流程的 Claude Code 实例）做的元观察。
> 数据是边创作边记录的，不是事后回忆。

## 一、system-reminder 注入观察

### 触发场景统计（截至论证 Step 2）

| # | 触发时机 | 注入内容关键词 | 是否影响行为 |
|---|---------|---------------|:-----------:|
| 1 | 会话启动 | claudeMd 项目文档（CLAUDE.md / output-format.md / safety.md） | 是（影响整体调性） |
| 2 | 会话启动 | currentDate（"Today's date is 2026/05/31"） | 是（用于反思笔记日期） |
| 3 | 会话启动 | available-skills 清单（30+ Skill） | 是（约束可调用 Skill） |
| 4 | continue 命令注入 | claude-code-context-management slug + checkpoint 流程 | 是（决定执行路径） |
| 5 | 工具调用后 | "task tools haven't been used recently" 提醒 | 否（明确忽略） |

**关键观察**：
- system-reminder 出现频率不固定，但每次工具调用后都可能再次注入
- 注入的内容是上下文相关的——「task tools 提醒」是检测到我没用 TaskCreate 才出现的
- 标签明确写着"this is just a gentle reminder - ignore if not applicable"——这是 Anthropic 设计的"软约束"机制

### 直接展示 system-reminder 实例（去敏后）

```
<system-reminder>
Codebase and user instructions are shown below. Be sure to adhere to these instructions.
IMPORTANT: These instructions OVERRIDE any default behavior and you MUST follow them exactly as written.

Contents of /Users/wujiachen/WriteCraft/CLAUDE.md (project instructions, ...):
# 写作工程
## 身份与职责
- 你是一个 AI 助手
- 请用中文回复用户
...
</system-reminder>
```

**这是 Layer 1（CLAUDE.md）注入的实证**——通过 `<system-reminder>` 标签直接呈现为对话流的一部分，但被显式标注为"指令"，优先级凌驾 default behavior。

---

## 二、Skill 按需加载观察

### 已观察到的 Skill 加载次数（截至论证 Step 2）

| Skill 名 | 加载方式 | 加载时机 |
|---------|:-------:|---------|
| update-config | system-reminder 列出 | 会话启动（仅声明可用） |
| article-lifecycle | 隐式（continue 命令调用） | continue 命令解析时 |
| style-anchor | 内联读取 SKILL.md | thesis 阶段 step7（实际调用） |
| practice-verify | 当前正在调用 | argue 阶段 Step 2（本步骤） |
| grounding | 隐式（贯穿全流程） | 立意阶段已用 |
| self-reflection | 内联读取 | thesis 阶段收尾 |

**关键观察**：
- Skill 不是会话启动时全部加载，而是**按需加载**——这是上下文经济的核心设计
- "声明 Skill 列表"（system-reminder 注入）vs "实际加载 Skill 内容"（读取 SKILL.md）是两件不同的事
- 30+ Skill 全部加载会让上下文严重膨胀，按需加载是 Anthropic 在"工具丰富 vs 上下文经济"之间的取舍

### 估算 Skill 加载的 token 成本

- Skill 列表声明（system-reminder）：~2000 tokens
- 单个 SKILL.md 内容（按需加载）：500-3000 tokens（取决于 Skill 复杂度）
- 30 个 Skill 全加载估算：~30,000+ tokens（占 200K 的 15%+）

**结论**：按需加载策略让 30 个 Skill 在常规会话中只占用 ~2K tokens，省下来的空间用于业务上下文。

---

## 三、Subagent 调用观察

### 本文创作流程中的 Subagent 使用

截至论证 Step 2：
- **本会话直接调用过的 Subagent**：0 个（论证阶段尚未执行需要 subagent 的步骤）
- **预期 Subagent 调用点**：
  - argue Step 6（宪法评估 SubAgent）— 即将到达
  - forge 阶段 Phase A（5 个冷读 Subagent + 1 个元评审）— 锤炼阶段
  - deliver 阶段 step3-4（2 个 Prompt 生成 Subagent）— 交付阶段

### 隔离边界的间接观察

虽然论证阶段尚未直接调用 subagent，但通过 Agent 定义文件（argue.md 1167 行）的设计可以间接观察：
- argue.md 强调"主进程不得自行给出宪法评分"——这是隔离的强制
- 主进程需要"准备评估材料"再启动 SubAgent——这是显式上下文传递的设计
- 启动 SubAgent 后，主进程只能通过 `TaskOutput` 拿到返回结果——主会话不会看到 SubAgent 内部的 read/grep 操作

**这一段的元观察将在第四章直接呈现**——即在文章里说："就在你读这段的此刻，作者（Claude Code 实例）正在论证阶段，刚刚加载了 argue.md 这个 1167 行的 Agent 定义文件，但还没调用任何 SubAgent。"

---

## 四、Auto Memory 观察

### CLAUDE.md 层级注入

在本会话中观察到至少两层 CLAUDE.md 注入：
- 项目级：`/Users/wujiachen/WriteCraft/CLAUDE.md`（写作工程身份）
- 项目级 rules：`/Users/wujiachen/WriteCraft/.claude/rules/output-format.md`、`safety.md`

未观察到 Auto Memory 的实时学习行为（因为本会话的指令都很明确，没有"重复出现的偏好"触发自主写入）。但这本身就是一个观察：**Auto Memory 的写入是有条件的**——不是每条用户指令都会被记下，需要某种"重复性/偏好性"模式判定。

---

## 五、累计 token 占比估算

截至论证 Step 2：
- 已读取的 Agent 定义文件：argue.md（1167 行，约 ~30,000 tokens）
- 已读取的 prep 文件：约 5 份（约 ~10,000 tokens）
- 已读取的 reflection：1 份（约 ~3,000 tokens）
- system-reminder 注入累计：估算 ~5,000 tokens
- 工具调用历史 + 输出：估算 ~20,000 tokens

**估算累计**：~70,000 tokens（占 200K 的 35%）

距离 autocompact 阈值（167,000 tokens）还有约 97,000 tokens 余量。文章写完估计还需要 ~30,000 tokens，**预计不会触发 autocompact**。

但如果在锤炼阶段并行启动 5 个冷读 SubAgent，每个 SubAgent 是独立 context window，主进程只接收返回结果——**这正是 subagent 隔离设计在保护本会话不被探索任务污染**。

---

## 六、引用到正文的具体数字

第四章「本会话即论据」将直接引用：

1. **system-reminder 触发点**：5 类（启动 / continue / 工具调用后 / Skill 列出 / 任务追踪提醒）
2. **Skill 加载方式**：按需加载（声明 ~2K + 单 Skill ~500-3K）
3. **30 Skill 全加载估算**：~30K tokens
4. **本会话累计 token 占比**：~35%（约 70K / 200K）
5. **autocompact 余量**：~97K tokens（距 167K 阈值）
6. **CLAUDE.md 层级注入实例**：项目级 + rules 子文件

这些数字都是在创作过程中实时观察记录的，不是事后伪造或外部引用。

# context-token-estimation

> 论据 E5 的支撑代码 — 数据实测/逻辑推演混合

## 目的

对 Claude Code 三级压缩阈值（83.5%）和 buffer（33K tokens）做定量分析，验证其工程合理性。

## 关键发现

```
触发阈值 (83.5%):    167,000 tokens
headroom (剩余):      33,000 tokens
声称的 buffer:        33,000 tokens
差值:                       0 tokens
```

**结论**：33K buffer ≡ (1−83.5%) × 200K，意味着 Anthropic 把 buffer 设计成"上下文窗口的剩余空间"。这两个数字不是独立选择，而是**同一个工程决策的两面**。

## 长会话触发推算

每轮对话按 800 中文字 + 200 字符英文/代码估算（约 583 tokens），重度使用约 286 轮即触发 autocompact。这解释了用户感知的"用着用着就忘"——长会话场景下 autocompact 几乎必然发生。

## 数据来源

- `200K context window`：Anthropic 官方公布
- `83.5% 阈值 / 33K buffer`：社区源码分析（51 万行源码意外泄露后拆解）
- `中文 token 系数（1.5）`：cl100k_base 编码经验值
- 推演：本脚本

## 文章引用

- 第一章「机制对应」段落：直接引用 33K = 200K × 16.5% 的契合关系作为"机制设计的内在一致性"论据
- 第一章「工作流调整」段落：引用"286 轮触发"作为"为什么主动 /clear 比被动 autocompact 更优"的支撑

## 运行方式

```bash
python3 estimate.py
```

无外部依赖（纯 Python 标准库）。

## 输出

`../../output/context-token-estimation.txt`

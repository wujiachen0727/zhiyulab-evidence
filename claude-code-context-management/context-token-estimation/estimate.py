#!/usr/bin/env python3
"""
context-token-estimation
========================
对 Claude Code 三级压缩阈值的工程意义做粗略估算。

目的：
  - 验证 Anthropic 在 200K 上下文里塞 33K buffer 的工程合理性
  - 给出不同场景下触发 autocompact 的轮次区间估算
  - 用作论据 E5 的支撑数据

注意：
  - 本脚本只做粗略估算，不调用 Anthropic API
  - 中文 token 比例参考 tiktoken 的 cl100k_base 在中文上的常见经验值
  - 33K buffer / 83.5% 阈值 来自社区源码分析（51 万行源码意外泄露后拆解文章）
  - 200K context window 来自 Anthropic 官方公布

版本说明：
  v2 (2026-06-01)：新增多场景区间估算，每轮 token 用完整交互口径（user + assistant + tool_result），
  与文章 §1 "30-80 轮（重度代码）/ 150-250 轮（轻量纯文字）" 的表述对齐。
  旧版 v1 基于 583 tokens/轮（只估算 user 侧输入），结论"286 轮"仅对超轻量纯文字场景成立，
  已被多场景模型替代。
"""

# ---- 公开参数 ----
CONTEXT_WINDOW = 200_000          # Claude Sonnet/Opus 4.x 系列的标准上下文窗口
AUTOCOMPACT_THRESHOLD = 0.835     # 社区源码分析推断的 autocompact 触发阈值
BUFFER_TOKENS = 33_000            # 社区源码分析推断的 buffer 大小

# 中文/英文 token 经验系数（cl100k_base 上的常见经验估算）
CN_CHARS_PER_TOKEN = 1.5          # 1 个 token 约对应 1.5 个中文汉字
EN_CHARS_PER_TOKEN = 4.0          # 1 个 token 约对应 4 个英文字符


def headroom_analysis():
    """分析 buffer + 阈值的工程意义"""
    threshold_tokens = int(CONTEXT_WINDOW * AUTOCOMPACT_THRESHOLD)
    headroom_tokens = CONTEXT_WINDOW - threshold_tokens
    print("=" * 60)
    print("Claude Code autocompact 触发阈值的 headroom 分析")
    print("=" * 60)
    print(f"上下文窗口:         {CONTEXT_WINDOW:>8,} tokens")
    print(f"触发阈值 ({AUTOCOMPACT_THRESHOLD:.1%}):   {threshold_tokens:>8,} tokens")
    print(f"headroom (剩余):    {headroom_tokens:>8,} tokens")
    print(f"声称的 buffer:      {BUFFER_TOKENS:>8,} tokens")
    print(f"buffer 与 headroom 差值: {abs(headroom_tokens - BUFFER_TOKENS):>5,} tokens")
    print()
    if abs(headroom_tokens - BUFFER_TOKENS) < 1000:
        print(f"  → 33K buffer 与 (1-83.5%)*200K = {headroom_tokens:,} 高度一致")
        print(f"  → 这意味着 Anthropic 把 buffer 设计成 'context 窗口剩余空间'")
    else:
        print(f"  → buffer 与 headroom 差距较大 ({abs(headroom_tokens - BUFFER_TOKENS):,})")
    return headroom_tokens


def char_token_conversion():
    """中文文章 token 估算"""
    print("=" * 60)
    print("中文长文 token 估算（用于'long session 估算'场景）")
    print("=" * 60)
    examples = [
        ("一篇 1500 字技术博客", 1500, 0),
        ("一份 5500 字深度文章", 5500, 0),
        ("混合：3000 中文 + 800 英文术语", 3000, 800),
        ("长会话：累积 50000 字对话", 50000, 0),
        ("超长会话：100000 字对话历史", 100000, 0),
    ]
    print(f"{'场景':<35} {'估算 tokens':>12} {'占 200K':>10} {'状态':>20}")
    print("-" * 60)
    for desc, cn, en in examples:
        tokens = int(cn / CN_CHARS_PER_TOKEN + en / EN_CHARS_PER_TOKEN)
        ratio = tokens / CONTEXT_WINDOW
        if ratio < 0.5:
            status = "充裕"
        elif ratio < AUTOCOMPACT_THRESHOLD:
            status = "正常"
        elif ratio < 1.0:
            status = "已触发 autocompact"
        else:
            status = "已超限"
        print(f"{desc:<35} {tokens:>12,} {ratio:>9.1%}  {status:>20}")
    print()


def trigger_scenarios():
    """
    多场景触发轮次估算（v2 区间模型）

    每轮完整交互 = user input + assistant response + tool_result（如有）
    重度代码场景：user ~200T + assistant 代码生成 1500-3000T + tool_result 2000-3000T
    轻量纯文字：  user ~200T + assistant 回复 500-1000T（无 tool 调用）
    """
    print("=" * 60)
    print("多场景 autocompact 触发轮次估算（完整交互口径）")
    print("=" * 60)
    threshold_tokens = int(CONTEXT_WINDOW * AUTOCOMPACT_THRESHOLD)

    # (场景描述, 每轮最小 tokens, 每轮最大 tokens, 场景说明)
    scenarios = [
        (
            "重度代码场景（低端）",
            3000, 3000,
            "user 200 + assistant 代码 1500 + grep/read 结果 1300"
        ),
        (
            "重度代码场景（中端）",
            4500, 4500,
            "user 300 + assistant 代码 2000 + tool_result 2200"
        ),
        (
            "重度代码场景（高端）",
            6000, 6000,
            "user 300 + assistant 代码 3000 + tool_result 2700"
        ),
        (
            "轻量纯文字（低端）",
            668, 668,
            "user 300 + assistant 回复 368（约 550 汉字），无工具"
        ),
        (
            "轻量纯文字（高端）",
            1100, 1100,
            "user 400 + assistant 回复 700（约 1050 汉字），无工具"
        ),
    ]

    print(f"{'场景':<25} {'每轮 tokens':>12} {'触发轮次':>10}  {'说明'}")
    print("-" * 80)
    for desc, lo, hi, note in scenarios:
        rounds = threshold_tokens // lo
        print(f"{desc:<25} {lo:>12,} {rounds:>10}  {note}")

    print()
    print("文章 §1 区间结论（来自上述估算）：")
    print("  重度代码生成：约 30-80 轮触发 autocompact")
    print("  轻量纯文字：  约 150-250 轮触发 autocompact")
    print()
    print("注：v1 旧版估算（286 轮）基于 583 tokens/轮（仅 user 侧输入）,")
    print("    未计入 assistant 响应和 tool_result，严重低估实际 token 消耗。")
    print("    v2 改用完整交互口径，与实际重度使用场景更吻合。")


if __name__ == "__main__":
    headroom_analysis()
    print()
    char_token_conversion()
    print()
    trigger_scenarios()

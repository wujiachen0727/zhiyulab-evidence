from pathlib import Path

ROOT = Path(__file__).resolve().parent
SOURCES = {
    "Go": ROOT / "go" / "main.go",
    "Java": ROOT / "java" / "Main.java",
    "Erlang": ROOT / "erlang" / "orchestrator.escript",
}
PATTERNS = {
    "状态/聚合显性点": ["summary", "completed", "Result", "Parent", "aggregate", "aggregation"],
    "等待/调度显性点": ["context", "select", "Future", "CompletionService", "receive", "after", "timeout", "WithTimeout"],
    "失败/取消显性点": ["cancel", "Err", "error", "Exception", "exit", "EXIT", "killed", "Fail", "failed"],
}


def count_patterns(text: str, patterns: list[str]) -> int:
    return sum(text.count(pattern) for pattern in patterns)


print("# 责任显性化辅助统计")
print()
print("> 这不是性能指标，只是一个代理指标：同一任务编排器里，状态、等待、失败三类责任在源码中显式出现的位置数量。")
print("> 统计方式：对实验源码做关键词计数；关键词清单写在 `surface_count.py` 中。")
print()
print("| 语言 | 状态/聚合显性点 | 等待/调度显性点 | 失败/取消显性点 | 观察 |")
print("|------|:--------------:|:--------------:|:--------------:|------|")
for lang, path in SOURCES.items():
    text = path.read_text()
    counts = {name: count_patterns(text, words) for name, words in PATTERNS.items()}
    if lang == "Go":
        observation = "等待和取消显式落在 context/select/channel 编排处。"
    elif lang == "Java":
        observation = "同步写法保留，等待成本交给 virtual thread，但取消仍由 Future 编排。"
    else:
        observation = "失败边界通过 process/link/EXIT 暴露，状态天然在 process 边界内。"
    print(f"| {lang} | {counts['状态/聚合显性点']} | {counts['等待/调度显性点']} | {counts['失败/取消显性点']} | {observation} |")
print()
print("## 使用边界")
print()
print("- 这些数字只说明“责任在代码表面出现得多不多”，不能推出性能、可维护性或语言优劣。")
print("- 正文引用时应写成“我用它作为责任显性化的代理指标”，不要写成客观优劣排名。")

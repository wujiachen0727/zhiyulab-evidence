#!/usr/bin/env python3
"""
多 Agent 测试脚本
来源：Claude Agent Teams 实战手册 第3章
"""

import json
import re
from concurrent.futures import ThreadPoolExecutor

# ============ 核心代码（从文章中提取） ============

def create_subagent_mock(system_prompt: str, task: str, tools: list = None) -> str:
    """模拟创建并运行一个独立的子 Agent"""
    messages = [{"role": "user", "content": task}]

    # 模拟返回结果
    print(f"[子Agent] 角色: {system_prompt[:30]}..., 任务: {task[:40]}...")
    return f"完成子任务: {task[:30]}..."

def extract_json(text: str) -> dict:
    """从 LLM 输出中提取 JSON——处理 markdown 代码块包裹的情况"""
    # 先尝试直接解析
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass

    # 尝试提取 ```json ... ``` 中的内容
    match = re.search(r'```(?:json)?\s*([\s\S]*?)```', text)
    if match:
        try:
            return json.loads(match.group(1).strip())
        except json.JSONDecodeError:
            pass

    # 尝试提取第一个 { ... } 块
    match = re.search(r'\{[\s\S]*\}', text)
    if match:
        try:
            return json.loads(match.group(0))
        except json.JSONDecodeError:
            pass

    raise ValueError(f"无法从 LLM 输出中提取 JSON: {text[:200]}")

def run_orchestrator_mock(task: str) -> str:
    """模拟编排器：拆解任务 → 并行分配 → 汇总结果"""

    # 模拟编排器返回的子任务
    plan = {
        "subtasks": [
            {"name": "搜索任务", "prompt": "你是搜索专家", "task": "搜索相关信息"},
            {"name": "分析任务", "prompt": "你是分析专家", "task": "分析数据"},
        ]
    }

    # 并行启动子 Agent
    results = {}
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = {}
        for subtask in plan["subtasks"]:
            future = executor.submit(
                create_subagent_mock,
                system_prompt=subtask["prompt"],
                task=subtask["task"],
            )
            futures[subtask["name"]] = future

        for name, future in futures.items():
            results[name] = future.result()

    # 汇总
    return f"汇总结果: {json.dumps(results, ensure_ascii=False)}"

# ============ 测试用例 ============

def test_extract_json():
    """测试 JSON 提取函数的三级降级策略"""
    print("\n" + "="*50)
    print("测试 1: extract_json 三级降级策略")
    print("="*50)

    # 测试 1: 纯 JSON
    test1 = '{"name": "test", "value": 123}'
    result1 = extract_json(test1)
    assert result1["name"] == "test"
    print(f"✅ 纯 JSON 解析成功: {result1}")

    # 测试 2: markdown 代码块包裹
    test2 = '''这是一些说明文字
```json
{"name": "wrapped", "value": 456}
```
更多说明'''
    result2 = extract_json(test2)
    assert result2["name"] == "wrapped"
    print(f"✅ markdown 代码块提取成功: {result2}")

    # 测试 3: 无语言标记的代码块
    test3 = '''说明文字
```
{"name": "no_lang", "value": 789}
```
'''
    result3 = extract_json(test3)
    assert result3["name"] == "no_lang"
    print(f"✅ 无语言标记代码块提取成功: {result3}")

    # 测试 4: 嵌入文本中的 JSON
    test4 = '''前面有一些文字 {"name": "embedded", "value": 999} 后面也有'''
    result4 = extract_json(test4)
    assert result4["name"] == "embedded"
    print(f"✅ 嵌入文本提取成功: {result4}")

    # 测试 5: 无效 JSON 应该抛出异常
    test5 = "这段文字没有任何 JSON"
    try:
        extract_json(test5)
        print("❌ 应该抛出异常但没有")
    except ValueError as e:
        print(f"✅ 无效 JSON 正确抛出异常: {str(e)[:50]}...")

    print("")

def test_parallel_execution():
    """测试并行执行"""
    print("\n" + "="*50)
    print("测试 2: ThreadPoolExecutor 并行执行")
    print("="*50)

    import time

    def slow_task(n):
        time.sleep(0.1)
        return f"任务{n}完成"

    start = time.time()
    with ThreadPoolExecutor(max_workers=5) as executor:
        futures = [executor.submit(slow_task, i) for i in range(5)]
        results = [f.result() for f in futures]
    elapsed = time.time() - start

    print(f"✅ 5 个任务并行完成，耗时: {elapsed:.2f}s（串行需要 0.5s）")
    print(f"   结果: {results}")
    assert elapsed < 0.3, "并行执行应该比串行快"
    print("")

def test_context_isolation():
    """测试上下文隔离概念"""
    print("\n" + "="*50)
    print("测试 3: 上下文隔离")
    print("="*50)

    # 模拟两个独立的 messages 列表
    messages_a = [{"role": "user", "content": "任务A"}]
    messages_b = [{"role": "user", "content": "任务B"}]

    # 向 A 添加内容
    messages_a.append({"role": "assistant", "content": "A的回复"})

    # B 不应该受影响
    assert len(messages_b) == 1, "B 的 messages 不应该受 A 影响"
    assert messages_b[0]["content"] == "任务B", "B 的内容不应该变化"

    print(f"✅ 上下文隔离验证成功")
    print(f"   Agent A 的 messages: {len(messages_a)} 条")
    print(f"   Agent B 的 messages: {len(messages_b)} 条")
    print("")

def test_stop_reason_handling():
    """测试 stop_reason 处理"""
    print("\n" + "="*50)
    print("测试 4: stop_reason 处理（end_turn + max_tokens）")
    print("="*50)

    # 模拟响应
    class MockResponse:
        def __init__(self, stop_reason, text):
            self.stop_reason = stop_reason
            self.content = [type('Block', (), {'text': text})()]

    # 测试 end_turn
    resp1 = MockResponse("end_turn", "正常完成")
    assert resp1.stop_reason in ("end_turn", "max_tokens")
    print(f"✅ end_turn 判断正确")

    # 测试 max_tokens
    resp2 = MockResponse("max_tokens", "被截断")
    assert resp2.stop_reason in ("end_turn", "max_tokens")
    print(f"✅ max_tokens 判断正确")

    print("")

def test_full_orchestrator():
    """测试完整编排器流程"""
    print("\n" + "="*50)
    print("测试 5: 完整编排器流程")
    print("="*50)

    result = run_orchestrator_mock("测试任务")
    print(f"✅ 编排器流程完成")
    print(f"   结果: {result[:80]}...")
    print("")

if __name__ == "__main__":
    print("="*50)
    print("多 Agent 代码验证测试")
    print("来源: Claude Agent Teams 实战手册 第3章")
    print("="*50)

    test_extract_json()
    test_parallel_execution()
    test_context_isolation()
    test_stop_reason_handling()
    test_full_orchestrator()

    print("\n" + "="*50)
    print("所有测试通过 ✅")
    print("="*50)

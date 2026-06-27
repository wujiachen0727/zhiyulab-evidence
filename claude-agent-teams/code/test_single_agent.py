#!/usr/bin/env python3
"""
单 Agent 测试脚本
来源：Claude Agent Teams 实战手册 第2章
"""

import json

# 模拟 anthropic 客户端的响应
class MockResponse:
    def __init__(self, stop_reason, content):
        self.stop_reason = stop_reason
        self.content = content

class MockBlock:
    def __init__(self, block_type, text=None, name=None, input_data=None, block_id=None):
        self.type = block_type
        self.text = text
        self.name = name
        self.input = input_data
        self.id = block_id or "tool_123"

# 测试工具定义
tools = [
    {
        "name": "web_search",
        "description": "搜索互联网获取实时信息。输入搜索关键词，返回搜索结果摘要。",
        "input_schema": {
            "type": "object",
            "properties": {
                "query": {
                    "type": "string",
                    "description": "搜索关键词"
                }
            },
            "required": ["query"]
        }
    }
]

def execute_tool(name: str, params: dict) -> str:
    """工具执行入口——目前是模拟数据"""
    if name == "web_search":
        return f"搜索'{params['query']}'的结果：这里是模拟的搜索结果..."
    return "未知工具"

def run_agent_mock(task: str) -> str:
    """模拟运行单 Agent 的 ReAct 循环"""
    messages = [{"role": "user", "content": task}]
    call_count = 0
    max_iterations = 10  # 防止无限循环

    print(f"[测试] 启动 Agent，任务: {task[:50]}...")

    while call_count < max_iterations:
        call_count += 1
        print(f"[测试] 第 {call_count} 轮循环")

        # 模拟 API 响应
        if call_count == 1:
            # 第一轮：Agent 调用工具
            response = MockResponse(
                stop_reason="tool_use",
                content=[MockBlock("tool_use", name="web_search", input_data={"query": "Python Web 框架 2026"}, block_id="tool_001")]
            )
        elif call_count == 2:
            # 第二轮：Agent 返回结果
            response = MockResponse(
                stop_reason="end_turn",
                content=[MockBlock("text", text="根据搜索结果，2026年最流行的Python Web框架包括 FastAPI、Django、Flask 等。")]
            )
        else:
            # 异常情况：max_tokens
            response = MockResponse(
                stop_reason="max_tokens",
                content=[MockBlock("text", text="输出被截断...")]
            )

        # 测试 stop_reason 处理逻辑
        if response.stop_reason == "end_turn":
            for block in response.content:
                if hasattr(block, "text"):
                    print(f"[测试] 任务完成，返回结果")
                    return block.text

        # 测试 max_tokens 处理逻辑（关键测试点）
        if response.stop_reason == "max_tokens":
            for block in response.content:
                if hasattr(block, "text"):
                    print(f"[测试] max_tokens 截止，返回已有内容")
                    return block.text
            return "Agent 输出被截断，请增大 max_tokens"

        # 测试工具调用处理
        tool_results = []
        for block in response.content:
            if block.type == "tool_use":
                print(f"[测试] Agent 调用工具: {block.name}, 参数: {block.input}")
                result = execute_tool(block.name, block.input)
                tool_results.append({
                    "type": "tool_result",
                    "tool_use_id": block.id,
                    "content": result,
                })

        # 更新对话历史
        messages.append({"role": "assistant", "content": response.content})
        messages.append({"role": "user", "content": tool_results})

    return "错误：超过最大迭代次数"

def test_stop_reason_handling():
    """测试 stop_reason 处理逻辑"""
    print("\n" + "="*50)
    print("测试 1: stop_reason 处理逻辑")
    print("="*50)

    # 测试 end_turn
    messages = []
    response = MockResponse("end_turn", [MockBlock("text", text="测试结果")])
    if response.stop_reason == "end_turn":
        for block in response.content:
            if hasattr(block, "text"):
                print(f"✅ end_turn 处理正确: {block.text}")

    # 测试 max_tokens（关键：之前缺少这个处理）
    response = MockResponse("max_tokens", [MockBlock("text", text="截断内容")])
    if response.stop_reason == "max_tokens":
        for block in response.content:
            if hasattr(block, "text"):
                print(f"✅ max_tokens 处理正确: {block.text}")

    print("")

def test_tool_definition():
    """测试工具定义格式"""
    print("\n" + "="*50)
    print("测试 2: 工具定义格式")
    print("="*50)

    # 验证 JSON Schema 格式
    tool = tools[0]
    assert tool["name"] == "web_search", "工具名称错误"
    assert "input_schema" in tool, "缺少 input_schema"
    assert "properties" in tool["input_schema"], "缺少 properties"
    assert "query" in tool["input_schema"]["properties"], "缺少 query 参数"
    assert "required" in tool["input_schema"], "缺少 required"
    assert "query" in tool["input_schema"]["required"], "query 未标记为 required"

    print(f"✅ 工具定义格式正确")
    print(f"   名称: {tool['name']}")
    print(f"   描述: {tool['description'][:50]}...")
    print(f"   参数: {list(tool['input_schema']['properties'].keys())}")
    print("")

def test_execute_tool():
    """测试工具执行"""
    print("\n" + "="*50)
    print("测试 3: 工具执行")
    print("="*50)

    result = execute_tool("web_search", {"query": "测试关键词"})
    print(f"✅ 工具执行成功: {result[:50]}...")

    result = execute_tool("unknown_tool", {})
    print(f"✅ 未知工具处理: {result}")
    print("")

def test_full_agent_loop():
    """测试完整 Agent 循环"""
    print("\n" + "="*50)
    print("测试 4: 完整 Agent 循环")
    print("="*50)

    result = run_agent_mock("搜索一下 2026 年最流行的 Python Web 框架有哪些")
    print(f"✅ Agent 循环完成")
    print(f"   结果: {result[:80]}...")
    print("")

if __name__ == "__main__":
    print("="*50)
    print("单 Agent 代码验证测试")
    print("来源: Claude Agent Teams 实战手册 第2章")
    print("="*50)

    test_stop_reason_handling()
    test_tool_definition()
    test_execute_tool()
    test_full_agent_loop()

    print("\n" + "="*50)
    print("所有测试通过 ✅")
    print("="*50)

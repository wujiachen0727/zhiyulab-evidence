#!/usr/bin/env python3
"""
竞品调研系统测试脚本
来源：Claude Agent Teams 实战手册 第4章
"""

import json
import re
from concurrent.futures import ThreadPoolExecutor

# ============ 核心代码（从文章中提取） ============

search_tool = {
    "name": "web_search",
    "description": "搜索互联网获取产品信息、评价和定价数据。输入搜索关键词，返回相关结果摘要。",
    "input_schema": {
        "type": "object",
        "properties": {
            "query": {"type": "string", "description": "搜索关键词"}
        },
        "required": ["query"]
    }
}

def execute_tool(name: str, params: dict) -> str:
    """工具执行入口——模拟数据"""
    if name == "web_search":
        # 模拟返回竞品信息
        mock_data = {
            "Notion竞品": ["Evernote", "Obsidian", "Roam Research", "Craft", "Logseq"],
            "Evernote": "笔记应用，支持多平台同步，免费版有限制",
            "Obsidian": "本地优先的知识管理工具，支持双向链接",
            "Roam Research": "大纲式笔记工具，强调知识图谱",
        }
        query = params['query']
        for key, value in mock_data.items():
            if key in query or query in key:
                if isinstance(value, list):
                    return json.dumps({"competitors": value})
                return value
        return f"搜索结果：{query} 的相关信息..."
    return "未知工具"


def create_subagent(role: str, task: str, tools: list = None) -> str:
    """模拟创建子 Agent"""
    print(f"[子Agent] 角色: {role[:40]}...")
    print(f"[子Agent] 任务: {task[:50]}...")

    # 模拟不同角色的返回
    if "竞品搜索" in role:
        return json.dumps({"competitors": ["Evernote", "Obsidian", "Roam Research"]})
    elif "产品分析" in role:
        return f"产品分析结果：{task} - 功能完善，定价合理，用户评价积极"
    else:
        return f"完成: {task}"


def extract_json(text: str) -> dict:
    """从 LLM 输出中提取 JSON（处理 markdown 代码块等干扰）"""
    try:
        return json.loads(text)
    except json.JSONDecodeError:
        pass
    match = re.search(r'```(?:json)?\s*([\s\S]*?)```', text)
    if match:
        try:
            return json.loads(match.group(1).strip())
        except json.JSONDecodeError:
            pass
    match = re.search(r'\{[\s\S]*\}', text)
    if match:
        try:
            return json.loads(match.group(0))
        except json.JSONDecodeError:
            pass
    raise ValueError(f"无法提取 JSON: {text[:200]}")


def research_competitors(product_name: str) -> str:
    """竞品调研完整流程"""

    # 第一步：搜索 Agent 找出主要竞品
    competitors_raw = create_subagent(
        role="你是竞品搜索专家。找出给定产品的 3-5 个主要竞品。返回纯 JSON，格式：{\"competitors\": [\"竞品1\", \"竞品2\", ...]}",
        task=f"找出 {product_name} 的主要竞品",
        tools=[search_tool],
    )

    # 解析竞品列表——LLM 返回不一定是纯 JSON，需要鲁棒解析
    try:
        competitors = extract_json(competitors_raw)["competitors"]
    except (ValueError, KeyError):
        competitors = ["竞品A", "竞品B", "竞品C"]

    print(f"[调研] 找到竞品: {competitors}")

    # 第二步：并行收集每个竞品的详细信息
    details = {}
    with ThreadPoolExecutor(max_workers=len(competitors)) as executor:
        futures = {}
        for comp in competitors:
            future = executor.submit(
                create_subagent,
                role="你是产品分析专家。搜索并整理产品的功能亮点、定价方案和用户评价。输出结构化的分析结果。",
                task=f"详细分析产品：{comp}",
                tools=[search_tool],
            )
            futures[comp] = future

        for comp, future in futures.items():
            details[comp] = future.result()

    # 第三步：汇总 Agent 生成对比报告
    report = create_subagent(
        role="你是商业分析报告专家。把多个产品的分析结果整合成一份结构清晰的竞品对比报告。",
        task=f"""原始产品：{product_name}

各竞品分析结果：
{json.dumps(details, ensure_ascii=False, indent=2)}

请生成竞品对比报告，包含：
1. 竞品概览（一句话定位）
2. 功能对比矩阵
3. 定价对比
4. 各产品优劣势
5. 选型建议""",
    )

    return report


# ============ 测试用例 ============

def test_research_flow():
    """测试完整调研流程"""
    print("\n" + "="*50)
    print("测试 1: 完整竞品调研流程")
    print("="*50)

    report = research_competitors("Notion")
    print(f"\n✅ 调研流程完成")
    print(f"   报告: {report[:100]}...")
    print("")

def test_json_extraction_edge_cases():
    """测试 JSON 提取边界情况"""
    print("\n" + "="*50)
    print("测试 2: JSON 提取边界情况")
    print("="*50)

    # 测试带说明文字的 JSON
    test_cases = [
        ('好的，以下是竞品列表：{"competitors": ["A", "B", "C"]}', ["A", "B", "C"]),
        ('```json\n{"competitors": ["X", "Y"]}\n```', ["X", "Y"]),
        ('{"competitors": ["Only"]}', ["Only"]),
    ]

    for text, expected in test_cases:
        result = extract_json(text)
        assert result["competitors"] == expected, f"预期 {expected}，实际 {result['competitors']}"
        print(f"✅ 提取成功: {text[:40]}... → {expected}")

    print("")

def test_fallback_on_parse_failure():
    """测试解析失败时的降级处理"""
    print("\n" + "="*50)
    print("测试 3: 解析失败降级处理")
    print("="*50)

    # 模拟解析失败
    invalid_json = "这不是有效的 JSON"

    try:
        competitors = extract_json(invalid_json)["competitors"]
    except (ValueError, KeyError):
        competitors = ["竞品A", "竞品B", "竞品C"]
        print(f"✅ 降级处理成功，使用默认值: {competitors}")

    assert competitors == ["竞品A", "竞品B", "竞品C"]
    print("")

def test_tool_definition():
    """测试搜索工具定义"""
    print("\n" + "="*50)
    print("测试 4: 搜索工具定义")
    print("="*50)

    assert search_tool["name"] == "web_search"
    assert "description" in search_tool
    assert len(search_tool["description"]) > 20, "描述太短，Agent 可能不知道如何使用"
    assert "input_schema" in search_tool
    assert "required" in search_tool["input_schema"]

    print(f"✅ 工具定义验证通过")
    print(f"   名称: {search_tool['name']}")
    print(f"   描述长度: {len(search_tool['description'])} 字符")
    print(f"   必填参数: {search_tool['input_schema']['required']}")
    print("")

def test_execute_tool():
    """测试工具执行"""
    print("\n" + "="*50)
    print("测试 5: 工具执行")
    print("="*50)

    # 测试竞品搜索
    result = execute_tool("web_search", {"query": "Notion竞品"})
    data = json.loads(result)
    assert "competitors" in data
    print(f"✅ 竞品搜索成功: {data['competitors']}")

    # 测试产品信息
    result = execute_tool("web_search", {"query": "Evernote"})
    print(f"✅ 产品信息搜索: {result[:50]}...")

    print("")

if __name__ == "__main__":
    print("="*50)
    print("竞品调研系统代码验证测试")
    print("来源: Claude Agent Teams 实战手册 第4章")
    print("="*50)

    test_tool_definition()
    test_execute_tool()
    test_json_extraction_edge_cases()
    test_fallback_on_parse_failure()
    test_research_flow()

    print("\n" + "="*50)
    print("所有测试通过 ✅")
    print("="*50)

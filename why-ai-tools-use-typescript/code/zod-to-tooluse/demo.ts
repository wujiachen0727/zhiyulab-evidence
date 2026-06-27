// 论据 E1：Zod → JSON Schema → LLM tool_use 的自动推导链
// 展示从 TypeScript 类型声明到 AI 可用的 tool definition 的完整路径

import { z } from "zod";
import { zodToJsonSchema } from "zod-to-json-schema";

// ===== Step 1: 开发者写 Zod schema（TypeScript 类型声明）=====

const WeatherInput = z.object({
  city: z.string().describe("城市名称，如 'Beijing'"),
  unit: z.enum(["celsius", "fahrenheit"]).default("celsius").describe("温度单位"),
  forecast_days: z.number().min(1).max(7).optional().describe("预报天数，1-7"),
});

// ===== Step 2: Zod 自动转为 JSON Schema =====

const jsonSchema = zodToJsonSchema(WeatherInput, {
  name: "get_weather",
  $refStrategy: "none",
});

console.log("=== Step 2: JSON Schema（自动生成）===");
console.log(JSON.stringify(jsonSchema, null, 2));

// ===== Step 3: JSON Schema 直接作为 LLM tool_use 的 input_schema =====
// 这就是发送给 Claude/GPT 的 tool definition

const toolDefinition = {
  name: "get_weather",
  description: "获取指定城市的天气预报",
  input_schema: {
    type: "object",
    properties: {
      city: { type: "string", description: "城市名称，如 'Beijing'" },
      unit: {
        type: "string",
        enum: ["celsius", "fahrenheit"],
        default: "celsius",
        description: "温度单位",
      },
      forecast_days: {
        type: "number",
        minimum: 1,
        maximum: 7,
        description: "预报天数，1-7",
      },
    },
    required: ["city"],
  },
};

console.log("\n=== Step 3: LLM Tool Definition（机器可读）===");
console.log(JSON.stringify(toolDefinition, null, 2));

// ===== 关键洞察 =====
//
// 整条链路：
// TypeScript 类型 → Zod schema → JSON Schema → LLM tool_use input_schema
//
// 这意味着：
// 1. 开发者写的类型声明，AI 能直接"读懂"——因为 JSON Schema 就是 AI 的"接口规格书"
// 2. 类型约束（string/number/enum/min/max）= 告诉 AI "这个参数只能是什么"
// 3. .describe() = 告诉 AI "这个参数是干什么的"
//
// 在 Python 中，同样的信息分散在：
// - 函数签名（参数名）
// - docstring（描述，格式不统一）
// - 运行时检查（类型约束，AI 看不到）
//
// TypeScript/Zod 把这三者统一到一个机器可解析的结构中。
// AI 不需要"理解"代码逻辑，只需要解析结构化的 schema。

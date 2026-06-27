# E3: Go 标准库 reflect 使用频率统计

## 环境

- Go 1.26.2
- 统计对象：$GOROOT/src/（不含 cmd/、testdata/、_test.go 文件）
- 统计日期：2026-05-26

## 核心数据

- 标准库总包数（不含 cmd/）：383
- 使用 reflect 的包数：36
- **占比：9.3%**

## 使用分布（按顶层目录）

| 领域 | 包数 | 说明 |
|------|:----:|------|
| encoding | 7 | json/xml/gob/binary/asn1 — 序列化天然需要 |
| internal | 7 | 内部工具包 |
| testing | 4 | 测试框架辅助 |
| net | 3 | http/rpc |
| database | 2 | sql |
| go | 2 | ast/doc |
| 其他 | 11 | fmt/flag/html/text/log 等 |

## 关键发现

1. **90% 的标准库包不用 reflect** — Go 团队自己在"避免"自己设计的 API
2. **使用集中在"必须动态处理类型"的场景**：序列化、格式化输出、SQL 绑定
3. **没有一个业务逻辑包使用 reflect** — 纯粹的基础设施工具
4. **Go 1.18 泛型发布后**：新增的标准库包（如 slices、maps）全部用泛型实现，零 reflect

## 正文可引用的表述

- "Go 标准库 383 个包中，只有 36 个（9.3%）使用了 reflect——Go 团队自己也在'认知税'面前选择绕道"
- "reflect 集中在 encoding/json、fmt.Printf 这类'必须处理未知类型'的基础设施中——不是你的业务代码该碰的"
- "Go 1.18 之后新增的标准库包（slices、maps、cmp）全部用泛型实现，0 个用 reflect"

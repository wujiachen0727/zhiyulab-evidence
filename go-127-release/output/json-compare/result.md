# E2 json 1.26 vs 1.27 对照

环境：darwin/arm64，Apple M4 Pro；1.26 测试在 `/tmp/json126`（go.mod `go 1.26`）避免 evidence 目录 `.go-version` 干扰。

## 行为测试

| 用例 | 1.26.4 | 1.27.0 |
|------|--------|--------|
| 重复 key `{"a":1,"a":2}` | 后者覆盖，pass | 同左 |
| 非法 UTF-8 字符串 | 报错 | 同左 |
| 类型不匹配 `{"n":"x"}` → int | 报错 | 同左 |

类型不匹配错误文案（本用例）**相同**：
`json: cannot unmarshal string into Go struct field .n of type int`

> 官方说明：v2 托底后错误文案**可能**变化；本用例未变，但不应假设所有错误字符串不变。

## BenchmarkUnmarshalLarge（小 payload）

| 版本 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| 1.26.4 | ~1050 | 1176 | 28 |
| 1.27.0 | ~1210 | 1088 | 27 |

解读：allocs/op 与堆字节下降，符合 v2 引擎换实现；本机小对象 bench 的 ns/op 1.27 略慢，**不得**外推为「全面变慢/变快」——官方强调的是典型 unmarshal 路径与实现差异。

代码：`evidence/code/json-compare/main_test.go`

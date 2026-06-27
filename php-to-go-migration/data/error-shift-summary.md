# 错误暴露位置实验摘要

## 实测环境

- Go：go1.26.2 darwin/arm64
- PHP：PHP 8.4.21 cli（Docker 镜像 `php:8.4-cli`）
- 原始输出：`../output/error-boundary-compare/result.txt`
- 代码目录：`../code/error-boundary-compare/`

## 输入

```json
{"user_id":"42","email":123}
```

## 结果对比

| 路径 | 结果 | 错误/转换发生位置 | 可用于正文的判断 |
|------|------|------------------|------------------|
| PHP 默认弱类型 | `user_id` 从字符串转为 integer，`email` 从数字转为 string | 函数参数调用时发生隐式类型转换 | 默认弱类型路径会让一部分输入问题继续向后流动 |
| PHP `strict_types=1` | `TypeError: Argument #1 ($userId) must be of type int, string given` | 函数调用边界 | 现代 PHP 也能通过 strict_types 把部分错误前移 |
| Go JSON struct decode | `json.UnmarshalTypeError`，无法把字符串解码到 int 字段 | JSON 解码边界 | Go 的强类型 struct 会在输入进入业务逻辑前暴露类型不匹配 |

## 推导结论

这个实验不支持“PHP 做不到，Go 做得到”的粗暴判断。更准确的结论是：

1. PHP 默认弱类型路径会进行一部分隐式转换，边界问题可能更晚暴露。
2. PHP 开启 `strict_types=1` 后，也能把一部分类型问题前移到函数调用边界。
3. Go 在把 JSON 解码到强类型 struct 时，会更早暴露类型不匹配，让错误停在业务逻辑之前。

所以本文应写成“默认工程习惯和复杂度归属差异”，而不是“语言能力优劣”。

## 正文呈现建议

正文中可以使用三行结果，不必展开全部代码：

```text
[PHP weak types] user_id: "42" -> integer 42, email: 123 -> string "123"
[PHP strict_types=1] TypeError: string given, int expected
[Go JSON struct decode] json.UnmarshalTypeError: cannot unmarshal string into int
```

接下来的判断必须克制：Go 让这类错误更早被看见，但 PHP 也有办法前移错误；迁移真正改变的是团队默认把复杂度放在哪里。

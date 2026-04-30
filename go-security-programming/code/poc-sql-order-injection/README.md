# E2：类型安全幻觉 — ORDER BY 注入 PoC

## 核心假设

> Go 的强类型系统能防住 SQL 注入。

## 证伪结论

**错。**类型安全只保证"变量是 string"，不保证"拼接后的 SQL 结构安全"。`database/sql` 不支持列名/排序方向作为 `?` 参数，开发者必须手动做白名单——这恰好是最容易漏的地方。

## 运行方式

```bash
cd evidence/code/poc-sql-order-injection
go mod init poc-sql 2>/dev/null
go get modernc.org/sqlite@latest
go run main.go
```

## 预期输出

```
=== 正常查询 ===
返回: [alice admin]

=== 类型安全幻觉：OrderBy 字段注入 ===
返回（已按密码内容排序，密码首字母 'r' 的排前面）: [admin alice]
（或返回 sqlite 错误，本身就是信息泄漏信号）
```

## 环境

- Go 1.26.2 darwin/arm64（实测）
- modernc.org/sqlite（纯 Go SQLite，无 CGO 依赖）

## 对应文章段落

**第三章 第 1 个幻觉：强类型不等于强安全**

## 论据要点

1. Go 编译器认 `OrderBy string` 是合法类型，不报错
2. 业务代码在 WHERE 用了 `?`，开发者以为"我做了参数化"
3. ORDER BY 必须拼接（SQL 标准不支持列名参数化），拼接点就是注入点
4. 类型系统对此无能为力——需要白名单防御

# Evidence 目录

Go 1.27 对照实验，计划公开至 zhiyulab-evidence。

| 目录 | 论据 | 说明 |
|------|------|------|
| code/json-compare | E2 | 1.27 行为测试；1.26 对照见 output |
| code/generic-method | E3,E4 | 接口失败 + reflect |
| code/malloc-bench | E5 | heap alloc bench |
| code/goroutineleak | E6 | local vs global channel |
| output/*/result.md | — | 运行结果摘要 |

**复现注意**：evidence 目录 pin `go1.27.0`；1.26 对照请在目录外或 `/tmp` 执行并改 go.mod 为 `go 1.26`。

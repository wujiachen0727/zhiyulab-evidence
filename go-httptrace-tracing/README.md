# 论据总索引

## 独立论据

| # | 类型 | 描述 | 路径 | 状态 |
|---|------|------|------|:----:|
| E1 | 实验验证 | httptrace CLI 工具源码 | `code/httptrace-cli/main.go` | ✅ 已验证 |
| E1 | 实验验证 | GitHub 实测输出 | `output/github-httptrace.txt` | ✅ |
| E1 | 实验验证 | Baidu 实测输出 | `output/baidu-httptrace.txt` | ✅ |
| E1 | 实验验证 | Go 官方文档实测输出 | `output/godoc-httptrace.txt` | ✅ |
| E2 | 场景模拟 | 连接复用/新建对比 | `code/connection-reuse/main.go` | ✅ 已验证 |
| E3 | 数据实测 | 多站点阶段耗时对比 | `data/http-phases-benchmark.md` | ✅ |

## 表达手法

| # | 类型 | 描述 | 说明 |
|---|------|------|------|
| E6 | 类比 | 连接池耗尽类比 | 辅助连接复用章节理解 |

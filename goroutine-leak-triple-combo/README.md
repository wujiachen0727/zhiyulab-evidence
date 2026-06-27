# 论据产物索引

## 实验代码

| 目录 | 描述 | 产出 |
|------|------|------|
| `code/leak-repro/` | 最小复现：无超时 client + 慢 server + 无 context | goroutine 数量持续增长（阻塞在 http.Get） |
| `code/leak-fix/` | 三重修复对比：client 超时 / context cancel / 双保险 | 三种修复均遏制泄漏，双保险最可靠 |

## 数据

| 文件 | 描述 |
|------|------|
| `data/goroutine-resource-estimate.md` | 10 万 goroutine 资源消耗推演（基于 Go runtime 文档） |

## 场景模拟（嵌入正文）

- E6：看似正确但 context 无效的代码片段（3 种常见写法）
- 直接嵌入文章第 2、5 章，不单独存文件

## 论据统计

- 独立论据：7 条（自造 7 + 外部引用 0-2）
- 预估自造度：78-87%
- 表达手法：3 条（类比、修辞问句、并置对比）

# 数据竞争演示代码

> 本文实验代码——用 Go 1.26.4 编写，依赖标准库。

## 依赖

- Go 1.26+（使用 `go run -race main.go` 运行）
- 无第三方依赖

## 运行方式

```bash
cd evidence/code/{实验名}/
go run -race main.go
```

## 子目录说明

| 目录 | 用途 | 关键结论 |
|------|------|---------|
| `data-race-demo/` | 多 goroutine 并发 append → data race | append 不是原子的 |
| `ownership-model-demo/` | 无所有权 vs 有所有权对比 | 同一需求，所有权模型 = 0 data race |
| `ownership-transfer-demo/` | channel 传切片的所有权转移 | 发送后不再碰 = 安全 |
| `pipeline-demo/` | pipeline 模式对比 | 链式传递所有权 = 0 data race |

## 数据引用

所有实验均使用 `go run -race` 验证，输出保存在 `evidence/output/` 下。

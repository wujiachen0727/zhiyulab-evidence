# Go 网络编程实验代码

**配套文章**：《别只会写 net.Listen：Go 网络编程的三层进阶》

## 实验列表

| 目录 | 实验 | 说明 |
|------|------|------|
| `goroutine-pressure/` | goroutine-per-conn 内存压力测试 | 100/1K/5K/10K 连接的内存开销实测 |
| `tcp-sticky/` | TCP 粘包复现 | 5 条消息粘包可视化 |
| `framing-bench/` | 分帧策略 benchmark | length-prefix vs delimiter vs fixed-length 性能对比 |
| `socket-options/` | TCP_NODELAY 效果对比 | 小包/中包/大包场景的延迟实测 |

## 环境

- Go 1.26.2
- darwin/arm64 (Apple M4 Pro)

## 运行

```bash
# goroutine 内存压力测试
cd goroutine-pressure && go run main.go

# TCP 粘包复现
cd tcp-sticky && go run main.go

# 分帧策略 benchmark
cd framing-bench && go test -bench=. -benchmem

# TCP_NODELAY 对比
cd socket-options && go run main.go
```

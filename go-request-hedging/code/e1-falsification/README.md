# E1 证伪实验：hedging 降 P99 但锁竞争指标不变

## 实验目的

证伪/支撑核心假设：**hedging 能降低 P99，但不解决病因（锁竞争）——病因依然存在。**

## 实验设计

### 服务端

- 启动一个 HTTP 服务，处理函数 `handler` 会获取一个全局 mutex
- 在持有 mutex 期间，模拟慢操作（`time.Sleep` 随机 10-50ms，10% 概率 200-500ms 模拟长尾）
- 通过 `runtime/pprof` 暴露 mutex profile，采集 `mutex_contention_total` 和平均 mutex 等待时间

### 客户端

- 场景 A：普通请求（无 hedging）
- 场景 B：hedging 请求（P95 延迟后发第二个请求，取先返回的）
- 两个场景各发 2000 个请求，记录 P50/P95/P99 延迟
- 同时采集服务端的 mutex profile，对比两个场景的锁竞争指标

### 关键观测点

- **P99**：hedging 场景应该明显下降
- **mutex 等待时间**：如果假设成立，两个场景应该接近（hedging 没有减少锁竞争）
- **吞吐量**：hedging 场景可能因为额外请求而下降

## 运行方法

```bash
# 服务端（在另一个终端）
go run server/main.go

# 客户端
go run client/main.go
```

## 环境

- Go 1.26.4 darwin/arm64
- macOS（本地开发机）

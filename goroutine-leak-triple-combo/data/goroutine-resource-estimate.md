# 10 万 goroutine 资源消耗推演

> 基于 Go runtime 文档和工程经验推演，非精确实测。

## goroutine 栈空间

- Go goroutine 初始栈：**2-8 KB**（Go 1.4+ 默认 2KB 初始栈，可动态增长）
- 10 万 goroutine 栈空间：10万 × 2KB = **~200MB**（最小值）
- 如果栈增长到 8KB：10万 × 8KB = **~800MB**

## 连接资源

- 每个阻塞的 HTTP 请求占用一个 TCP 连接
- 10 万 TCP 连接 = 10 万 file descriptor
- Linux 默认 ulimit -n = 1024（需要调大）
- 实际上在 fd 耗尽之前就会触发 `dial tcp: socket: too many open files`

## 内存

- 除栈空间外，每个 HTTP 请求还占用：
  - http.Request 结构体
  - DNS 解析缓存
  - TLS 状态（如果是 HTTPS）
- 综合估算：**10 万 goroutine ≈ 500MB-1.5GB 内存**

## 关键结论

10 万 goroutine 本身不会让 Go runtime 崩溃（Go 可以轻松处理百万级 goroutine），但：
1. **fd 耗尽**先于 OOM 发生——`too many open files` 会让新请求失败
2. **内存泄漏是渐进的**——不是突然崩，而是缓慢恶化
3. **连锁反应**：fd 耗尽 → 新请求失败 → 上游健康检查失败 → 被踢出负载均衡 → 服务不可用

这就是为什么 goroutine 泄漏被称为"慢性病"——你不会立刻注意到，但最终会致命。

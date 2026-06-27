# E6：经验落地——SSE 生产环境陷阱与解决方案

> **类型**：经验落地
> **优先级**：5
> **关联章节**：第 6 章（生产陷阱与救火指南）
> **来源**：用户经验 + 已知工程实践

---

## 陷阱 1：Nginx 缓冲

**问题**：Nginx 默认缓冲响应体，SSE 的流式数据被缓存后才一次性推送给客户端 → 前端收不到实时推送

**表现**：EventSource 连接正常，但数据"卡住"，一段时间后一次性收到大量数据

**解法**：
```
# Nginx 配置
proxy_buffering off;
# 或者通过响应头控制
X-Accel-Buffering: no
```

**注意**：`X-Accel-Buffering: no` 是 Nginx 特有的响应头，需要在服务端代码中设置。

## 陷阱 2：反向代理超时

**问题**：SSE 连接长期不关闭，反向代理（ALB、Cloudflare、Nginx）有默认超时

**常见超时配置**：
- AWS ALB：默认 60s 空闲超时
- Cloudflare：默认 100s
- Nginx `proxy_read_timeout`：默认 60s

**解法**：
- ALB：将空闲超时设置为更大的值（如 3600s）或关闭
- Nginx：`proxy_read_timeout 3600s;`
- Cloudflare：免费版无法调整，需要企业版或使用 `Cloudflare Workers` 做代理
- **应用层心跳**：每 30-60s 发一条 `: keepalive\n\n` 注释行，刷新超时计数器

## 陷阱 3：浏览器连接数限制

**问题**：HTTP/1.1 下，每个域名最多 6 个并发 SSE 连接

**影响**：如果一个页面同时打开多个 SSE 连接，超过 6 个后新的连接会被阻塞

**解法**：
- 使用 HTTP/2（多路复用后无此限制）
- 合并多个推送通道为单个 SSE 连接（用事件类型区分）
- 如果必须使用 HTTP/1.1，控制连接数 ≤ 5

## 陷阱 4：EventSource 默认行为

**问题**：EventSource 在连接断开时会自动重连，但如果服务端返回 4xx/5xx，也会重连

**表现**：服务端宕机后恢复，大量客户端同时重连 → 羊群效应

**解法**：
- 服务端实现指数退避（通过 `Last-Event-ID` 头判断）
- 客户端可以设置 `withCredentials` 控制 CORS 行为
- 如果使用 fetch + ReadableStream 替代 EventSource，可以精细控制重连策略

## 陷阱 5：事件格式错误

**问题**：SSE 协议要求严格的事件格式，常见错误包括：
- 缺少 `data:` 前缀
- 缺少双换行符 `\n\n` 分隔事件
- 多行数据未正确编码

**正确格式**：
```
data: {"message": "hello"}\n\n
```

**多行数据**：
```
data: {\n
data: "message": "hello",\n
data: "id": 1\n
data: }\n\n
```

## 陷阱 6：进程退出导致连接中断

**问题**：部署新版本时，旧进程退出 → 所有 SSE 连接中断 → 客户端全部重连

**解法**：
- 优雅关闭（graceful shutdown）：在进程退出前，等待正在处理的 SSE 连接完成
- 使用连接池管理 + 信号量控制
- 配合负载均衡的 draining 机制

## 生产环境 SSE 部署 Checklist

- [ ] Nginx 关闭 `proxy_buffering` 或设置 `X-Accel-Buffering: no`
- [ ] 反向代理超时设置为 ≥ 3600s 或关闭
- [ ] 应用层心跳（每 30-60s 发 `: keepalive`）
- [ ] 使用 HTTP/2 避免连接数限制
- [ ] 优雅关闭处理
- [ ] 监控连接数，异常增长及时告警
- [ ] 考虑使用 fetch + ReadableStream 替代 EventSource（更灵活）

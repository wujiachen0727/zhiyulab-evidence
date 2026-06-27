# 场景模拟：线上排查路径

> 基于工程常识的合理推演，虚构人物和公司名已标注

## 背景

某日 02:03，告警系统触发 P0：核心订单查询接口超时率 23%。

## 第一轮排查：直觉判断

值班同学小李第一反应："是不是数据库挂了？"

- 检查数据库：连接正常，慢查询无异常
- 检查 Redis：响应正常
- 检查 CPU/内存：均处于正常水位

**直觉排除**：不是基础设施问题。

## 第二轮排查：为什么健康检查过了？

"奇怪——健康检查是通的，/health 返回 200，但业务接口全部超时。"

- 查看健康检查代码：`/health` 只是 `fmt.Fprintf(w, "ok")` — 不需要任何锁
- 业务接口的路径：`Handler → Service.GetOrders → Cache.GetData → RWMutex.RLock`
- **问题不在进程级，在 goroutine 级**

## 第三轮排查：拉 goroutine dump

```bash
# 方式一：pprof（推荐，生产环境最常用）
curl http://localhost:6060/debug/pprof/goroutine?debug=2 -o dump.txt

# 方式二：SIGQUIT（需要进程权限）
kill -QUIT <pid>
```

dump 内容关键片段（虚构线上场景）：

```
goroutine 12043 [sync.RWMutex.RLock]:
sync.runtime_SemacquireRWMutexR(0xc00012a038)
    /usr/local/go/src/runtime/sema.go:100
sync.(*RWMutex).RLock(...)
    /usr/local/go/src/sync/rwmutex.go:74
main.(*Cache).GetData(0xc00012a000, 0xc0004a2000, 0x20)
    /app/cache.go:25
main.(*OrderService).GetOrders(0xc00012a040, 0xc0004a1f00)
    /app/service.go:42
...
```

**信号特征**：
- ✅ 大批 goroutine 卡在 `sync.RWMutex.RLock`
- ✅ 入口是 `runtime_SemacquireRWMutexR`
- ✅ 堆栈统一：`GetOrders → GetData → RLock`

## 第四轮排查：是谁在等写锁？

继续往下看 dump：

```
goroutine 1 [sync.RWMutex.Lock]:
sync.runtime_SemacquireRWMutex(0xc00012a038)
    /usr/local/go/src/runtime/sema.go:105
sync.(*RWMutex).Lock(...)
    /usr/local/go/src/sync/rwmutex.go:155
main.(*Cache).RefreshCache(...)
    /app/cache.go:50
```

**信号特征**：
- ✅ 一个 goroutine 卡在 `sync.RWMutex.Lock`
- ✅ 是 `RefreshCache` 方法在等待写锁

**问题还原**：
1. 某个长时间运行的请求持有了 RLock（读锁）
2. RefreshCache 启动，尝试获取写锁 — 等待
3. 新的读请求到达，尝试获取 RLock — **被阻塞！**（writer-preference）
4. 所有读请求排队等 — 连锁阻塞

## 修复方案

```go
// 问题：RefreshCache 用了 Lock，导致 writer-preference 阻塞所有 reader
// 方案一：将 RWMutex 改为 Mutex（如果读远多于写，这不是最优解）
// 方案二：加一个超时兜底
func (c *Cache) GetData(ctx context.Context, key string) (string, bool) {
    done := make(chan struct{})
    var val string
    var ok bool
    go func() {
        c.mu.RLock()
        defer c.mu.RUnlock()
        val, ok = c.data[key]
        close(done)
    }()
    select {
    case <-done:
        return val, ok
    case <-ctx.Done():
        return "", false
    }
}
```

## 排查速查（三种模式）

| 排查信号 | 想到什么 | 怎么验证 |
|---------|---------|---------|
| `sync.RWMutex.RLock` 批量出现 | RWMutex 递归读阻塞 | 找持有写锁的 goroutine |
| `chan send` | channel send 接收方退出 | 找 channel 的另一端是否还在 |
| `chan receive` + context | context 链断裂 | 检查 goroutine 是否监听了正确的 ctx.Done() |

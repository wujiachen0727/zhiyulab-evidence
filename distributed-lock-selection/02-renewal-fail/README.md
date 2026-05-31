# E2：续期失败 / GC pause 双持

## 复现什么
watchdog 续期机制能解决"业务比 TTL 长"的常规场景，但当 STW（GC pause、虚拟机挂起、OS 调度延迟）超过 TTL 时，watchdog 跟业务在同一进程里，**一起被冻结**——锁过期被自动释放，他人抢走。

A 苏醒后如果不做 owner 校验，会以为自己还持锁，扣库存/写数据 → 双持事故。

## 前置环境
- 本地 6379 端口有 Redis（同 E1）

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code
go run ./02-renewal-fail
```

## 关键参数
- `ttl = 3s`
- `renewInterval = 1s`
- `stwDuration = 5s`（>TTL，必触发过期）

## 关键证据
- A 在 T+0 拿到锁，watchdog 续期 2 次（T+1s、T+2s）
- T+2s 进入 STW，watchdog 一起停
- TTL 过期后 B 抢锁成功
- A 苏醒后 GET → owner 已是 B

## 工程修复方向（写进正文）
1. **owner 校验**：用 Lua 脚本做 `GET+DEL` 原子比较，避免误删别人的锁
2. **fencing token**（Kleppmann 推荐）：每次获锁拿单调递增 token，存储层拒绝旧 token
3. **改用 etcd lease**：lease 在客户端断连时由 server 主动 revoke，不依赖客户端续期

## 与 E1 的关系
E1 复现的是"持锁人崩溃"的成本（B 必须等到 TTL）；
E2 复现的是"持锁人虽然没崩溃但被 STW 冻结"导致**双持**——更严重，因为 A 还在跑。

# E4：etcd lease + Txn(CAS) 锁

## 演示什么
1. **Lease keepalive**：客户端在线时 lease 不过期；客户端断连/崩溃后 server 主动 revoke
2. **Txn(CAS) 获锁**：只在 key 不存在时写入，原子保证互斥
3. **Watch + Delete 事件**：锁释放的瞬间所有 watcher 立即被通知，无需轮询

这三件事正好对应 Redis SETNX 锁的三个失败场景的解药。

## 前置环境
```bash
docker run -d --name dl-etcd -p 2379:2379 -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.18 \
  /usr/local/bin/etcd \
  --name=dl-etcd \
  --listen-client-urls=http://0.0.0.0:2379 \
  --advertise-client-urls=http://0.0.0.0:2379 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --initial-advertise-peer-urls=http://0.0.0.0:2380 \
  --initial-cluster=dl-etcd=http://0.0.0.0:2380
```

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code
go run ./04-etcd-lock
```

## 关键代码片段

```go
// 1. Grant lease（带 TTL 的租约）
leaseA, _ := cli.Grant(ctx, 10)

// 2. KeepAlive（客户端在线 → lease 不过期）
keepCh, _ := cli.KeepAlive(ctx, leaseA.ID)

// 3. Txn(CAS) 获锁
txnResp, _ := cli.Txn(ctx).
    If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
    Then(clientv3.OpPut(lockKey, "client_A", clientv3.WithLease(leaseA.ID))).
    Commit()

// 4. 释放：直接 revoke lease，绑定的 key 自动消失
cli.Revoke(ctx, leaseA.ID)
```

## 与 Redis SETNX 的对比

| 失败场景 | Redis SETNX | etcd lease+Txn |
|---------|-------------|----------------|
| 持锁人崩溃 | 必须等 TTL（E1）| keepalive 断了 server 立即 revoke |
| 续期失败/STW | watchdog 一起卡死，锁过期被抢，A 苏醒后双持（E2） | lease 在 server 端独立计时；A 苏醒后 GET 立即知道自己已不持锁 |
| 主从切换 | 异步复制，可能丢数据（E3） | 走 raft，写入需过半节点 ack，不丢数据 |
| 抢锁公平性 | 轮询争抢 | Watch + Delete 事件，瞬时通知 |

## 不解决什么
- **etcd 的 QPS 比 Redis 低**：etcd 单集群写入 QPS ~10k，Redis 单实例 ~50-100k
- **etcd 的运维复杂度高**：raft 集群、定期 compact、配置 quota、备份/恢复
- **lease 仍有 TTL 概念**：keepalive 心跳间隔 + 过期判定有最小窗口（默认 ~1s）

这些是选型时要权衡的代价，决策树章会展开。

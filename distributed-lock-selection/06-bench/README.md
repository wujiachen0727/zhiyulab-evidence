# 06-bench: Redis vs etcd 获锁 QPS 对比

## 测什么
单 client 串行的"获锁 + 释放"round-trip QPS。**不是绝对吞吐**，是同环境下的相对比例。
分布式锁的核心瓶颈就是网络 round-trip 延迟，串行 QPS 反映服务端处理 + 网络往返成本。

## 前置环境
- Redis 7-alpine 在 6379
- etcd 3.5 在 2379

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code
go run ./06-bench
```

## 输出格式
```
Redis  SETNX+DEL  5000 次  耗时 ...   QPS=...   P50≈.../op
etcd   Txn+Delete 5000 次  耗时 ...   QPS=...   P50≈.../op
速率比：Redis / etcd = ...x
```

## 本机参考值（macOS arm64 + colima docker）
- Redis SETNX + DEL：1978 QPS，P50 ≈ 506µs/op
- etcd Txn(CAS) + Delete：896 QPS，P50 ≈ 1.12ms/op
- 比例：2.21x

## 不适用什么
- 不能直接拿来做容量规划——生产环境的并发吞吐会远高于这个串行数字
- 不同硬件/网络/数据规模差异巨大
- etcd 的写入需 raft 共识，多节点集群下延迟会上升

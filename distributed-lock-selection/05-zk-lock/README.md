# E5：ZooKeeper 临时顺序节点 + Watch 前驱

## 演示什么
ZK 分布式锁的"教科书做法"：
1. 候选者在锁路径下创建 EPHEMERAL_SEQUENTIAL（临时顺序）节点
2. 序号最小的是持锁者
3. 其他人 Watch 自己**前一个**节点的删除事件——**避免惊群**
4. session 断 → 节点自动删 → 锁自动释放

## 前置环境
```bash
docker run -d --name dl-zk -p 2181:2181 zookeeper:3.9
```

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code
go run ./05-zk-lock
```

## 关键证据
- 三个 client 启动时错开 200ms，依次创建 lock-0000000001/2/3 节点
- C1 拿到锁，C2 watch C1，C3 watch C2
- C1 业务完成 delete 节点 → C2 被唤醒拿锁
- C2 完成 → C3 被唤醒拿锁
- 任何 client 中途崩溃，session 超时后 ephemeral 节点被 ZK 删掉，锁自动释放

## 与 etcd 的对比
| 维度 | ZK | etcd |
|------|-----|------|
| 共识协议 | ZAB | raft |
| 临时节点 | EPHEMERAL（session 断即删）| Lease（带 TTL，主动 keepalive）|
| 抢占公平 | EPHEMERAL_SEQUENTIAL + Watch 前驱（无惊群）| Watch + Txn(CAS)（首先收到 Delete 的拿到）|
| Fencing token | 节点序号天然单调递增 | revision 号天然单调递增 |
| 客户端生态 | Java（Curator）成熟 | Go 原生支持，多语言 grpc |
| 运维负担 | 中（JVM 调优、observer 节点）| 中（compact、quota、备份）|

## 不解决什么
- **QPS 比 Redis 低**：3-node ZK 写入吞吐 ~10k QPS 量级
- **客户端 session 超时调参敏感**：太短易误判客户端"崩溃"，太长崩溃恢复慢
- **运维负担**：Java 进程、JVM GC 调优、跨机房延迟敏感

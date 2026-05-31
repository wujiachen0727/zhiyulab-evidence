# E3：Redis 主从切换丢锁

## 复现什么
异步复制是 Redis 单主架构的根本约束。这个实验展示一个最坏时序：
1. 客户端 A 在 master 上 SETNX 成功
2. **复制还没到达 replica**，master 就挂了
3. replica 被提升为新 master（哨兵或人工 failover）
4. 客户端 B 连新 master，对同一个 key 做 SETNX → 成功
5. **此刻 A 和 B 都认为自己持有锁**——互斥被破坏

## 前置环境
- Docker
- 6379、6380 端口空闲

## 运行
```bash
cd articles/distributed-lock-selection/evidence/code/03-master-failover
docker compose up -d
sleep 3   # 等待 replica 完成首次同步
go run main.go
docker compose down
```

## 关键证据
- 程序在 master 上 SETNX 后立即触发 replica `REPLICAOF NO ONE`，再关掉 master
- 在新 master（原 replica）上 GET key → 锁不存在
- 在新 master 上 SETNX 同一 key → 成功
- **同一锁被两个客户端持有**

## 这不是 Redis 的 bug
这是 CAP 三选二的代价。Redis 默认选了 AP（高可用 + 分区容忍），异步复制是为了写入低延迟。
- **要 CP**：可以配 `min-replicas-to-write`，但写入延迟会涨
- **要 P 也要 C**：换 etcd / ZK，它们用 raft/zab 协议，写入要求过半节点确认，但 QPS 会下降

## 与 Kleppmann 论战的关系
Kleppmann 的"How to do distributed locking"指出这个场景，并主张用 fencing token + 单调存储；
antirez 的 Redlock 是另一种方案：N 个独立 master 投票获锁，避免依赖单一主从复制。
两者各有论据，本实验只是把"问题客观存在"展示出来——选哪条路是工程权衡。

# Evidence 总索引

> 本文论据全自造，可在本机复现。所有代码、数据、输出均存档于此目录。

## 目录结构

```
evidence/
├── code/                  # 可执行实验代码（Go）
│   ├── 01-process-crash/  # E1: 进程崩溃锁不释放
│   ├── 02-renewal-fail/   # E2: STW 续期失败双持
│   ├── 03-master-failover/# E3: Redis 主从切换丢锁
│   ├── 04-etcd-lock/      # E4: etcd lease+Txn(CAS) 锁
│   ├── 05-zk-lock/        # E5: ZK 临时顺序节点锁
│   ├── 06-bench/          # E6 配套：QPS benchmark
│   ├── go.mod
│   └── go.sum
├── output/                # 实验运行日志
│   ├── 01-process-crash.log
│   ├── 02-renewal-fail.log
│   ├── 03-master-failover.log
│   ├── 04-etcd-lock.log
│   ├── 05-zk-lock.log
│   └── 06-bench.log
└── data/
    └── decision-matrix.md # E6 决策矩阵（基于 E1-E5 + 06-bench 数据综合）
```

## 实验环境

- macOS 14.x (arm64) + colima docker
- Go 1.26
- Redis 7-alpine（docker）
- etcd 3.5.18（docker, quay.io/coreos/etcd）
- ZooKeeper 3.9（docker）

## 关键数据点（正文将引用）

| 实验 | 关键证据 | 来源 |
|------|---------|------|
| E1 | client_B 等待 10.095s 才拿到锁（TTL=10s）| 01-process-crash.log |
| E2 | T+5.04s B 抢锁成功，T+7s A 苏醒发现锁已易主 | 02-renewal-fail.log |
| E3 | 新 master GET → 不存在；B 在新 master SETNX → true | 03-master-failover.log |
| E4 | 场景 1: A revoke 后 GET → 0 个 key；场景 2: 3.489s 后 C 经 Watch 拿到锁 | 04-etcd-lock.log |
| E5 | 三 client 序号 0/1/2 依次拿锁，watch 前驱无惊群 | 05-zk-lock.log |
| 06-bench | Redis SETNX QPS=1978，etcd Txn QPS=896，比例 2.21x | 06-bench.log |

## 自造 vs 引用

- **自造**：E1/E2/E3/E4/E5（全部 5 个失败/解法实验代码）+ E6 决策矩阵 + 06-bench
- **引用**（≤ 3 处，仅在最后一章）：
  - R1 Kleppmann "How to do distributed locking"（fencing token 论点）
  - R2 antirez "Is Redlock safe?"（反驳）
  - R3 Redis 官方 Redlock 文档（基准定义）

**自造度估算**：6 项独立论据中 6 项自造，外部引用仅做平衡视角佐证不支撑核心论点 → 自造度 ≈ 100%（按论据计数），按内容权重估算 ≥ 85%。

## 复现指南

### 单机一键复现 E1+E2+E4+E6
```bash
docker run -d --name dl-redis -p 6379:6379 redis:7-alpine
docker run -d --name dl-etcd -p 2379:2379 -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.18 /usr/local/bin/etcd \
  --name=dl-etcd \
  --listen-client-urls=http://0.0.0.0:2379 \
  --advertise-client-urls=http://0.0.0.0:2379 \
  --listen-peer-urls=http://0.0.0.0:2380 \
  --initial-advertise-peer-urls=http://0.0.0.0:2380 \
  --initial-cluster=dl-etcd=http://0.0.0.0:2380

cd articles/distributed-lock-selection/evidence/code
go run ./01-process-crash
go run ./02-renewal-fail
go run ./04-etcd-lock
go run ./06-bench
```

### E3 主从需要单独环境
```bash
docker network create dl-net
docker run -d --name dl-redis-master --network dl-net -p 6379:6379 redis:7-alpine \
  redis-server --appendonly no
docker run -d --name dl-redis-replica --network dl-net -p 6380:6379 redis:7-alpine \
  redis-server --appendonly no --replicaof dl-redis-master 6379
sleep 5  # 等待主从同步
go run ./03-master-failover
```

### E5 ZK
```bash
docker run -d --name dl-zk -p 2181:2181 zookeeper:3.9
sleep 5
go run ./05-zk-lock
```

## 已知未验证的边界

- 不同 Redis 持久化配置（RDB/AOF）对锁丢失概率的影响 — 未实验
- Redlock 5 节点的实际 QPS — 未本机复现，决策矩阵中是推演
- 跨机房场景下的 etcd/ZK 性能下降幅度 — 未实验

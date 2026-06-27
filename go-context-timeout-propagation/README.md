# 论据自造产物索引

## 概览

| # | Demo | 类型 | 证明要点 | 状态 |
|---|------|------|---------|:----:|
| 1 | demo1-timing-start | 实验验证 | WithTimeout 计时从调用瞬间开始，排队时间算在内 | ✅ 通过 |
| 2 | demo2-child-longer-timeout | 实验验证 | 子 context 设更长 timeout 无效——最短者胜 | ✅ 通过 |
| 3 | demo3-grpc-vs-http | 实验验证 | gRPC 自动传播 deadline，HTTP 不会 | ✅ 通过 |
| 4 | demo4-db-pool-conflict | 实验验证 | DB 连接池超时和 context 超时会打架 | ✅ 通过 |
| 5 | demo5-deadline-absolute | 实验验证 | deadline 是绝对时间点，不是倒计时 | ✅ 通过 |

## 运行环境

- Go 版本：使用系统默认 Go（go run 直接运行）
- 依赖：demo4 需要 `github.com/mattn/go-sqlite3`，其余为纯标准库
- 所有 demo 均为自包含可复现代码

## 关键输出摘要

### Demo 1: 计时起点
```
设定超时: 3s → 排队 2s → 实际剩余: 1s → 2s 操作被取消
```

### Demo 2: 子不能比父长
```
父 3s / 子 10s → 子实际 deadline == 父 deadline → 子存活 3s 后取消
```

### Demo 3: gRPC vs HTTP
```
gRPC: 框架自动传播 deadline (grpc-timeout header)
HTTP: 无原生机制，下游 ctx.Deadline() 返回 ok=false
```

### Demo 4: DB 连接池冲突
```
5s timeout → 等连接池 ~4s → 实际查询可用时间 ~1s
```

### Demo 5: deadline 绝对时间
```
A 设 5s → 网络 1s → B 剩 4s → B 处理 2s → C 剩 2s
deadline 不会重置，每跳衰减
```

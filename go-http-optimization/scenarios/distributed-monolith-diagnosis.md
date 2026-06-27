# 场景模拟：分布式单体诊断

> 这是模拟场景，基于多个真实案例的共性特征构造，非特定公司经历。

## 场景背景

**公司**：某中型互联网公司，日活 50 万
**团队**：后端团队 8 人，原来维护一个 Go 单体服务
**时间线**：18 个月前"拥抱微服务"拆分了单体

## 当前状态

5 个微服务：
1. **user-service** — 用户认证和资料
2. **order-service** — 订单创建和管理
3. **inventory-service** — 库存管理
4. **payment-service** — 支付处理
5. **notification-service** — 消息通知

## 诊断结果

### 症状1：一个服务挂了全告警

半夜 2 点 payment-service 因为一个 Redis 连接池泄漏 OOM 了。
结果：order-service 报超时，inventory-service 的"扣减后通知"失败，notification-service 的支付成功通知发不出去。
4 个团队的 PagerDuty 全响了。

本质问题：**服务拆了，但依赖链没解耦**。payment-service 是所有业务流的必经节点。

### 症状2：改一个 API 需要同步改 3 个服务

产品需求：订单列表加一个"预计送达时间"字段。
改动：
1. order-service：加字段 + 查询 inventory-service 的仓储位置
2. inventory-service：新增"按 SKU 查仓储位置"接口
3. notification-service：订单状态变更通知里加预计送达时间

3 个服务的 API 都要改。部署要协调 3 个团队的发布窗口。
上线花了 2 周，其中 1.5 周在等发布窗口和联调。

### 症状3：首页加载 300+ 网络调用

用户打开 App 首页，后端聚合了：
- user-service: 2 次调用（用户信息 + 权限）
- order-service: 3 次调用（最近订单 + 待付款 + 推荐商品）
- inventory-service: 1 次调用
- notification-service: 1 次调用
- payment-service: 1 次调用

每个调用平均 8ms（HTTP + JSON 序列化/反序列化 + 网络）。
总延迟：8 × 8ms = 64ms（最理想并行情况）
实际延迟：由于部分调用有串行依赖，实测 P99 = 320ms

### 诊断结论

**这是典型的"分布式单体"**——服务在物理上拆了，在逻辑上没解耦。

### 自救方向

1. **先合再拆**：把 order-service + inventory-service + payment-service 合回一个"交易服务"（它们本来就是同一个业务域）
2. **用模块化单体过渡**：合并后的服务内部用 Go interface + internal 包做模块边界
3. **独立部署验证**：确保模块间的依赖是单向的、通过接口而非数据库共享

### 量化对比

| 指标 | 当前（分布式单体） | 合并后（模块化单体） |
|------|:----------------:|:------------------:|
| 网络调用数 | 8+ | 0（进程内） |
| P99 延迟 | 320ms | 预估 20ms（推演） |
| API 变更影响服务数 | 3 | 1 |
| 发布协调团队数 | 3 | 1 |

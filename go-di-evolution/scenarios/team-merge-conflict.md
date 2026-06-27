# 场景模拟：3 人团队并行修改依赖图

> 这是模拟场景，非真实项目经历

## 场景背景

- 团队：3 名后端开发者
- 项目：一个中等规模的 Go 微服务，当前有 10 个核心依赖
- 周一早会分配任务：
  - 小张（开发者 A）：给支付模块添加风控服务 ServiceD
  - 小李（开发者 B）：给通知模块添加消息队列服务 ServiceE
  - 小王（开发者 C）：给配置中心添加 Redis 连接

## 手动 DI 下的协作冲突

三个人各自拉分支开发。他们的任务都需要改 `main.go`：

- 小张需要在 `main()` 里添加 `svcD := NewServiceD(repo, logger)`
- 小李需要在 `main()` 里添加 `svcE := NewServiceE(repo, cache)`
- 小王需要修改 `cfg := &Config{DSN: "..."}` 为 `cfg := &Config{DSN: "...", RedisURL: "..."}`

三人的代码改动集中在 `main()` 的同一个区域。谁先合进去，后面的人就要解冲突。

## Wire 下的协作

- 小张改 `provider_service.go`（添加 ServiceD）+ `wire.go`（添加 NewServiceD）
- 小李改 `provider_handler.go`（添加 ServiceE）+ `wire.go`（添加 NewServiceE）
- 小王改 `provider_infra.go`（修改 Config 结构体）

三人改了 3 个不同的 provider 文件，只有 `wire.go` 存在轻微重叠（添加 provider 列表行），且冲突范围小、容易解决。

## 模拟结论

手动 DI 的中心化组装代码导致协作摩擦——不是"会不会冲突"的问题，是"第几次合并时冲突"的问题。

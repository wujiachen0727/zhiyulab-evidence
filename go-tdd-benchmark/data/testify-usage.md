# E6: testify 使用率统计

[推演] 2026-04-14

## 数据来源

- GitHub stars: 23k+（来自多个来源确认）
- pkg.go.dev "Imported by" 数据不完整（仅索引到极少数项目）
- 精确的社区使用率无法从公开数据直接获取

## 已知事实

1. testify 是 Go 生态中 stars 最多的测试库（23k+），远超其他竞争者
2. Go 标准库 0 处 testify 使用（E2 已验证）
3. testify 提供了 assert/require/mock/suite 四个包，覆盖了 Go testing 包没有提供的断言、mock、测试套件能力

## 推演

**testify 的高 stars 数说明 Go 社区对 assert/mock 有强烈需求**——Go 标准库不提供这些，社区自己填上了。这恰恰验证了"没有 assert 是设计选择而非遗漏"——如果 Go 团队认为 assert 应该内置，他们有充分的能力和资源去做。

**但标准库坚持不用 testify**——这形成了一个有趣的分叉：官方走"自建辅助函数"路线（t.Helper() 2559 次），社区走"第三方 assert 库"路线（testify 23k+ stars）。两条路线并存，说明"Go 的测试哲学"不是所有人都认同的——但它确实是 Go 官方选择的路线。

## 降级说明

原计划为"数据实测"（GitHub API 统计 testify 使用率），降级为"逻辑推演+参考数据"。原因：精确的 Go 项目 testify 依赖率需要大量 API 调用，且结果受采样偏差影响。

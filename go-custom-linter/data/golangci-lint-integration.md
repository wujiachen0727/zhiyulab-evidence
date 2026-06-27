# E3 golangci-lint 集成踩坑记录

> [推演 + 文档溯源] 基于 golangci-lint v2.11.4 官方文档和 GitHub Issues

## 集成路径

golangci-lint v2 提供两种集成自定义 linter 的方式：

### 方式一：module 插件（推荐）

```yaml
# .golangci.yml
linters:
  custom:
    archlayer:
      type: module
      path: github.com/yourteam/archlayer-linter
```

**坑点**：
1. 自定义 linter 的 go.mod 必须与 golangci-lint 使用完全相同版本的 `golang.org/x/tools`，否则编译时报 `module version mismatch`
2. linter 必须导出 `func New(conf any) ([]*analysis.Analyzer, error)` 签名，不是 `func init()` 注册
3. CI 环境中需要 `go build -buildmode=plugin` 支持——Alpine 等精简镜像可能缺少 CGO 依赖

### 方式二：Go plugin 模式（旧版，逐步废弃）

**坑点**：
1. plugin 模式要求 Go 版本、GOOS/GOARCH、CGO 设置完全一致——CI 和本地不一致就挂
2. macOS 上 plugin 模式有已知兼容性问题
3. v2 已建议迁移到 module 模式

## 关键发现

1. **版本对齐是最大的坑**：golangci-lint 锁定了特定版本的 golang.org/x/tools，你的 linter 也必须用同一版本。每次 golangci-lint 升级，你的 linter 可能需要同步升级依赖。
2. **本地 vs CI 差异**：本地跑通不代表 CI 能跑通。Docker 镜像的 Go 版本、CGO 支持、plugin build mode 都可能不一致。
3. **调试困难**：集成失败时的错误信息通常是 Go module 层面的，需要对 Go module 机制有深入理解才能定位问题。

## 替代方案

对于小团队，更务实的集成方式是不走 golangci-lint 插件，而是：
1. 把 linter 编译为独立二进制
2. 在 CI 的 Makefile 中独立调用
3. 用 `go vet -vettool=./your-linter ./...` 集成到 go vet 链路

这种方式牺牲了 golangci-lint 的统一输出格式，但避免了版本对齐地狱。

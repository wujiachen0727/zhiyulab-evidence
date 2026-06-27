# 论据总索引

## E1: CGO=0 vs CGO=1 对比（实验验证 ✅ 已执行）

- **代码**：`evidence/code/cgo-comparison/run-test.sh`
- **输出**：`evidence/output/e1-cgo-comparison.txt`
- **关键数据**：
  - CGO=0：一条命令编译 5 平台，binary 3.43-3.72MB
  - CGO=1：交叉编译直接失败（缺 C 工具链）
  - 平台覆盖：CGO=0 支持全部 47 个组合，CGO=1 仅支持当前平台
- **承载章节**：开场 + 第一章

## E2: 静态/动态链接部署兼容性（逻辑推演 + 工程经验）

- **代码**：无需独立实验（E1 已展示 otool 输出）
- **输出**：基于 E1 的 otool 输出 + Alpine musl 兼容性的工程共识
- **关键数据**：
  - CGO=0 在 macOS 仍有系统库依赖（libSystem、Security framework）
  - Linux 下 CGO=0 + `-extldflags '-static'` 可产出完全无依赖 binary
  - Alpine（musl）vs Ubuntu（glibc）的兼容性差异是 Docker 部署的经典坑
- **推演标注**：Alpine 兼容性数据基于 Go 社区公认经验，非本次实测
- **承载章节**：第二章

## E3: embed vs 外置文件（实验验证 ✅ 已执行）

- **代码**：`evidence/code/embed-comparison/run-test.sh`
- **输出**：`evidence/output/e3-embed-comparison.txt`
- **关键数据**：
  - 100KB 资源：binary 增长 6%（可忽略）
  - 1MB 资源：binary 增长 64%（需权衡）
  - 10MB 资源：binary 增长 637%（通常不可接受）
  - 编译时间：10MB 嵌入仅增加 0.055s（影响可忽略）
- **承载章节**：第三章

## E4: 交叉编译 vs 容器内编译（逻辑推演 + CI 配置对比）

- **代码**：无独立实验
- **输出**：基于 GitHub Actions 公开文档 + Go 社区实践
- **关键数据**：
  - 交叉编译（CGO=0）：单 job 编译所有平台，~30s（增量缓存后）
  - 容器内编译（CGO=1）：每平台 1 个 job × 对应 Docker image，~2-5min/平台
  - GoReleaser 配置：~30 行 YAML vs 手写 matrix ~80 行 YAML
- **推演标注**：CI 时间数据为典型值估算，非精确实测
- **承载章节**：第四章

## E5: 单 binary vs 多 binary（场景模拟）

- **代码**：无独立实验
- **输出**：场景分析
- **关键数据**：
  - 单 binary 多子命令（cobra）：用户下载 1 个文件，`mycli serve` / `mycli migrate`
  - 多 binary：用户按需下载，每个更小，但分发/版本管理更复杂
  - GoReleaser 支持两种模式，配置差异约 5 行
- **承载章节**：第五章

## 自造论据统计

| # | 论据 | 类型 | 类别 | 来源 |
|---|------|------|:----:|:----:|
| 1 | CGO=0 vs 1 全维度对比 | 实验验证 | 独立论据 | 自造 |
| 2 | 静态/动态链接部署差异 | 逻辑推演 | 独立论据 | 自造 |
| 3 | embed 体积膨胀梯度 | 实验验证 | 独立论据 | 自造 |
| 4 | CI 方案对比分析 | 逻辑推演 | 独立论据 | 自造 |
| 5 | 单/多 binary 场景分析 | 场景模拟 | 独立论据 | 自造 |
| 6 | 决策树框架 | 逻辑推演 | 独立论据 | 自造 |
| 7 | Go 官方支持列表 | 引用 | 独立论据 | 引用 |
| 8 | GoReleaser 配置范式 | 引用 | 独立论据 | 引用 |

**自造度**：6/8 = 75% ✅（目标 ≥ 70%）

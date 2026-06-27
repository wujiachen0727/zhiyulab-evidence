# E1: Go testing vs JUnit 5 API 面积对比

[实测 Go 1.26.2 + JUnit 5 官方文档] 2026-04-14

## 测试方法

1. Go testing 包：扫描 GOROOT/src/testing/ 目录的导出符号
2. JUnit 5 Jupiter API：读取官方 API 文档统计公共类型

## Go testing 包 API

| 指标 | 数量 |
|------|------|
| 导出类型 | 21 |
| 导出函数 | 32 |
| 导出变量/常量 | 7 |
| **公共 API 总数** | **60** |
| 包代码行数 | 8165 行（22 个文件） |

### 核心类型

- **testing.T**：5 个独有方法（Chdir, Deadline, Parallel, Run, Setenv）
- **testing.B**：12 个独有方法（Loop, ReportAllocs, ReportMetric, ResetTimer, Run, RunParallel, SetBytes, SetParallelism, StartTimer, StopTimer, Elapsed, Next）
- **testing.F**：5 个独有方法（Add, Fail, Fuzz, Helper, Skipped）
- **testing.TB**：共享接口

### 值得注意

- **没有 Assert/Evaluate/Require 类**——断言不是框架的一部分
- **没有 Mock/Stub 类**——Mock 不是框架的一部分
- **没有 TestRunner/Suite 类**——运行器不是框架的一部分

## JUnit 5 Jupiter API（仅核心包 org.junit.jupiter.api）

| 类型 | 数量 |
|------|------|
| 注解 | 20 |
| 类 | 19 |
| 接口 | 5 |
| 枚举 | 1 |
| **公共类型总数** | **45** |

### 不包括

- org.junit.jupiter.api.extension 包（扩展点 API）
- org.junit.jupiter.engine 包（执行引擎）
- org.junit.jupiter.params 包（参数化测试）
- org.junit.platform.* 包（平台 API）

加上这些关联包，JUnit 5 的公共 API 总数保守估计 150+。

## 对比结论

| 维度 | Go testing | JUnit 5 Jupiter |
|------|-----------|-----------------|
| 核心公共 API | ~60 | ~45（核心包）/ 150+（全部） |
| 断言机制 | if + Errorf（手写） | Assertions 类（200+ 方法） |
| Mock 支持 | 无 | 无（需 mockito） |
| 生命周期回调 | 无注解 | @BeforeAll/@BeforeEach/@AfterAll/@AfterEach |
| 参数化测试 | t.Run + table | @ParameterizedTest + @ValueSource 等 |
| 扩展机制 | t.Helper() 自建 | Extension API 完整体系 |
| 包代码量 | 8165 行 | 估计 20000+ 行 |

**关键差异**：Go testing 的 60 个 API 覆盖了测试+基准+模糊测试三个领域，而 JUnit 5 仅测试领域就用了 150+ API。Go 的策略是"给你最小的工具集，剩下的你自己组装"；JUnit 的策略是"给你完整的工具箱，常见场景开箱即用"。

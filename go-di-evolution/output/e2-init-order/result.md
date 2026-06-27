# E2 实验结果：初始化顺序错误暴露时机

[实测 Go 1.26.2, Wire v0.7.0]

## 实验方法

构造三种初始化顺序错误场景，分别在手动 DI、Wire、Fx 下测试错误暴露时机。

## 结果

| 方案 | 错误类型 | 暴露时机 | 检测能力 |
|------|---------|---------|---------|
| 手动 DI | nil 依赖（空指针） | 运行时 panic | ❌ 最晚 |
| 手动 DI | 顺序倒置（静默错误） | 可能永远不暴露 | ❌❌ 最差 |
| Wire | 缺少 provider | 编译/代码生成时 | ✅ 最早 |
| Fx/Dig | 缺少 provider | 启动时（依赖图校验） | ⚠️ 较早 |

## 关键发现

1. **手动 DI 最危险的不是 panic，是静默错误**：`NewRepository(nil)` 编译通过，运行不 crash，但逻辑有 bug
2. **Wire 在 `wire generate` 阶段就能发现缺少 provider**，连编译都到不了
3. **Fx 在 `app.Start()` 时校验依赖图**，比运行时早，但比 Wire 晚

## Wire 错误信息示例

```
wire: inject InitializeApp: no provider found for e2wire.ServiceB
needed by *e2wire.ServiceA in provider "NewServiceA"
```

## 可供正文引用的数据点

- "手动 DI 的初始化顺序错误，编译器不会告诉你——它不知道 `nil` 不是合法的依赖"
- "Wire 连代码生成都不让你过——`no provider found for ServiceB`，缺什么一目了然"
- "最可怕的不是运行时 panic，是 `&{<nil>}`——静默错误，可能上线了才发现"

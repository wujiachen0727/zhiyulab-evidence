# 代码实验对比报告

> 实验时间：2026-05-25
> 实验 feature：用户注册（邮箱校验 + 密码强度校验 + 事件通知）
> 语言：Go

## 实验设计

同一个 feature（用户注册），分别用两种架构实现：
- **简单三层**：handler → service → repo
- **DDD 战术模式**：controller → app service → domain service → entity/value objects → repository interface → repository impl → events

## 量化对比

| 维度 | 简单三层 | DDD 战术模式 | 倍率 |
|------|:-------:|:-----------:|:----:|
| 文件数 | 3 | 7（核心），9（含拆分） | 2.3x - 3x |
| 代码行数（不含注释/空行） | ~95 | ~210 | 2.2x |
| 依赖层数 | 2 | 4 | 2x |
| 新人理解跳转次数 | 2 | 6 | 3x |
| 新增一个字段的改动文件数 | 2 | 4-5 | 2-2.5x |
| 接口/抽象数量 | 0 | 3（Repository接口、Email值对象、Publisher接口） | — |

## 跳转深度分析（E2）

### 简单三层：新人理解路径

```
handler.RegisterHandler (入口)
  └→ service.Register (业务逻辑+校验+持久化+事件，都在这)
       └→ repo.CreateUser (SQL)
```

**跳转 2 次**，每次跳转都有明确的职责变化。新人看完 3 个文件就理解全貌。

### DDD 版：新人理解路径

```
controller.Register (入口)
  └→ dto.RegisterCommand (要先理解"为什么参数不直接传")
       └→ application.UserApplicationService.Register (编排)
            └→ service.UserDomainService.RegisterUser (业务规则)
                 ├→ valueobject.NewEmail (为什么校验在值对象里？)
                 ├→ valueobject.NewPassword (同上)
                 ├→ entity.NewUser (聚合根是什么？)
                 └→ repository.UserRepository.Save (为什么是接口？)
                      └→ infrastructure.MySQLUserRepository.Save (实现在另一个包)
```

**跳转 6 次**，其中有 3 次需要理解"为什么要这样拆"（值对象存在的意义、接口/实现分离的原因、应用服务 vs 领域服务的区别）。

### 关键发现

跳转次数从 2 增加到 6，但更重要的是**认知门槛**的变化：
- 简单三层的每次跳转都是"下一层做什么"——直觉可理解
- DDD 的跳转中有 3 次需要理解 DDD 概念本身（值对象、聚合根、端口-适配器）——需要前置知识

**新人 onboarding 估算**：
- 简单三层：看懂全流程 ~15 分钟
- DDD 版：看懂全流程 ~45-60 分钟（含理解"为什么要这样"）
- 如果新人没有 DDD 背景：+2-4 小时学习 DDD 概念

## 成本分解

### 初始开发成本

| 成本项 | 简单三层 | DDD | 差异来源 |
|--------|:-------:|:---:|---------|
| 写代码 | 基准 | +115% | 更多文件、接口、值对象 |
| 设计思考 | 低 | 高 | 需要决定聚合根边界、值对象粒度 |
| 测试 | 简单 mock | 多层 mock | Repository 接口需要 mock |

### 每个新接口的"仪式成本"

DDD 版每新增一个 CRUD 接口，**固定开销**：
1. 新增/修改 DTO（Command + Result）
2. Application Service 新增方法（编排）
3. Domain Service 新增方法（规则）
4. 可能的新值对象
5. Repository 接口新增方法
6. Repository 实现新增方法
7. Controller 新增路由

简单三层只需：handler + service 方法 + repo 方法 = **3 步 vs 7 步**。

## 结论

DDD 战术模式的工程成本确实在 **2-3 倍**范围内（验证了掘金文章的估算），但更关键的隐性成本是：
1. **认知跳转 3x**——不是"多写几个文件"的问题，是"每次改动都要理解整条链路"
2. **仪式代码不可跳过**——一旦团队确立 DDD 规范，即使是最简单的 CRUD 也要走完全流程
3. **新人门槛非线性增长**——理解代码需要 3x 时间，但理解"为什么这样写"需要前置知识

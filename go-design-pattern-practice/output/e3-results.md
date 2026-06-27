# E3 实验记录

## 实验名
e3-middleware-evolution — HTTP 中间件链的自然演化

## 核心洞察
1. 从"简单 Handler"到"中间件链"的演化是一个自然过程——不是预先设计好的
2. Go 的 http.Handler 接口（1 个方法）使得装饰器模式的实现成本极低
3. 不同的中间件组合 = 不同的策略选择
4. Go 中不需要 Filter/FilterChain/配置 XML——一个 Middleware 函数类型就够了

## 对比：Java 式 vs Go 式

| 维度 | Java | Go |
|------|------|----|
| 接口定义 | Filter (doFilter) | Handler (ServeHTTP) |
| 包装方式 | FilterChain 链式调用 | 函数返回 Handler |
| 组合配置 | web.xml / 注解 | 切片 + 函数组合 |
| 新增一个中间件 | 实现 Filter + 配置 | 实现 Middleware 函数 |

## 运行环境
- Go 版本：通过 go version 确认
- 运行时间：2026-06-08

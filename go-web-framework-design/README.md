# Evidence 论据索引

> 文章：《Gin 很好，但你的项目可能需要更多》
> 核心论点：Gin 覆盖了 80% 的 Web 开发需求，但 DI、错误处理分层、测试隔离等 20% 才是区分"能跑"和"能维护"的关键。

## 代码论据

### E1: 无分层 Gin 项目
- **位置**：`code/e1-flat-gin/main.go`
- **内容**：一个文件包含所有 handler 的用户 CRUD，展示 handler 直接操作 sql.DB 的典型模式
- **论证**：Gin 的 80%——快速启动一个能跑的项目多容易
- **问题标注**：通过注释标出 5 个典型维护性问题

### E2: 分层版项目
- **位置**：`code/e2-layered/`
- **内容**：同样的用户 CRUD，使用接口分层（handler → service → repository）
- **论证**：那 20%——接口抽象、依赖注入、测试隔离
- **文件结构**：
  - `main.go` — 入口 + 依赖注入
  - `model/user.go` — 领域模型
  - `repository/repository.go` — 数据访问接口定义
  - `repository/mysql_user.go` — MySQL 实现
  - `service/user_service.go` — 业务逻辑
  - `service/user_service_test.go` — 单测示例（展示 mock 多简单）
  - `handler/user_handler.go` — HTTP 处理 + 统一错误映射

### E4: 错误处理对比
- **位置**：`code/e4-error-handling/`
- **内容**：
  - `flat_errors.go` — 错误散落在各 handler 的模式（5 个典型问题）
  - `layered_errors.go` — 自定义 AppError + 中间件统一处理
- **论证**：错误处理是"能跑"和"能维护"的分水岭之一

## 数据论据

### 结构对比统计
- **位置**：`data/comparison-stats.md`
- **内容**：E1 vs E2 的量化对比——文件数、代码行数、换数据源改动量、mock 难度评估
- **关键数据**：E2 多 71% 代码量，但换数据源从改 N 个文件降到改 1 个文件

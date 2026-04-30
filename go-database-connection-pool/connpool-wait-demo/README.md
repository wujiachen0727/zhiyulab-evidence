# connpool-wait-demo

这个实验用一个自定义的 `database/sql/driver` 模拟固定耗时的数据库查询。

目的不是测试某个真实数据库的性能，而是隔离观察 `database/sql` 连接池本身：当 `MaxOpenConns` 小于并发请求数时，请求会在 Go 侧排队，`DB.Stats().WaitCount` 和 `WaitDuration` 会增长。

## 运行环境

- Go 1.26.2 darwin/arm64（本次实测环境）
- 无外部数据库依赖

## 运行方式

```bash
go run .
```

程序会输出：

- 每个场景的 Markdown 表格
- 每个场景的 CSV 行，方便复制到数据文件

## 实验边界

- 查询耗时由 `time.Sleep` 模拟，属于连接池行为测试，不代表 MySQL/PostgreSQL 的真实性能。
- 实验用于证明“等待连接会进入应用延迟”，不是给生产环境推荐参数值。

# E1 实验结果：三层分层对比

**测试环境**：Go 1.26.2, darwin/arm64 (Apple M4 Pro)
**测试方法**：模拟 HTTP 微服务的 DB 查询错误，对比"无分层"vs"三层分层"的 HTTP 响应

## 版本1：无分层

```
HTTP Status: 500
Response Body: query failed: pq: connection refused (SQLSTATE 08006), database=users_db, query=SELECT id, email FROM users WHERE id=$1
```

**泄露信息检测**：
- ⚠️ 泄露了: 'pq:' → 暴露数据库类型（PostgreSQL）
- ⚠️ 泄露了: 'SQLSTATE' → 暴露 SQL 标准错误码
- ⚠️ 泄露了: 'database=' → 暴露数据库名
- ⚠️ 泄露了: 'SELECT' → 暴露查询语句
- ⚠️ 泄露了: 'users' → 暴露表名

## 版本2：三层分层

```
HTTP Status: 503
Response Body: {"error":"service temporarily unavailable","code":"SERVICE_UNAVAILABLE","trace_id":"abc-123"}
```

**泄露信息检测**：
- ✅ 未泄露: 'pq:'
- ✅ 未泄露: 'SQLSTATE'
- ✅ 未泄露: 'database='
- ✅ 未泄露: 'SELECT'
- ✅ 未泄露: 'users'

## 结论

无分层版本暴露了 5 项敏感信息（数据库类型、SQL状态码、数据库名、查询语句、表名），足以让攻击者定位数据库类型并推测表结构。三层分层版本只返回通用错误码 + trace_id，零泄露。

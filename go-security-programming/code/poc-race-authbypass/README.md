# E4: goroutine 隔离不保证 TOCTOU 原子性

## 证伪结果

**假设**：race condition 可导致认证绕过。  
**结果**：`go run -race main.go` 稳定报出 2 处 DATA RACE——
指向 `sneakyAttacker` 对 `sess.Role` 的原地写 vs `adminOperation` 的权限检查读。

## 运行

```bash
go run -race main.go
```

## 观察要点

- sync.RWMutex 保护了 map 访问，没保护 map value 指向的 struct 字段
- 即使拿到锁读 `store[token]` 返回的是 *Session 指针——
  读完之后另一个 goroutine 仍可通过同一指针原地修改字段
- 这是真实业务代码里非常常见的 pattern

## 环境

Go 1.26.2 darwin/arm64

## 对应章节

第三章 第 3 个幻觉 · goroutine 隔离等于状态安全

# Evidence Index — go-map-fatal-not-panic

## 环境预检

| 项目 | 结果 |
|---|---|
| Go 环境 | ✅ go1.26.2 darwin/arm64 |
| 文章目录 | ✅ `articles/go-map-fatal-not-panic/` |
| evidence 目录 | ✅ 已创建并写入产物 |

## 论据执行结果

| ID | 类型 | 描述 | 状态 | 产出路径 |
|---|---|---|---|---|
| E1 | 实验验证 | 对比普通 `panic` 可 recover 与 `concurrent map writes` fatal 不可 recover | ✅ 完成 | `evidence/code/recover-vs-fatal/`、`evidence/output/recover-vs-fatal/` |
| E2 | 数据实测 | benchmark 普通 map、`map+RWMutex`、`sync.Map` 在读多写少/写多场景下的开销差异 | ✅ 完成 | `evidence/code/map-sync-bench/`、`evidence/output/map-sync-bench/`、`evidence/data/map-sync-bench.md` |
| E3 | 逻辑推演 | 推导 recover 只能恢复控制流，不能恢复已不可信的数据结构 | ✅ 完成 | `evidence/scenarios/untrusted-state.md` |
| E4 | 场景模拟 | 构造服务继续使用不可信 map 状态的工程场景 | ✅ 完成 | `evidence/scenarios/untrusted-state.md` |
| E5 | 外部引用 | Go FAQ 关于 map 不默认原子/并发安全的设计理由 | ✅ 已在 thesis 求证日志确认 | `drafts/grounding-log.md` |

## 自造度统计

- 独立论据：5 项
- 自造论据：4 项（E1/E2/E3/E4）
- 外部引用：1 项（E5）
- 自造占比：4 / 5 = 80%
- 降级项：无

## 正文引用建议

### E1

普通 `panic` 可被同 goroutine 的 `defer recover` 接住，exit code 为 0；并发写 map 触发 `fatal error: concurrent map writes`，exit code 为 2。

### E2

本机 Go 1.26.2 / Apple M4 Pro 实测：读多写少场景 plain map 约 3.86 ns/op，`map+RWMutex` 约 4.85 ns/op，`sync.Map` 约 15.46 ns/op；写多场景 plain map 约 5.76 ns/op，`map+RWMutex` 约 6.57 ns/op，`sync.Map` 约 33.36 ns/op 且 1 alloc/op。

### E3/E4

`recover` 能恢复控制流，但不能证明 map 内部状态仍然可信。对已经不可信的数据结构继续服务请求，比进程退出更危险。

# alloc-vs-gc

分配成本 vs GC 成本分离实验（分配新对象 vs 复用对象池），验证 GC 回收是短生命周期分配的主要开销。

## 运行环境

- Go 1.26.2（darwin/arm64，Apple M4 Pro）
- 无第三方依赖（仅标准库）

## 运行方式

```bash
cd alloc-vs-gc
go run .
```

## 产出

- 运行输出见 `../output/alloc-vs-gc/`
- 数据解读见 `../../evidence/README.md`

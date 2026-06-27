# E2: 跳表 vs B+ 树内存场景 Benchmark

## 运行环境

- Go 1.26.4 darwin/arm64
- macOS Darwin 25.5.0 ARM64

## 运行步骤

```bash
cd evidence/code/skiplist-vs-btree-bench
go run main.go
```

## 实现说明

### 跳表实现
- 参考 Redis t_zset.c 的设计
- maxLevel=32, p=0.25（Redis 默认）
- 含 backward 指针和 span 维护（模拟 Redis 跳表开销）

### B+ 树实现
- 简化版，btreeOrder=64（模拟 cache line 对齐）
- 只处理根分裂（benchmark 足够，非生产级）
- 叶子节点用 slice（连续内存，cache 友好）

## 预期输出

见 `../../output/skiplist-vs-btree-bench-output.txt`

关键数据：
- 插入：B+ 树比跳表快 2x-11x（N 越大差距越大）
- 范围查询：B+ 树比跳表快 8x-364x

## 重要声明

- 本 benchmark 是**简化实现对比**，非生产级数据结构
- 跳表含 Redis 风格的额外开销（span/backward），B+ 树叶子用连续 slice
- 比较不完全公平，但趋势明确：B+ 树的 cache locality 优势在内存场景依然存在
- 结论：Redis 选跳表**不是因为性能**，而是综合权衡（简洁性 + 双结构协作 + 性能足够）

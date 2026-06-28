# E3: 写入密集场景 fork 期间 RSS 变化曲线

## 实验目的

测量写入密集负载下 fork() 期间 Redis RSS 内存变化曲线，验证 fork/COW 的内存风险。

## 运行方式

```bash
cd evidence/code
bash run-all-experiments.sh
```

结果输出到 `evidence/output/e1-e6-results.md` § E3。

## 关键发现

- RSS 峰值达基准 used_memory 的 6.41 倍（49MB → 252MB）
- **诚实修正**：RSS 暴涨主因是父进程持续写入分配新内存，不是 COW 复制（COW 仅占 1.1%）

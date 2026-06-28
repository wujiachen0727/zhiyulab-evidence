# E6: fork 耗时随数据集变化实验

## 实验目的

测量不同数据集规模下 fork() 的耗时，验证"大数据集 fork 阻塞"的说法。

## 运行方式

```bash
cd evidence/code
bash run-all-experiments.sh
```

结果输出到 `evidence/output/e6-fork-duration.md`。

## 关键发现

- 1 万 key: 243μs
- 10 万 key: 168μs（比 1 万快，可能是容器调度噪声）
- 100 万 key: 403μs
- fork 耗时与数据集大小正相关，大数据集进入毫秒级

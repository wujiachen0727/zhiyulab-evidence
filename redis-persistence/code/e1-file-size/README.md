# E1: 文件大小对比实验

## 实验目的

对比 RDB / AOF / 混合三种持久化方式在相同数据集下的文件大小，验证 aof-use-rdb-preamble 启用后 AOF base file 与 RDB 文件的关系。

## 运行方式

```bash
cd evidence/code
bash run-all-experiments.sh
```

结果输出到 `evidence/output/e1-e6-results.md` § E1。

## 关键发现

aof-use-rdb-preamble 启用后（Redis 7.0+ 默认），AOF base.rdb 与 dump.rdb 完全一致（24,777,889 字节）。

# E2 GIL 并发边界实验结果

[实测 Python 3.9.6]

## 核心假设与证伪点

- 假设：CPU-bound 场景下，`threading` 不会随线程数线性提速。
- 证伪点：如果本机结果出现线程线性提速，则不使用性能结论，只保留官方 GIL 事实。
- 结果：证伪未发生。本机 4 线程比顺序执行略慢，4 进程有明显提速。

## 方法

同一个纯 Python CPU-bound 函数执行 4 个任务，每个任务做 300 万次整数运算。对比三种方式：顺序执行、4 个线程、4 个进程。

## 原始输出

```text
[实测 Python 3.9.6] CPU-bound 并发边界最小实验
platform=macOS-26.5-arm64-arm-64bit
executable=/Library/Developer/CommandLineTools/usr/bin/python3
cpu_count=14, tasks=4, n_per_task=3000000
sequential: elapsed=0.531s, speedup_vs_sequential=1.00x, checksum=576000968
threading-4: elapsed=0.550s, speedup_vs_sequential=0.97x, checksum=576000968
multiprocessing-4: elapsed=0.260s, speedup_vs_sequential=2.04x, checksum=576000968
```

## 可供正文引用的结论

在这组本机 CPU-bound 实验里，4 个线程没有带来线性提速，反而略慢于顺序执行；4 个进程能利用更多核心，但要引入进程模型。正文中只写“本机观察：CPU-bound threading 未线性提速”，不写夸张泛化倍数，并与官方 GIL 定义配合使用。

# gil-threading-demo

用途：验证 CPU-bound 场景里，Python 的 `threading` 写法简单不等于多核并行自然发生。

运行：

```bash
python3 gil_threading_demo.py
```

说明：脚本比较单线程顺序执行、4 个线程、4 个进程三种方式。结果只作为本机观察，不写泛化性能倍数；正文使用时只表达“CPU-bound threading 未线性提速”，并与官方 GIL 定义一起使用。

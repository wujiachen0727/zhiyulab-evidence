# Java virtual thread 任务编排器实验

## 目的

用 JDK 26 的 virtual thread 实现同一个任务编排器，观察 Java 线程心智模型下三类责任的位置：

1. 状态所有权：调用方仍用普通对象和集合持有聚合状态；virtual thread 不自动解决共享状态语义。
2. 调度责任：代码保持同步阻塞写法，等待成本主要由 JVM virtual thread 调度承接。
3. 失败边界：超时和子任务失败后的取消仍由应用编排 Future 完成。

## 运行

```bash
/opt/homebrew/opt/openjdk/bin/javac Main.java
/opt/homebrew/opt/openjdk/bin/java Main
```

## 输出说明

输出包含 `java-success`、`java-timeout`、`java-worker-error` 三个场景。本实验不做性能评测，只观察责任表达位置。

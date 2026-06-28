# E2: 恢复时间对比实验

## 实验目的

对比 RDB 和 AOF 在相同数据集下的重启恢复时间，验证"RDB 恢复快于 AOF"的优缺点。

## 运行方式

```bash
cd evidence/code/e2-recovery-time
bash run.sh
```

结果输出到 `evidence/output/e2-recovery-time.md`。

## 降级说明

Redis 7.4 加载速度极快，即使 500 万 key 也在毫秒级完成，`loading_total_time` 无法通过轮询采集。改为推演（基于文件结构分析）。

**标注**：[推演] 基于文件结构分析，非直接实测对比。

# E5: AOF everysec 断电丢失窗口实验

## 实验目的

用 kill -9 模拟断电，验证 AOF everysec 策略的丢失窗口，纠正"AOF 不丢数据"的认知偏差。

## 运行方式

```bash
cd evidence/code/e5-power-loss
bash run.sh
```

结果输出到 `evidence/output/e5-power-loss.md`。

## 关键发现

- kill -9 后丢失 1230 key（2.46%）
- 验证了"AOF everysec 最多丢 1 秒"的承诺
- **边界条件**：kill -9 ≠ 真实断电，OS page cache 仍可能持久化部分数据

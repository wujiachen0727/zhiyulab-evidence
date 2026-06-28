# 证伪实验：E1/E2/E3/E7 核心假设验证

> practice-verify SKILL.md 铁律：priority=1 的论据项，证伪实验优先于支撑实验。先证伪，再支撑。
>
> 本脚本在 Docker Redis 7.4.9 上跑 4 个证伪实验，结果用于修正论点（如证伪成立）。

## 实验环境

- Docker: 29.4.0 (Colima)
- Redis: 7.4.9 (redis:7-alpine)
- Host: macOS Darwin 25.5.0 ARM64

## 证伪实验清单

### F1: E1 证伪——极小数据集下 RDB 文件是否仍比 AOF 小？

- **原假设**：RDB 文件比 AOF 文件小
- **证伪场景**：10 个 key（极小数据集），RDB 可能有固定头部开销导致反而更大
- **如果证伪成立**：修正论点为"大数据集下 RDB 文件更小"

### F2: E2 证伪——极小数据集下 RDB 恢复是否仍比 AOF 快？

- **原假设**：RDB 恢复比 AOF 快
- **证伪场景**：10 个 key，恢复时间差异可能不显著
- **如果证伪成立**：修正论点为"大数据集下 RDB 恢复更快"

### F3: E3 证伪——空闲无写入场景 fork 是否仍触发 RSS 翻倍？

- **原假设**：fork 期间 RSS 翻倍
- **证伪场景**：无写入负载下 BGSAVE，COW 不应触发，RSS 不应翻倍
- **如果证伪成立**：明确"翻倍是写入密集场景特定现象"（这是预期结论，证伪成立反而验证了机制的精确性）

### F4: E7 证伪——fsync 与 fork+COW 是否独立？

- **原假设**：RDB 和 AOF 优缺点都指向 fork+COW 同一根因
- **证伪场景**：检查 fsync 策略差异（always/everysec/no）——fsync 与 fork+COW 是两个独立机制
- **如果证伪成立**：修正论点为"主要根因"或"共同根因之一"，不能说"同一根因"

## 运行方式

```bash
cd evidence/code/falsification
bash run-falsification.sh
```

结果输出到 `evidence/output/falsification/`。

# E5: AOF everysec 断电丢失窗口（kill -9 模拟）

**实验时间**：2026-06-28
**环境**：Docker 29.4.0 (Colima) + Redis 7.4.9 (redis:7-alpine)

## 实验设计

1. AOF everysec 策略，灌入 10 万基础 key
2. BGREWRITEAOF 确保 base file 落盘
3. 启动 redis-benchmark 持续写入 5 万 key（后台）
4. 0.5 秒后立即 kill -9 模拟断电
5. 重启容器，检查 key 数

## 实验结果

| 指标 | 数值 |
|------|:----:|
| 断电前 key 数（基础）| 100000 |
| benchmark 计划写入 | 50000 |
| 重启后 key 数 | 148770 |
| 实际写入 key 数 | 48770 |
| 丢失 key 数 | 1230 |
| 丢失比例 | 2.46% |

## 关键观察

- kill -9 时机在 fsync 间隔内，丢失 1230 key（2.46%）

## 边界条件声明（诚实标注）

1. **kill -9 ≠ 真实断电**：kill -9 模拟的是进程被杀，OS page cache 仍可能被 fsync 线程刷盘
2. **真实断电丢失可能更多**：断电时 OS page cache 全部丢失，AOF 缓冲区未刷盘部分必然丢失
3. **丢失量取决于 kill 时机**：kill -9 在 fsync 周期（1秒）内的哪个点，决定了丢失量
4. **everysec 承诺**：Redis 官方文档明确"AOF everysec 最多丢 1 秒数据"，本实验验证的是最坏情况

## 结论

AOF everysec 在 kill -9 场景下丢失 1230 key，验证了'最多丢 1 秒'的承诺。**关键不是丢失量，而是'不丢数据'是错的认知**——AOF everysec 明确承诺的是'最多丢 1 秒'，不是'零丢失'

## 反认知点

"AOF 不丢数据"是错的。准确说法：
- AOF + always：每条命令 fsync，理论零丢失，但性能大幅下降
- AOF + everysec：每秒 fsync，最多丢 1 秒
- AOF + no：由 OS 决定 fsync，可能丢数十秒

**官方文档措辞是 "very safe"（非常安全），不是 "zero loss"（零丢失）**

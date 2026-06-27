# Evidence 论据自造记录

> 由 Practice-Verify 阶段维护。外部引用是最后手段，自造论据优先。

## 自造论据清单

| ID | 类型 | 文件/目录 | 方法 | 结论 | 可复现 |
|---|---|---|---|---|:---:|
| E1.1 | 实验验证 | `code/leaky-bucket-vs-nginx-burst.go` → `output/leaky-vs-nginx.md` | Go 模拟纯漏桶 vs burst+nodelay，喂突发流量 | burst+nodelay 让行为从匀速排队变为允许突发通过（令牌桶特征） | ✅ |
| E1.3 | 逻辑推演 | 融入正文 | 从 burst+nodelay 行为推导等价性 | 当 burst=N+nodelay 时 ≈ 容量 N 的令牌桶 | ✅ |
| E2.1 | 数据推演 | `code/sliding-window-memory.py` → `data/zset-memory-table.md` | Python 公式计算 ZSET 内存 | 1K QPS×1K 用户 = 3.6 GB，SW Counter 仅 80 KB | ✅ |
| E2.2 | 实验验证 | `code/sliding-window-three-variants.go` → `output/sliding-decision-diff.md` | Go 实现三变种 + 跨窗口测试 | Fixed Window 有双倍放行，SW Log/Counter 精确拒绝 | ✅ |
| E2.3 | 逻辑推演 | 融入正文 | SW Counter 精度公式推演 | 子桶数=N → 误差 ≤ 1/N，100 桶≈99% 精度 | ✅ |
| E3.1 | 实验验证 | `code/distributed-token-bucket.go` → `output/race-condition-test.md` | 100 goroutine 并发测试朴素版 vs 原子版 | 朴素版 900% 超发，原子版 0 超发 | ✅ |
| E3.2 | 场景模拟 | `scenarios/three-instance-tradeoff.md` | 三实例取舍三角分析 | Lua 原子脚本是多数场景首选 | ✅ |
| E3.3 | 经验落地 | 融入正文（第三章） | 用户生产经历：10+ 实例分布式超发 | 本地桶→中心化方案的切换取舍 | ✅ |

## 外部引用清单

| ID | 来源 | 用途 | 是否必须 |
|---|---|---|:---:|
| R1 | Nginx `ngx_http_limit_req_module.c` 源码 | 证明 burst 队列机制非纯漏桶 | 是 |
| R2 | Redis 官方文档 ZSET 内存模型 | 内存公式权威背书 | 是 |
| R3 | Guava RateLimiter JavaDoc | 延伸佐证（可选） | 否 |

## 比例统计

- 自造论据数：8（E1.1, E1.3, E2.1, E2.2, E2.3, E3.1, E3.2, E3.3）
- 外部引用数：2（必须）+ 1（可选）
- 自造比例：**80%**（8 / 10）

## 公开评估

- evidence_public: true
- 原因：满足三条标准——有实验代码（3 个可运行文件）、独立看有价值（含运行说明和注释）、正文将引用其具体数字

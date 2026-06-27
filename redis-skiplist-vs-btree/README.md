# Evidence 总索引

> 论据自造产物索引。所有产物可复现。

## 实验代码

| ID | 路径 | 说明 | 环境 |
|----|------|------|------|
| E1 | `evidence/code/zset-encoding-test/test.sh` | ZSET 编码切换实测（listpack → skiplist） | Redis 8.8.0 |
| E2 | `evidence/code/skiplist-vs-btree-bench/main.go` | 跳表 vs B+ 树内存场景 benchmark | Go 1.26.4 darwin/arm64 |

## 实验输出

| ID | 路径 | 关键数据 |
|----|------|---------|
| E1 | `evidence/output/zset-encoding-test-output.txt` | 128→129 元素内存 1596→12576 字节（7.9x 跃升） |
| E2 | `evidence/output/skiplist-vs-btree-bench-output.txt` | B+ 树范围查询比跳表快 8x-364x |

## 数据分析

| ID | 路径 | 说明 |
|----|------|------|
| E3 | `evidence/data/node-size-media-alignment.md` | 节点大小 vs 介质对齐分析（B+ 树 4KB 磁盘页 vs 跳表 64B cache line） |
| E4 | `evidence/data/t_zset-source-analysis.md` | Redis 源码 t_zset.c 拆解 + antirez 原话考证 |

## 融入正文的论据（无独立文件）

| ID | 类型 | 说明 |
|----|------|------|
| E5 | 逻辑推演 | 推演 B+ 树在内存场景的额外开销（分裂、合并、平衡维护） |
| E6 | 场景模拟 | 面试场景揭穿"并发优势论"——Redis 单线程，并发无从发挥 |
| E7 | 逻辑推演 | 反驳"实现简单论"的简化版本——antirez 原话是"简单是原因之一"但非主要原因 |
| E8 | 数据整理 | MySQL/LevelDB/RocksDB/Redis 数据结构选型交叉验证"介质决定结构"框架 |

## 外部引用

| ID | 内容 | 来源 |
|----|------|------|
| R1 | William Pugh《Skip Lists: A Probabilistic Alternative to Balanced Trees》 | ACM TOCS 1990 / citeseerx |
| R2 | B-skiplist 论文（arxiv 2507.21492） | arxiv.org/abs/2507.21492 |
| R3 | antirez 2010-03-06 邮件列表原话 | 已考证，见 evidence/data/t_zset-source-analysis.md |

## 自造度统计

- 独立论据：8 项（E1-E8）
- 外部引用：3 项（R1/R2/R3）
- 自造度 = 8 / (8 + 3) = **73%** ✅ 达标（≥70%）

## 核心证伪结果

E2 benchmark 推翻了初始假设"内存场景下跳表和 B+ 树性能相近"。论点已修正为"Redis 选跳表不是因为性能，而是综合权衡"。

详见 `drafts/grounding-log.md` § 论证阶段 · 核心证伪结果。

# E3：节点大小 vs 介质对齐分析

## 分析目标

验证"B+ 树为磁盘优化，跳表为内存优化"的介质决定结构框架。

## B+ 树节点对齐分析

### 磁盘场景（MySQL InnoDB）

- **InnoDB 页大小**：16KB（默认，可配置 4K/8K/16K/32K/64K）
- **磁盘扇区大小**：512B - 4KB
- **B+ 树节点设计**：每个节点对齐一个页（16KB）
  - 非叶子节点：存 ~1170 个指针（16KB / (6B child pointer + 12B key) ≈ 1170）
  - 叶子节点：存 ~16 行数据（每行 ~1KB）
- **3 层 B+ 树可索引**：1170 × 1170 × 16 ≈ **2000 万行**
- **cache locality 来源**：一次磁盘 IO 读一个页，页内所有 key 都可用

### 内存场景（假设 B+ 树用于内存）

- **CPU cache line**：64B（主流架构）
- **B+ 树节点如果对齐 cache line**：节点大小应为 64B 的倍数
- **但 B+ 树的传统优势（页对齐）在内存场景失效**：
  - 内存随机访问无 IO 惩罚
  - 紧凑节点的好处是 cache 命中，但需要节点大小匹配 cache line
  - 如果节点太大（如 16KB），反而超出 L1 cache（通常 32-64KB）

## 跳表节点对齐分析

### Redis 跳表节点结构（t_zset.c）

```c
typedef struct zskiplistNode {
    sds ele;                          // 8B 指针
    double score;                     // 8B
    struct zskiplistNode *backward;   // 8B 指针
    struct zskiplistLevel {
        struct zskiplistNode *forward; // 8B 指针
        unsigned int span;             // 4B
    } level[];                        // 动态数组，平均 1.33 层（p=0.25）
} zskiplistNode;
```

- **单节点平均大小**：8+8+8 + 1.33×(8+4) ≈ **36B**（不含 SDS 本身）
- **接近 cache line 大小**（64B）——但单个节点不保证对齐
- **稀疏指针**：跳表的指针是跨节点的，cache locality 不如 B+ 树紧凑节点

### 内存场景下的权衡

| 维度 | B+ 树 | 跳表 |
|------|-------|------|
| 节点大小 | 紧凑（可对齐 cache line） | 稀疏（~36B + SDS） |
| cache locality | 强（节点内连续） | 弱（指针跳跃） |
| 实现复杂度 | 高（分裂、合并、平衡） | 低（概率平衡，无旋转） |
| 范围查询 | 快（叶子链表连续） | 快（底层链表） |
| 插入开销 | 高（可能触发分裂） | 低（随机层级，无平衡） |

## 结论

**E3 验证了"介质决定结构"框架的部分正确性，但需要修正**：

- B+ 树的"磁盘友好"特性（页对齐、减少 IO）在内存场景**确实不再必要**
- 但 B+ 树的"内存友好"特性（cache locality）**依然存在**——E2 benchmark 证实
- 跳表的"内存友好"特性**不是 cache locality**，而是**实现简洁 + 插入低开销 + 与 dict 协作自然**

**修正后的框架**：
- B+ 树为磁盘而生：✅ 正确
- 跳表为内存而生：⚠️ 部分正确——跳表不是"内存最优"，而是"内存场景下够用且简洁"
- Redis 选跳表的根因：**不是性能优势，而是综合权衡（简洁性 + 双结构协作 + 性能足够）**

## 数据来源

- InnoDB 页大小：MySQL 官方文档
- CPU cache line：主流架构常识（x86/ARM 均 64B）
- Redis 跳表结构：Redis 源码 src/t_zset.c
- 跳表平均层数：William Pugh 论文（p=0.25 时平均 1.33 层）

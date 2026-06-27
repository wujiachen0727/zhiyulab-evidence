# E4：Redis 源码 t_zset.c 拆解 + antirez 原话考证

## 源码注释（已查证，来自 GitHub redis/redis unstable 分支）

t_zset.c 顶部注释关键内容：

```c
/* ZSETs are ordered sets using two data structures to hold the same elements
 * in order to get O(log(N)) INSERT and REMOVE operations into a sorted
 * data structure.
 *
 * The elements are added to a hash table mapping Redis objects to scores.
 * At the same time the elements are added to a skip list mapping scores
 * to Redis objects (so objects are sorted by scores in this "view").
 *
 * Note that the SDS string representing the element is the same in both
 * the hash table and skiplist in order to save memory.
 *
 * This skiplist implementation is almost a C translation of the original
 * algorithm described by William Pugh in "Skip Lists: A Probabilistic
 * Alternative to Balanced Trees", modified in three ways:
 * a) this implementation allows for repeated scores.
 * b) the comparison is not just by key (our 'score') but by satellite data.
 * c) there is a back pointer, so it's a doubly linked list with the back
 *    pointers being only at "level 1". This allows to traverse the list
 *    from tail to head, useful for ZREVRANGE. */
```

**关键事实**：
1. ZSET 是**双结构**：hash table + skip list（不是单纯的跳表）
2. hash table 映射 Redis objects → scores（O(1) 单点查找）
3. skip list 映射 scores → Redis objects（O(logN) 有序操作）
4. SDS 字符串在两个结构间共享，节省内存
5. 跳表实现是 William Pugh 论文的 C 翻译，修改了 3 点
6. **注释中未直接对比 B+ 树**——只引用 Pugh 论文

## antirez 原话考证（已找到！）

**来源**：2010 年 3 月 6 日，antirez 在邮件列表/HN 的回复
**原始问题**：
> "Is there any particular reason you chose skip list instead of btrees except for simplicity? Skip lists consume more memory in pointers and are generally slower than btrees because of poor memory locality so traversing them means lots of cache misses."

**antirez 原话（英文原文）**：
> There are a few reasons:
>
> 1. They are not very memory intensive. It's up to you basically. Changing parameters about the probability of a node to have a given number of levels will make then less memory intensive than btrees.
>
> 2. A sorted set is often target of many ZRANGE or ZREVRANGE operations, that is, traversing the skip list as a linked list. With this operation the cache locality of skip lists is at least as good as with other kind of balanced trees.
>
> 3. They are simpler to implement, debug, and so forth. For instance thanks to the skip list simplicity I received a patch (already in Redis master) with augmented skip lists implementing ZRANK in O(log(N)). It required little changes to the code.

**关键发现**：

1. **antirez 确实把"实现简单"列为原因之一**——但这不是唯一原因，也不是主要原因
2. **antirez 的对比对象是 "btrees" 和 "other kind of balanced trees"**——不是专门对比 B+ 树
3. **antirez 说"跳表的 cache locality 至少和其他平衡树一样好"**——这个说法需要审视：
   - 他说的是 ZRANGE/ZREVRANGE 遍历场景（链表顺序扫描）
   - 对比对象是"其他平衡树"（如红黑树），不是 B+ 树
   - E2 benchmark 显示，在范围查询场景下，B+ 树确实显著快于跳表
4. **antirez 没有提到"并发优势"**——这是后人添加的误解！原文三条理由中**没有任何一条提到并发**

## 对市面流传说法的修正

| 市面说法 | antirez 原话 | 修正 |
|---------|-------------|------|
| "Redis 选跳表因为并发性能好" | 原文未提及并发 | ❌ 后人添加的误解 |
| "Redis 选跳表因为实现简单" | 原文第 3 点确实提到 | ✅ 但只是原因之一，且 antirez 排在第 3 |
| "Redis 选跳表因为内存占用少" | 原文第 1 点提到，但说法是"not very memory intensive" | ⚠️ 原文是说"可调参数后比 btree 省"，不是"必然省" |
| "Redis 选跳表因为范围查询快" | 原文第 2 点提到 cache locality | ⚠️ 原文说"至少和其他平衡树一样好"，但 E2 显示 B+ 树更快 |

## 关键函数（来自 t_zset.c）

### zslCreate
创建跳表，初始化 header 节点（32 层），level=1，length=0。

### zslInsert
插入节点的核心逻辑：
1. 从最高层开始，逐层下降查找插入位置
2. 记录每层的 update 节点和 rank
3. 调用 randomLevel() 决定新节点层数
4. 创建节点，更新各层 forward 指针和 span
5. 维护 backward 指针

### zslGetRangeByRank
范围查询的核心逻辑：
1. 从最高层下降到 rank 起点
2. 沿 level[0] 链表顺序遍历

**关键观察**：跳表的范围查询确实是"链表顺序扫描"，这正是 antirez 第 2 点说的场景。但在 B+ 树中，叶子节点的链表也是顺序扫描，且叶子节点内部是紧凑数组——所以 B+ 树的范围查询更快（E2 证实）。

## 结论

**antirez 的原话是诚实且具体的**，但市面流传的版本经过了简化和演绎：

1. "实现简单"是原因之一，但 antirez 排在第 3（不是主要原因）
2. "并发优势"是后人添加的，antirez 原文未提及
3. "内存占用少"被简化了——antirez 原意是"可调参数后可比 btree 省"
4. "cache locality"antirez 限定在"和其他平衡树比"（如红黑树），不是和 B+ 树比

**本篇论证的修正**：
- 不能再说"antirez 没说过实现简单"——他确实说过，但只是原因之一
- 应该呈现 antirez 原话的三条理由，然后逐一审视在现代视角下是否依然成立
- "并发优势论"明确是后人添加的误解，本篇要主动揭穿

# E2.1: Redis ZSET 滑动窗口日志 内存测算
# 证明：Sliding Window Log 在高 QPS 下内存代价惊人
# 环境：Python 3.9+，无外部依赖
# 运行：python3 sliding-window-memory.py
#
# 内存模型依据：Redis ZSET
#   - ziplist 编码（元素 < 128）：每个 entry ≈ 21 字节（prevlen + encoding + score 8B + member ~5B）
#   - skiplist 编码（元素 ≥ 128）：每个节点 ≈ 64 字节（含 zskiplistNode + dictEntry + SDS）
#   - 本计算使用 skiplist 编码（高 QPS 下必然超过 ziplist 阈值）
# [推演] 基于 Redis 官方内存模型公式，非实际 Redis 实测

SKIPLIST_ENTRY_BYTES = 64  # skiplist 节点开销（保守估计）
ZSET_OVERHEAD_BYTES = 200  # ZSET 对象头 + dict + skiplist 结构体

def calc_memory_per_key(qps, window_sec):
    """单个 key（单用户）的内存"""
    entries = qps * window_sec
    mem = ZSET_OVERHEAD_BYTES + entries * SKIPLIST_ENTRY_BYTES
    return mem, entries

def format_bytes(b):
    if b >= 1024**3:
        return f"{b / 1024**3:.1f} GB"
    elif b >= 1024**2:
        return f"{b / 1024**2:.1f} MB"
    elif b >= 1024:
        return f"{b / 1024:.1f} KB"
    return f"{b} B"

print("=" * 60)
print("E2.1: Redis ZSET 滑动窗口日志 内存测算 [推演]")
print("=" * 60)
print(f"模型：skiplist 编码，每节点 {SKIPLIST_ENTRY_BYTES} 字节")
print(f"窗口：60 秒")
print()

# 单 key 内存
print("--- 单用户（单 key）内存 ---")
print(f"{'QPS':<10} {'条目数':<12} {'单key内存':<12}")
print("-" * 34)
for qps in [100, 1000, 10000, 100000]:
    mem, entries = calc_memory_per_key(qps, 60)
    print(f"{qps:<10} {entries:<12,} {format_bytes(mem):<12}")

print()

# 多用户扩展
print("--- 多用户总内存（QPS × 用户数 × 60s 窗口）---")
print(f"{'QPS':<8} {'用户数':<10} {'总条目':<15} {'总内存':<12}")
print("-" * 50)
scenarios = [
    (100, 1000),
    (1000, 1000),
    (1000, 10000),
    (10000, 1000),
    (10000, 10000),
    (100000, 1000),
]
for qps, users in scenarios:
    mem_per_key, entries = calc_memory_per_key(qps, 60)
    total_mem = mem_per_key * users
    total_entries = entries * users
    print(f"{qps:<8} {users:<10,} {total_entries:<15,} {format_bytes(total_mem):<12}")

print()
print("--- 结论 ---")
print("• 1K QPS × 1K 用户 = 3.6 GB → 一台 Redis 能扛但吃力")
print("• 10K QPS × 1K 用户 = 35.8 GB → 超过常规 Redis 内存")
print("• 100K QPS × 1K 用户 = 357.6 GB → 完全不可行")
print()
print("对比：Sliding Window Counter（10 个子桶）")
print("  10K QPS × 1K 用户：1K × 10 个计数器 × 8B ≈ 80 KB")
print("  内存差距：35.8 GB vs 80 KB = 约 45 万倍")
print()
print("[推演] 以上为基于 Redis 内存模型的理论计算，非 Redis 实例实测")

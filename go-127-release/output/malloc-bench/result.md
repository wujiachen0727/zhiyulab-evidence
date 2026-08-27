# E5 小对象堆分配 bench（1.26 vs 1.27）

环境：darwin/arm64，Apple M4 Pro；测试在 `/tmp` 临时目录分别 pin go 1.26 / 1.27。

方法：`BenchmarkHeapAlloc{64,80}B`，结果写入全局 `sink` 防止栈分配消除 heap allocs。

| Benchmark | 1.26.4 ns/op | 1.27.0 ns/op | 变化 |
|-----------|-------------:|-------------:|-----:|
| HeapAlloc80B | ~14.5 | ~11.5 | ~21% ↓ |
| HeapAlloc64B | ~12.3 | ~10.2 | ~17% ↓ |

allocs/op 均为 1，B/op 分别为 80 / 64。

**限定**：本机单点 bench，仅说明 `<80B` 路径有可观差异；不得写成「全面加速 30%」。官方表述：小对象最高约 30%，分配密集程序整体约 1%。

代码：`evidence/code/malloc-bench/bench_test.go`

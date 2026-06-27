# E6 实测结果：panic+recover vs error return 性能对比

**测试环境**：Go 1.26.2, darwin/arm64 (Apple M4 Pro)
**测试方法**：go test -bench=. -benchmem -count=3

## 错误路径（触发错误时的性能）

| 方式 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| panic+recover | 157.0 | 48 | 2 |
| error return | 0.23 | 0 | 0 |

**结论**：panic+recover 比 error return 慢约 **670 倍**，且每次触发产生 2 次堆分配（48 字节）。error return 在错误路径上零分配。

## 正常路径（无错误时的性能）

| 方式 | ns/op | B/op | allocs/op |
|------|------:|-----:|----------:|
| 直接返回（无defer） | 0.24 | 0 | 0 |
| error return | 0.23 | 0 | 0 |

**结论**：正常路径下两者几乎没有区别。性能差异完全来自 panic 的栈展开和 defer 机制。

## 原始 Benchmark 输出

```
BenchmarkPanicRecover-14    	 6672399	       159.4 ns/op	      48 B/op	       2 allocs/op
BenchmarkPanicRecover-14    	 7719168	       154.5 ns/op	      48 B/op	       2 allocs/op
BenchmarkPanicRecover-14    	 7775743	       157.2 ns/op	      48 B/op	       2 allocs/op
BenchmarkErrorReturn-14     	1000000000	         0.2354 ns/op	       0 B/op	       0 allocs/op
BenchmarkErrorReturn-14     	1000000000	         0.2330 ns/op	       0 B/op	       0 allocs/op
BenchmarkErrorReturn-14     	1000000000	         0.2348 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoPanicPath-14     	1000000000	         0.2390 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoPanicPath-14     	1000000000	         0.2414 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoPanicPath-14     	1000000000	         0.2394 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoErrorPath-14     	1000000000	         0.2429 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoErrorPath-14     	1000000000	         0.2335 ns/op	       0 B/op	       0 allocs/op
BenchmarkNoErrorPath-14     	1000000000	         0.2344 ns/op	       0 B/op	       0 allocs/op
```

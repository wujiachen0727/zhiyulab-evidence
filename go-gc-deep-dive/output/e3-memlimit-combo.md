# E3：GOMEMLIMIT + GOGC 组合效果 — 实测数据

> 🟢已验证 — 本机实测，Go 1.26.2 darwin/arm64

## 实测结果

**场景**：逐步分配 20MB 对象（20000 个 1KB），观察 GC 触发时机

### GOGC=200（无 GOMEMLIMIT）

```
i=2000  | HeapAllocMB=2  | HeapSysMB=7  | NumGC=0
i=10000 | HeapAllocMB=10 | HeapSysMB=15 | NumGC=1  ← GC 在 10MB 时触发
i=18000 | HeapAllocMB=18 | HeapSysMB=23 | NumGC=1  ← 堆持续增长
释放后  | HeapAllocMB=0  | HeapSysMB=23 | NumGC=2
```

### GOGC=200 + GOMEMLIMIT=32MB

```
i=2000  | HeapAllocMB=2  | HeapSysMB=7  | NumGC=0
i=10000 | HeapAllocMB=10 | HeapSysMB=15 | NumGC=1
i=18000 | HeapAllocMB=18 | HeapSysMB=23 | NumGC=1  ← 堆在 GOMEMLIMIT 之下，行为相同
释放后  | HeapAllocMB=0  | HeapSysMB=27 | NumGC=2
```

### GOGC=off + GOMEMLIMIT=32MB

```
i=2000  | HeapAllocMB=2  | HeapSysMB=7  | NumGC=0
i=10000 | HeapAllocMB=10 | HeapSysMB=15 | NumGC=0  ← GOGC=off，不主动触发
i=18000 | HeapAllocMB=18 | HeapSysMB=23 | NumGC=0  ← 堆到 18MB，仍不触发
释放后  | HeapAllocMB=0  | HeapSysMB=27 | NumGC=1  ← 只有手动 GC
```

## 关键洞察

1. **GOGC=off + GOMEMLIMIT=32MB**：堆增长到 18MB 时仍未触发 GC（23MB < 32MB），说明 GOMEMLIMIT 在小堆场景下不会过早干预
2. **GOMEMLIMIT 的价值在大堆场景更明显**：当存活堆 > GOMEMLIMIT/2 时，GC 会更积极运行，防止 OOM
3. **GOGC=off 时 GC 完全不主动运行**：如果这个程序不是手动 GC，堆会持续增长直到 OOM。GOMEMLIMIT 是安全网

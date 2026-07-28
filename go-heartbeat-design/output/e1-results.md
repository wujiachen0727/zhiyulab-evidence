# E1: time.After / Timer / Ticker GC 行为对比实验结果（最终版）

## 核心发现

**Go 1.23 的 Timer/Ticker GC 改进是真实的，但受 go.mod 版本指令控制。**

go.mod 中 `go 1.23` 及以上：GC 能回收未引用的未触发 Timer（新行为）
go.mod 中 `go 1.22` 及以下：GC 不能回收（旧行为，兼容模式）

## 实验环境
- macOS Darwin arm64
- Go 1.22.12, 1.23.8, 1.24.13, 1.25.11, 1.26.4（via goenv, GOTOOLCHAIN=local）
- 50000 次 `time.After(1 * time.Hour)` 调用，函数作用域内创建释放
- 20 次 GC 循环后检查 HeapObjects

## 完整结果矩阵

| go.mod 版本 | Go 1.22 | Go 1.23 | Go 1.24 | Go 1.25 | Go 1.26 |
|:-----------:|:-------:|:-------:|:-------:|:-------:|:-------:|
| go 1.21 | 150,041 | 150,111 | 150,114 | 150,120 | 150,112 |
| go 1.22 | 150,040 | 150,107 | 150,114 | 150,117 | 150,109 |
| go 1.23 | N/A | **23** | **14** | **16** | **27** |

- **150K+ objects** = Timer 未被 GC 回收（泄漏）
- **14-27 objects** = Timer 被 GC 回收（正常）

## 关键洞察

1. **Go 1.23 确实修复了 Timer GC**——但只在 go.mod 声明 `go 1.23` 及以上时生效
2. **go.mod `go 1.22` 及以下触发兼容模式**——即使运行 Go 1.26，也使用旧 GC 行为
3. **大量生产项目 go.mod 仍写 `go 1.21/1.22`**——这些项目的心跳 Timer 仍然泄漏
4. **GODEBUG=asynctimerchan=1 可强制旧行为**——反向证明新行为需要 go.mod >= 1.23

## 机制解释

Go 1.23 通过 GODEBUG `asynctimerchan` 控制 Timer GC 行为：
- `asynctimerchan=0`（默认，go.mod >= 1.23）：GC 可回收未引用的未触发 Timer
- `asynctimerchan=1`（go.mod < 1.23）：旧行为，Timer 必须 Stop 或触发才能被 GC

Go 的 GODEBUG 兼容机制：go.mod 中的 `go 1.XX` 决定使用哪些 GODEBUG 默认值。`go 1.22` 的项目即使运行在 Go 1.26 上，也会启用 `asynctimerchan=1` 以保持行为兼容。

## 对文章立意的影响

原计划"一半坑已被语言修复"需要修正为更精确的表述：

**"Go 1.23 修复了 Timer/Ticker GC——但你的 go.mod 可能让你享受不到这个修复。"**

这是一个比"修复了"或"没修复"都更有趣的发现：
- 说"修复了"不准确——大量项目因为 go.mod 版本问题仍然泄漏
- 说"没修复"也不准确——go.mod >= 1.23 的项目确实不再泄漏
- 真相是"修复了，但兼容机制让很多项目无感知"

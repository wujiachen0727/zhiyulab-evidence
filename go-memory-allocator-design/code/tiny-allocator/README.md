# tiny-allocator

tiny allocator 合并分配实测（HeapObjects 对象数对比），验证小对象合并对 GC 扫描量的影响。

## 运行环境

- Go 1.26.2（darwin/arm64，Apple M4 Pro）
- 无第三方依赖（仅标准库）

## 运行方式

```bash
cd tiny-allocator
go test -v -run=TestTinyObjectCount -count=1
```

## 产出

- 运行输出见 `../output/tiny-allocator/`
- 数据解读见 `../../evidence/README.md`

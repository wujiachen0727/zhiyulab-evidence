# simple-allocator

简化 allocator 三版本对比（无缓存/每P缓存/大小类），验证缓存设计的权衡。

## 运行环境

- Go 1.26.2（darwin/arm64，Apple M4 Pro）
- 无第三方依赖（仅标准库）

## 运行方式

```bash
cd simple-allocator
go run .
```

## 产出

- 运行输出见 `../output/simple-allocator/`
- 数据解读见 `../../evidence/README.md`

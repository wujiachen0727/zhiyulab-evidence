# sizeclass-boundary

跨 size class 边界分配实测（32/33、64/65、128/129B），验证对象大小对分配开销的影响。

## 运行环境

- Go 1.26.2（darwin/arm64，Apple M4 Pro）
- 无第三方依赖（仅标准库）

## 运行方式

```bash
cd sizeclass-boundary
go test -bench=. -benchmem -benchtime=1s -run=^$
```

## 产出

- 运行输出见 `../output/sizeclass-boundary/`
- 数据解读见 `../../evidence/README.md`

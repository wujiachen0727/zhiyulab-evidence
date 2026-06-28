# E6：Bitmap vs Set vs Hash 内存对比

## 运行方式

```bash
cd evidence/code/storage-compare
go mod init storage-compare
go get github.com/redis/go-redis/v9
go run main.go
```

## 结论

连续 ID 场景 Bitmap 完胜（100 万用户 0.13MB vs 38MB）。稀疏 ID 场景 Bitmap 反而大 2.8 倍。详见 `../../data/bitmap-vs-set-vs-hash.md`。

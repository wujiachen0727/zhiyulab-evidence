# E5：BITCOUNT 大 key 耗时实测

## 运行方式

```bash
cd evidence/code/bitcount-latency
go mod init bitcount-latency
go get github.com/redis/go-redis/v9
go run main.go
```

## 结论

512MB Bitmap P99=8.83ms（本地直连）。详见 `../../data/bitcount-latency.md`。

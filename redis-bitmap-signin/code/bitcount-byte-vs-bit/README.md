# E2：BITCOUNT 字节偏移 vs 位偏移对比

## 运行方式

```bash
cd evidence/code/bitcount-byte-vs-bit
go mod init bitcount-byte-vs-bit
go get github.com/redis/go-redis/v9
go run main.go
```

## 结论

BITCOUNT 的 start/end 是字节偏移，不是位偏移。详见 `../../data/bitcount-byte-vs-bit.md`。

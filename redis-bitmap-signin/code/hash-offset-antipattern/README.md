# E4：hash offset 反模式复现

## 运行方式

```bash
cd evidence/code/hash-offset-antipattern
go mod init hash-offset-antipattern
go get github.com/redis/go-redis/v9
go run main.go
```

## 结论

10 个用户 + hash offset = 300-470MB，是连续 offset 的 600-900 万倍。详见 `../../output/hash-offset-antipattern/result.txt`。

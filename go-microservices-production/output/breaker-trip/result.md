# gobreaker gRPC 状态码误 trip 实验

| 策略 | 20 次 NotFound 后状态 | Trip 次数 |
|---|---|---:|
| naive (NotFound=失败) | open | 1 |
| correct (NotFound=成功) | closed | 0 |

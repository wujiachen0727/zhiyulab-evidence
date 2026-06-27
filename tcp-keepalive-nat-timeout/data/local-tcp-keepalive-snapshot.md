# 本机 TCP keepalive 参数快照

## 证据性质

[实测 macOS Darwin / sysctl]

本快照用于证明作者在论证阶段实际检查了当前机器的 TCP keepalive 参数。它不能替代 Linux 默认参数的权威来源；正文如果写 Linux 默认值，仍以 `tcp(7)` 文档为准。

## 环境预检

| 项目 | 结果 |
|---|---|
| Python | `/usr/bin/python3`，Python 3.9.6 |
| sysctl | `/usr/sbin/sysctl` |
| 操作系统 | macOS / Darwin |

## 原始命令

```bash
sysctl net.inet.tcp.keepidle net.inet.tcp.keepintvl net.inet.tcp.keepcnt
```

## 原始输出

```text
net.inet.tcp.keepidle: 7200000
net.inet.tcp.keepintvl: 75000
net.inet.tcp.keepcnt: 8
```

## 换算

| 参数 | 原始值 | 换算 | 含义 |
|---|---:|---:|---|
| `net.inet.tcp.keepidle` | 7200000 ms | 7200s / 2h | 空闲多久后首次探测 |
| `net.inet.tcp.keepintvl` | 75000 ms | 75s | 探测间隔 |
| `net.inet.tcp.keepcnt` | 8 | 8 次 | 放弃前探测次数 |

## 可供正文引用的结论

1. 本机 macOS 上也能看到“2 小时首次探测、75 秒探测间隔”这组默认量级。
2. `keepcnt` 与 Linux 文档中的默认值不同：本机为 8，Linux `tcp(7)` 文档为 9。正文不能把本机实测直接写成 Linux 事实。
3. 这反而提醒读者：跨平台参数名、单位和默认值可能不同，排查时必须查运行环境自己的配置。

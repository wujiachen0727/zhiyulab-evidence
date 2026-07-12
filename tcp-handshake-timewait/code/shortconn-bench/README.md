# 短连接压测 / 端口耗尽 / 连接复用对比

本目录复现文章 E1、E3、E4 三组实验。

## 环境
- macOS Darwin arm64 / Python 3.9.6（实测机）
- Linux 读者可将 `netstat -an -p tcp` 换成 `ss -tan`

## 复现步骤
```bash
# E1：2000 次短连接后 TIME_WAIT 计数飙升
python3 server.py bench 2000

# E3：客户端固定源端口反复短连接，复现本地端口耗尽
#   注意：macOS/BSD 默认允许 TIME_WAIT 端口复用，本机不会报 EADDRINUSE；
#   Linux 默认语义下会直接报 address already in use。
python3 client_port_exhaust.py

# E4：短连接 vs 长连接复用 对 TIME_WAIT 数量的影响
python3 e4_compare.py
```

## 文件说明
- `server.py`：本地回环 echo 服务，支持 bench 子命令批量发起短连接
- `client_port_exhaust.py`：固定源端口反复短连接，复现端口耗尽
- `e4_compare.py`：短连接 / 长连接复用两种模式对比
- `port_exhaust.go`：E3 的 Go 复现版本（最终采用 Python 版诚实标注内核差异）
- `e3_output.txt` / `e3_result.txt`：E3 原始输出与结论
- `e4_result.txt`：E4 原始输出与结论

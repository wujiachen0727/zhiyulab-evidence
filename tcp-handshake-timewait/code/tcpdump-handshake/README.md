# 连接状态机观测（E2）

本目录复现文章 E2 实验：观测一条短连接从 ESTABLISHED 到 TIME_WAIT 的状态转换。

## 环境
- macOS Darwin arm64 / Python 3.9.6（实测机）
- 无抓包权限时，用 `netstat` 高频采样替代 `tcpdump`

## 复现步骤
```bash
python3 observe_states.py
```

## 文件说明
- `observe_states.py`：建立一条连接并主动 close，随后高频采样 netstat，记录出现过的 TCP 状态集合
- 上游 `output/tcpdump-handshake/e2_result.txt` 为原始输出与结论

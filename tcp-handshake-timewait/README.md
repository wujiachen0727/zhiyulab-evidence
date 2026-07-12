# 论据总索引（evidence）

> 文章：面试背得出三次握手，为什么还是不懂 TCP
> slug：tcp-handshake-timewait
> 自造论据类型：实验验证（E1-E4）+ 场景模拟（E5）+ 逻辑推演（E7-E8）
> 外部引用：R1（tcp_tw_recycle 移除）/ R2（RFC 793 2MSL）

## 独立论据清单

| ID | 类型 | 一句话描述 | 自造/引用 | 产出路径 |
|----|------|-----------|:--------:|---------|
| E1 | 实验验证 | 2000 次短连接后 TIME_WAIT 从 22 飙到 2049 | 自造 | `output/shortconn-bench/e1_result.txt` |
| E2 | 实验验证 | netstat 采样到完整挥手状态链，确认主动关闭方进入 TIME_WAIT | 自造 | `output/tcpdump-handshake/e2_result.txt` |
| E3 | 实验验证 | 客户端固定源端口复现本地端口耗尽（含 Linux/macOS 内核语义差异诚实标注） | 自造 | `output/shortconn-bench/e3_result.txt` |
| E4 | 实验验证 | 短连接 2041 个 vs 长连接复用 6 个 TIME_WAIT，差 340 倍 | 自造 | `output/shortconn-bench/e4_result.txt` |
| E5 | 场景模拟 | 面试场景：三类误答拆解"只背握手"的认知断层 | 自造 | 融入正文（第1章） |
| E7 | 逻辑推演 | 2MSL 两个理由（旧报文防串 + 对端 ACK 必达）辩证推演 | 自造 | 融入正文（第4章） |
| E8 | 逻辑推演 | tcp_tw_recycle 被移除（NAT 时间戳冲突）反证"盲目消除 TIME_WAIT 危险" | 自造 | 融入正文（第4章） |
| R1 | 外部引用 | Linux 4.12 移除 tcp_tw_recycle + NAT 时间戳冲突 | 引用 | 行内引用，带"我怎么看" |
| R2 | 外部引用 | RFC 793 定义 2MSL | 引用 | 行内引用，带"我怎么看" |

## 自造度统计

- 独立论据合计：9 项（自造 7 + 引用 2）
- 自造占比：7/9 ≈ **78%**（目标 ≥ 70%，达标）
- 外部引用：2 处（≤ 3 硬上限，达标）
- 引用依赖度：去掉 R1/R2 后核心观点（握手≠懂连接生命周期、TIME_WAIT 是设计必然）仍成立 ✅

## 实验环境

- 系统：macOS Darwin arm64
- Python：3.9.6
- Go：1.26.4（E3 曾用 Go 复现，最终采用 Python 版诚实标注内核差异）

## 可复现命令

```bash
# E1 / E3 / E4：短连接 / 端口耗尽 / 连接复用对比
cd evidence/code/shortconn-bench
python3 server.py bench 2000              # E1
python3 client_port_exhaust.py            # E3
python3 e4_compare.py                      # E4

# E2：观测连接状态机到 TIME_WAIT
cd evidence/code/tcpdump-handshake
python3 observe_states.py                 # E2
```

> 注意：macOS 无 `ss` 命令，统一用 `netstat -an -p tcp` 等价观测；Linux 读者可换 `ss -tan`。
> 端口耗尽（E3）的 `EADDRINUSE` 在 Linux 默认语义下才会复现，本机 macOS 已诚实标注未复现。

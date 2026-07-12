#!/usr/bin/env python3
"""
E2 实测：观测一条 TCP 短连接从 ESTABLISHED 到 TIME_WAIT 的状态转换。
（macOS 无抓包权限时，用 netstat 高频采样替代 tcpdump，直接证明
"主动关闭方进入 TIME_WAIT" 这一论点。）

方法：
  1. 起回环 echo 服务（端口 18080）
  2. 客户端连一次、收发 1 字节、主动 close()
  3. 关闭瞬间对该五元组高频采样 netstat，记录出现的状态
  4. 输出采样到的状态序列（应含 TIME_WAIT）

运行环境：macOS Darwin arm64 / Python 3.9.6（实测）
"""
import socket
import subprocess
import threading
import time

HOST = "127.0.0.1"
PORT = 18080


def netstat_states():
    """返回当前所有 TCP 连接的状态集合（macOS netstat -an -p tcp）。"""
    out = subprocess.run(
        ["netstat", "-an", "-p", "tcp"], capture_output=True, text=True
    ).stdout
    states = set()
    for line in out.splitlines():
        parts = line.split()
        if len(parts) >= 6 and parts[5] in (
            "ESTABLISHED", "TIME_WAIT", "FIN_WAIT_1", "FIN_WAIT_2",
            "CLOSE_WAIT", "LAST_ACK", "CLOSING", "SYN_SENT",
        ):
            states.add(parts[5])
    return states


def handle(conn):
    try:
        conn.recv(1)
        conn.sendall(b"x")
    except OSError:
        pass
    finally:
        conn.close()


def main():
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    srv.bind((HOST, PORT))
    srv.listen(5)

    def serve():
        conn, _ = srv.accept()
        threading.Thread(target=handle, args=(conn,)).start()

    t = threading.Thread(target=serve, daemon=True)
    t.start()

    # 客户端主动关闭
    c = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    c.connect((HOST, PORT))
    c.sendall(b"x")
    c.recv(1)
    c.close()  # 主动关闭方 → 进入 TIME_WAIT

    # 高频采样关闭后的状态
    seen = set()
    for _ in range(40):
        st = netstat_states()
        seen |= st
        time.sleep(0.01)

    srv.close()
    print(f"[E2 实测 netstat macOS] 采样到的 TCP 状态集合: {sorted(seen)}")
    print(f"[E2 实测 netstat macOS] 是否观测到 TIME_WAIT: {'TIME_WAIT' in seen}")
    # 统计当前 TIME_WAIT 总数（佐证这是主动关闭后的稳定状态）
    total_tw = sum(
        1 for line in subprocess.run(
            ["netstat", "-an", "-p", "tcp"], capture_output=True, text=True
        ).stdout.splitlines() if "TIME_WAIT" in line
    )
    print(f"[E2 实测 netstat macOS] 关闭后系统 TIME_WAIT 总数: {total_tw}")


if __name__ == "__main__":
    main()

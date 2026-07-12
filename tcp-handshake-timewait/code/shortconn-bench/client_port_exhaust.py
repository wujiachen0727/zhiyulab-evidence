#!/usr/bin/env python3
"""
E3 实测：客户端固定源端口反复发起短连接，复现本地端口耗尽（EADDRINUSE）。

核心论点支撑：高并发短连接客户端（爬虫、压测、HTTP/1.0 close 客户端）
的"本地端口耗尽"不是服务端端口不够，而是自己的 ephemeral 端口
被 TIME_WAIT 占满 —— 这是 TIME_WAIT 最常被低估的真实代价。

方法：
  1. 起回环 echo 服务（端口 18090）
  2. 客户端每次都 bind 到固定源端口 18092，连 18090，收发 1 字节
  3. 主动 close 后立即用同端口再 bind，记录成功 / 报 EADDRINUSE
  4. 把结果直接写入 e3_result.txt（不依赖 shell 重定向，避免缓冲丢失）

运行环境：macOS Darwin arm64 / Python 3.9.6（实测）
诚实记录：macOS/BSD 默认允许 TIME_WAIT 端口被新连接复用，因此原生可能
不复现"失败"；程序会如实记录两种结果，正文据此标注 Linux 与 macOS 的内核差异。
"""
import socket
import threading
import time
import os

HOST = "127.0.0.1"
SRV_PORT = 18090
SRC_PORT = 18092  # 固定客户端源端口，用于复现端口耗尽

RESULT_PATH = os.path.join(os.path.dirname(__file__), "e3_result.txt")


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
    srv.bind((HOST, SRV_PORT))
    srv.listen(5)

    def serve():
        for _ in range(60):
            try:
                conn, _ = srv.accept()
            except OSError:
                break
            threading.Thread(target=handle, args=(conn,)).start()

    threading.Thread(target=serve, daemon=True).start()
    time.sleep(0.2)

    lines = []
    succeeded = 0
    first_err = None
    for i in range(60):
        try:
            c = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            c.bind((HOST, SRC_PORT))  # 固定源端口
            c.connect((HOST, SRV_PORT))
            c.sendall(b"x")
            c.recv(1)
            c.close()
            succeeded += 1
        except OSError as e:
            if first_err is None:
                first_err = str(e)
            lines.append(f"[E3 实测 Py] 第 {i + 1} 次用固定源端口 {SRC_PORT} 失败: {e}")
            break
        time.sleep(0.02)

    lines.insert(0, f"[E3 实测 Py] 成功完成 {succeeded} 次; 首个错误={first_err}")
    lines.append(
        "[E3 实测 Py] 结论: 本机内核"
        + ("禁止 TIME_WAIT 端口复用 → 复现 EADDRINUSE（端口耗尽）。"
           if first_err else
           "默认允许 TIME_WAIT 端口被新连接复用 → 未复现 EADDRINUSE。"
           " Linux 默认语义下（未设 tcp_tw_reuse），此场景会直接报 address already in use。")
    )
    with open(RESULT_PATH, "w") as f:
        f.write("\n".join(lines) + "\n")
    print("\n".join(lines))


if __name__ == "__main__":
    main()

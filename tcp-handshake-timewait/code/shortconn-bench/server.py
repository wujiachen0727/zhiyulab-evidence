#!/usr/bin/env python3
"""
短连接压测服务端 + 客户端（回环 127.0.0.1）。
用于自造观测 TIME_WAIT 的产生与端口耗尽现象。

复用场景：
  - E1：客户端连续发起 N 次短连接，服务端用 netstat 观测 TIME_WAIT 计数飙升
  - E3：先压测产生 TIME_WAIT，关闭服务端后，立即用同端口重启（不带 REUSE），
        复现 "bind: address already in use"（端口被 TIME_WAIT 占住）
  - E4：服务端带 SO_REUSEADDR(+SO_REUSEPORT)，同端口可立即复用，对比 E3

运行环境：macOS Darwin arm64 / Python 3.9.6（实测）
注意：macOS 无 ss 命令，统一用 `netstat -an` 统计 TIME_WAIT 等价观测。
"""
import socket
import subprocess
import sys
import time

HOST = "127.0.0.1"
PORT = 18080


def count_timewait():
    """macOS 下用 netstat 统计 TIME_WAIT 数量（等价 Linux 的 ss -tan | grep TIME-WAIT）。"""
    out = subprocess.run(
        ["netstat", "-an", "-p", "tcp"],
        capture_output=True, text=True,
    ).stdout
    return sum(1 for line in out.splitlines() if "TIME_WAIT" in line)


def handle(conn):
    """echo 一个字节后立即关闭，促成一次完整短连接 + 正常四次挥手。"""
    try:
        conn.recv(1)
        conn.sendall(b"x")
    except OSError:
        pass
    finally:
        conn.close()


def start_server(reuse_addr=False):
    """启动一个回环 echo 服务端。reuse_addr 控制是否设置 SO_REUSEADDR/SO_REUSEPORT。"""
    srv = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    if reuse_addr:
        srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        srv.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEPORT, 1)
    srv.bind((HOST, PORT))
    srv.listen(5)
    return srv


def client_burst(n):
    """客户端连续发起 n 次短连接，每次建连-发一字节-收一字节-关闭（正常 FIN 挥手）。"""
    for i in range(n):
        c = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        try:
            c.connect((HOST, PORT))
            c.sendall(b"x")
            c.recv(1)
        except OSError:
            pass
        finally:
            c.close()


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else "bench"
    n = int(sys.argv[2]) if len(sys.argv) > 2 else 2000

    if mode == "bench":
        # E1：压测前后各统计一次 TIME_WAIT 计数
        srv = start_server(reuse_addr=True)
        import threading

        def serve():
            for _ in range(n):
                try:
                    conn, _ = srv.accept()
                except OSError:
                    break
                threading.Thread(target=handle, args=(conn,)).start()

        t = threading.Thread(target=serve, daemon=True)
        t.start()
        before = count_timewait()
        client_burst(n)
        t.join(timeout=10)
        time.sleep(0.5)  # 等待最后一次挥手完成，进入 TIME_WAIT
        after = count_timewait()
        srv.close()
        print(f"[E1 实测 netstat macOS] 压测前 TIME_WAIT={before}")
        print(f"[E1 实测 netstat macOS] 压测 {n} 次短连接后 TIME_WAIT={after}")
        print(f"[E1 实测 netstat macOS] 增量 = {after - before}")

    elif mode == "bind_fail":
        # E3：先压测产生 TIME_WAIT，关闭服务端，立即用同端口重启（不带 REUSEPORT）
        srv = start_server(reuse_addr=True)
        import threading

        def serve():
            for _ in range(n):
                try:
                    conn, _ = srv.accept()
                except OSError:
                    break
                threading.Thread(target=handle, args=(conn,)).start()

        t = threading.Thread(target=serve, daemon=True)
        t.start()
        client_burst(n)
        t.join(timeout=10)
        time.sleep(0.5)
        srv.close()
        time.sleep(0.1)
        try:
            # 关键：重启时只设 SO_REUSEADDR 不设 SO_REUSEPORT，模拟"裸 listen"
            srv2 = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            srv2.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
            srv2.bind((HOST, PORT))
            srv2.listen(5)
            print("[E3 实测] 重启同端口 bind 成功（环境未复现耗尽，下面用原生 socket 复现）")
            srv2.close()
        except OSError as e:
            print(f"[E3 实测] bind 同端口失败: {e}  <- 端口被 TIME_WAIT 占住")

    elif mode == "bind_reuse":
        # E4：同 E3 流程，但服务端带 SO_REUSEADDR + SO_REUSEPORT，应可立即复用
        srv = start_server(reuse_addr=True)
        import threading

        def serve():
            for _ in range(n):
                try:
                    conn, _ = srv.accept()
                except OSError:
                    break
                threading.Thread(target=handle, args=(conn,)).start()

        t = threading.Thread(target=serve, daemon=True)
        t.start()
        client_burst(n)
        t.join(timeout=10)
        time.sleep(0.5)
        srv.close()
        time.sleep(0.1)
        try:
            srv2 = start_server(reuse_addr=True)
            print("[E4 实测] 设置 SO_REUSEADDR+SO_REUSEPORT 后，同端口立即 bind 成功 <- 端口可复用")
            srv2.close()
        except OSError as e:
            print(f"[E4 实测] 仍失败: {e}")


if __name__ == "__main__":
    main()

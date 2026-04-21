package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// TCP echo server + client benchmark
// 对比默认配置 vs 优化配置（TCP_NODELAY + 调缓冲区）的延迟和吞吐

func startServer(addr string, optimized bool) net.Listener {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "listen error: %v\n", err)
		os.Exit(1)
	}
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				if optimized {
					if tc, ok := c.(*net.TCPConn); ok {
						tc.SetNoDelay(true)
						tc.SetReadBuffer(64 * 1024)
						tc.SetWriteBuffer(64 * 1024)
					}
				}
				buf := make([]byte, 4096)
				for {
					n, err := c.Read(buf)
					if err != nil {
						return
					}
					_, err = c.Write(buf[:n])
					if err != nil {
						return
					}
				}
			}(conn)
		}
	}()
	return ln
}

func runBenchmark(addr string, concurrency int, duration time.Duration, optimized bool, msgSize int) (totalReqs int64, avgLatency time.Duration, p99Latency time.Duration) {
	var ops atomic.Int64
	var totalLatNs atomic.Int64
	latencies := make([]int64, 0, 100000)
	var mu sync.Mutex
	var wg sync.WaitGroup

	msg := make([]byte, msgSize)
	for i := range msg {
		msg[i] = 'A'
	}

	deadline := time.Now().Add(duration)
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			conn, err := net.Dial("tcp", addr)
			if err != nil {
				fmt.Fprintf(os.Stderr, "dial error: %v\n", err)
				return
			}
			defer conn.Close()
			if optimized {
				if tc, ok := conn.(*net.TCPConn); ok {
					tc.SetNoDelay(true)
					tc.SetReadBuffer(64 * 1024)
					tc.SetWriteBuffer(64 * 1024)
				}
			}
			buf := make([]byte, msgSize)
			for time.Now().Before(deadline) {
				start := time.Now()
				_, err := conn.Write(msg)
				if err != nil {
					return
				}
				_, err = io.ReadFull(conn, buf)
				if err != nil {
					return
				}
				lat := time.Since(start).Nanoseconds()
				ops.Add(1)
				totalLatNs.Add(lat)
				mu.Lock()
				latencies = append(latencies, lat)
				mu.Unlock()
			}
		}()
	}
	wg.Wait()

	total := ops.Load()
	if total == 0 {
		return 0, 0, 0
	}
	avg := time.Duration(totalLatNs.Load() / total)

	// 简单排序取 P99
	mu.Lock()
	defer mu.Unlock()
	// 用简单选择找 P99 位置
	n := len(latencies)
	if n == 0 {
		return total, avg, 0
	}
	// 排序
	sortInt64s(latencies)
	p99Idx := int(float64(n) * 0.99)
	if p99Idx >= n {
		p99Idx = n - 1
	}
	return total, avg, time.Duration(latencies[p99Idx])
}

func sortInt64s(a []int64) {
	// 简单快排
	if len(a) < 2 {
		return
	}
	pivot := a[len(a)/2]
	left, right := 0, len(a)-1
	for left <= right {
		for a[left] < pivot {
			left++
		}
		for a[right] > pivot {
			right--
		}
		if left <= right {
			a[left], a[right] = a[right], a[left]
			left++
			right--
		}
	}
	if right > 0 {
		sortInt64s(a[:right+1])
	}
	if left < len(a) {
		sortInt64s(a[left:])
	}
}

func main() {
	concurrency := flag.Int("c", 100, "并发连接数")
	duration := flag.Int("d", 5, "测试时长(秒)")
	msgSize := flag.Int("m", 128, "消息大小(字节)")
	flag.Parse()

	dur := time.Duration(*duration) * time.Second

	fmt.Printf("=== TCP Echo Benchmark ===\n")
	fmt.Printf("并发: %d | 时长: %ds | 消息: %d bytes\n\n", *concurrency, *duration, *msgSize)

	// 默认配置
	ln1 := startServer("127.0.0.1:0", false)
	addr1 := ln1.Addr().String()
	time.Sleep(100 * time.Millisecond)
	reqs1, avg1, p99_1 := runBenchmark(addr1, *concurrency, dur, false, *msgSize)
	ln1.Close()
	time.Sleep(200 * time.Millisecond)

	// 优化配置
	ln2 := startServer("127.0.0.1:0", true)
	addr2 := ln2.Addr().String()
	time.Sleep(100 * time.Millisecond)
	reqs2, avg2, p99_2 := runBenchmark(addr2, *concurrency, dur, true, *msgSize)
	ln2.Close()

	qps1 := float64(reqs1) / float64(*duration)
	qps2 := float64(reqs2) / float64(*duration)
	improvement := (qps2 - qps1) / qps1 * 100

	fmt.Printf("%-20s %15s %15s\n", "", "默认配置", "优化配置")
	fmt.Printf("%-20s %15d %15d\n", "总请求数", reqs1, reqs2)
	fmt.Printf("%-20s %15.0f %15.0f\n", "QPS", qps1, qps2)
	fmt.Printf("%-20s %15s %15s\n", "平均延迟", avg1.String(), avg2.String())
	fmt.Printf("%-20s %15s %15s\n", "P99 延迟", p99_1.String(), p99_2.String())
	fmt.Printf("\nQPS 提升: %.1f%%\n", improvement)
	fmt.Printf("P99 延迟变化: %s → %s\n", p99_1.String(), p99_2.String())
}

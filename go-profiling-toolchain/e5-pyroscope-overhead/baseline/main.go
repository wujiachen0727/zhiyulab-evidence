// E5 Pyroscope 开销实测 —— 基线版（无 Pyroscope agent）
//
// 和 agent 版完全相同的业务逻辑，唯一区别是不挂 Pyroscope Go SDK。
// 用于建立"没有持续 profiling"的性能基线。
//
// 运行：go run main.go
// 默认端口：6061
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
)

var reqCount int64

// 业务函数：和 E3 的快路径相同——模拟典型的 HTTP 请求 CPU 工作
func businessWork() {
	sum := 0
	for i := 0; i < 50000; i++ {
		sum += i * i
	}
	_ = sum
}

func handler(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&reqCount, 1)
	businessWork()
	fmt.Fprintln(w, "ok")
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "requests: %d\n", atomic.LoadInt64(&reqCount))
}

func main() {
	http.HandleFunc("/work", handler)
	http.HandleFunc("/stats", statsHandler)

	log.Printf("=== E5 基线服务启动（无 Pyroscope agent）===")
	log.Printf("监听: :6061")
	log.Printf("端点: /work, /stats, /debug/pprof")

	if err := http.ListenAndServe(":6061", nil); err != nil {
		log.Fatal(err)
	}
}

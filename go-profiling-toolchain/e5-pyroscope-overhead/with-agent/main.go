// E5 Pyroscope 开销实测 —— 带 agent 版（Pyroscope Go SDK 推送）
//
// 和基线版完全相同的业务逻辑，唯一区别是挂上 Pyroscope Go SDK（push mode）。
// Pyroscope agent 会在后台周期性（默认 15s）抓取 pprof 并推送到 Pyroscope server。
//
// 依赖：github.com/grafana/pyroscope-go
// 运行：go run main.go
// 默认端口：6062（区别于基线版 6061）
package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync/atomic"

	"github.com/grafana/pyroscope-go"
)

var reqCount int64

// 业务函数——和基线版完全相同
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
	// 启动 Pyroscope agent（push mode）
	// 默认配置：CPU profile + memory profile + mutex profile + block profile
	//           上传频率 15s/次
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: "e5-overhead-test",
		ServerAddress:   "http://localhost:4040",
		Logger:          pyroscope.StandardLogger,
		Tags:            map[string]string{"experiment": "e5"},
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		log.Fatalf("Pyroscope start failed: %v", err)
	}

	http.HandleFunc("/work", handler)
	http.HandleFunc("/stats", statsHandler)

	log.Printf("=== E5 agent 服务启动（带 Pyroscope agent）===")
	log.Printf("监听: :6062")
	log.Printf("Pyroscope server: http://localhost:4040")
	log.Printf("端点: /work, /stats, /debug/pprof")
	log.Printf("Pyroscope agent: %v profile types 已开启", 10)

	log.Printf("PID: %d", os.Getpid())

	if err := http.ListenAndServe(":6062", nil); err != nil {
		log.Fatal(err)
	}
}

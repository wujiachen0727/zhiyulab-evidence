// SSE vs WebSocket 资源消耗基准测试
//
// 测试目的：对比 SSE 和 WebSocket 在同等推送量下的资源消耗
// 测试环境：Go 1.22+, darwin/arm64
// 测试方法：模拟 1000 个并发连接，测量 CPU、内存和吞吐量
// 注意：这是模拟测试，使用 time.Sleep 模拟网络延迟
//
// 使用方法：
//   go test -bench=. -benchmem ./... 2>&1 | tee ../output/benchmark-results.txt

package benchmark

import (
	"bufio"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

// ============================================================
// SSE 服务端 + 客户端模拟
// ============================================================

// sseHandler 模拟 SSE 推送服务端
func sseHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 模拟推送 100 条事件
	for i := 0; i < 100; i++ {
		fmt.Fprintf(w, "data: {\"id\":%d,\"message\":\"event-%d\",\"timestamp\":%d,\"payload\":\"%s\"}\n\n",
			i, i, time.Now().UnixMilli(), strings.Repeat("x", 150))
		flusher.Flush()
		time.Sleep(1 * time.Millisecond) // 模拟网络延迟
	}
}

// sseClient 模拟 SSE 客户端连接
func sseClient(baseURL string) {
	resp, err := http.Get(baseURL + "/sse")
	if err != nil {
		return
	}
	defer resp.Body.Close()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		// 读取事件数据
		_ = scanner.Text()
	}
}

// ============================================================
// WebSocket 服务端 + 客户端模拟（基于 net/http 的简易 WS）
// ============================================================

// wsHandler 模拟 WebSocket 升级 + 推送
func wsHandler(w http.ResponseWriter, r *http.Request) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	conn, bufrw, err := hj.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// 发送 HTTP 101 Switching Protocols 响应
	bufrw.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
	bufrw.WriteString("Upgrade: websocket\r\n")
	bufrw.WriteString("Connection: Upgrade\r\n")
	bufrw.WriteString("\r\n")
	bufrw.Flush()

	// 模拟推送 100 条消息（简易文本帧格式）
	for i := 0; i < 100; i++ {
		payload := fmt.Sprintf("{\"id\":%d,\"message\":\"event-%d\",\"timestamp\":%d,\"payload\":\"%s\"}",
			i, i, time.Now().UnixMilli(), strings.Repeat("x", 150))
		frame := createTextFrame([]byte(payload))
		conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		_, err := conn.Write(frame)
		if err != nil {
			return
		}
		time.Sleep(1 * time.Millisecond) // 模拟网络延迟
	}
}

// createTextFrame 创建 WebSocket 文本帧（简化版，无 mask）
func createTextFrame(payload []byte) []byte {
	frame := []byte{0x81} // FIN + text opcode
	length := len(payload)
	if length < 126 {
		frame = append(frame, byte(length))
	} else if length < 65536 {
		frame = append(frame, 126, byte(length>>8), byte(length&0xFF))
	} else {
		frame = append(frame, 127)
		for i := 7; i >= 0; i-- {
			frame = append(frame, byte(length>>(i*8)))
		}
	}
	frame = append(frame, payload...)
	return frame
}

// ============================================================
// Benchmark 测试
// ============================================================

// BenchmarkSSE 测试 SSE 服务端在并发下的资源消耗
func BenchmarkSSE(b *testing.B) {
	sseSrv := httptest.NewServer(http.HandlerFunc(sseHandler))
	defer sseSrv.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		numClients := 1000

		var memStatsStart, memStatsEnd runtime.MemStats
		runtime.ReadMemStats(&memStatsStart)

		for j := 0; j < numClients; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sseClient(sseSrv.URL)
			}()
		}
		wg.Wait()

		runtime.ReadMemStats(&memStatsEnd)
		_ = memStatsEnd.TotalAlloc - memStatsStart.TotalAlloc
	}
}

// BenchmarkWebSocket 测试 WebSocket 服务端在并发下的资源消耗
func BenchmarkWebSocket(b *testing.B) {
	wsSrv := httptest.NewServer(http.HandlerFunc(wsHandler))
	defer wsSrv.Close()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		var wg sync.WaitGroup
		numClients := 1000

		var memStatsStart, memStatsEnd runtime.MemStats
		runtime.ReadMemStats(&memStatsStart)

		for j := 0; j < numClients; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				resp, err := http.Get(wsSrv.URL + "/ws")
				if err != nil {
					return
				}
				buf := make([]byte, 4096)
				for {
					_, err := resp.Body.Read(buf)
					if err != nil {
						break
					}
				}
				resp.Body.Close()
			}()
		}
		wg.Wait()

		runtime.ReadMemStats(&memStatsEnd)
		_ = memStatsEnd.TotalAlloc - memStatsStart.TotalAlloc
	}
}

func init() {
	fmt.Println("=== 环境信息 ===")
	fmt.Printf("Go 版本: %s\n", runtime.Version())
	fmt.Printf("CPU: %d 核\n", runtime.NumCPU())
	fmt.Printf("GoMaxProcs: %d\n", runtime.GOMAXPROCS(0))
}

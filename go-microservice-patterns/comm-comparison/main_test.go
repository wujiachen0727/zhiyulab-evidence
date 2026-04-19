// E3+E5: 同一个"扣库存"操作的三种通信方式对比
// [实测 Go 1.26.2] 对比 HTTP/JSON、模拟 gRPC（protobuf 序列化）、模拟异步（channel）
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ========== 公共数据结构 ==========

type DeductRequest struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

type DeductResponse struct {
	Success bool   `json:"success"`
	Remain  int    `json:"remain"`
	Error   string `json:"error,omitempty"`
}

// ========== 方式1: HTTP/JSON ==========

func httpHandler(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req DeductRequest
	json.Unmarshal(body, &req)

	resp := DeductResponse{Success: true, Remain: 99}
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func callHTTP(server *httptest.Server) (resp DeductResponse, err error) {
	req := DeductRequest{SKU: "sku-001", Qty: 1}
	data, _ := json.Marshal(req)

	httpResp, err := http.Post(server.URL+"/deduct", "application/json", strings.NewReader(string(data)))
	if err != nil {
		return
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)
	json.Unmarshal(body, &resp)
	return
}

// ========== 方式2: 模拟 gRPC（直接函数调用 + 序列化/反序列化模拟二进制开销） ==========
// 真实 gRPC 需要 proto 定义和代码生成，这里模拟核心差异点：序列化方式

func grpcLikeCall() (resp DeductResponse, err error) {
	// 模拟 protobuf 序列化（用 json 近似，实际 protobuf 更快）
	req := DeductRequest{SKU: "sku-001", Qty: 1}
	data, _ := json.Marshal(req) // 实际 gRPC 用 protobuf，更快
	_ = data

	// 模拟服务端处理
	resp = DeductResponse{Success: true, Remain: 99}
	return
}

// ========== 方式3: 异步消息（channel 模拟） ==========

func asyncCall(ch chan DeductRequest, resultCh chan DeductResponse) {
	ch <- DeductRequest{SKU: "sku-001", Qty: 1}
	<-resultCh // 等待结果
}

// ========== Benchmarks ==========

func BenchmarkHTTP(b *testing.B) {
	mux := http.NewServeMux()
	mux.HandleFunc("/deduct", httpHandler)
	server := httptest.NewServer(mux)
	defer server.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		callHTTP(server)
	}
}

func BenchmarkGRPCLike(b *testing.B) {
	for i := 0; i < b.N; i++ {
		grpcLikeCall()
	}
}

func BenchmarkAsync(b *testing.B) {
	ch := make(chan DeductRequest, 1)
	resultCh := make(chan DeductResponse, 1)

	// 模拟消费者
	go func() {
		for range ch {
			resultCh <- DeductResponse{Success: true, Remain: 99}
		}
	}()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		asyncCall(ch, resultCh)
	}
}

func main() {
	fmt.Println("请使用 go test -bench=. -benchmem 运行")
}

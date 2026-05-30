// Package main 实验 E2：端到端调用延迟分解
//
// 目的：把一次"客户端调用 → 服务端响应"的全过程拆成可测量的几段，
// 量化各环节占总耗时的比例，从而看清楚"协议层 + 序列化"占多少。
//
// 分段：
//   1. TCP 建连（含 DNS 跳过，因为是 127.0.0.1）
//   2. 请求序列化（JSON Marshal）
//   3. 网络往返（写入 socket → 读出 socket）
//   4. 响应反序列化（JSON Unmarshal）
//   5. 服务端业务逻辑（模拟一次"DB 查询"——sleep 10ms 代表数据库 I/O）
//
// 结论指标：协议+序列化 / 总耗时 = 多少百分比
//
// 注意：业务逻辑用 sleep 10ms 模拟，这是"中等量级 DB/RPC 下游调用"的
// 典型耗时。读者可以把它视作 1 次缓存读 + 1 次轻量 DB 查询的合计。
// 真实生产中业务逻辑通常 5~50ms，sleep 10ms 在中位数附近。
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"time"
)

type Req struct {
	UserID int64 `json:"user_id"`
}

type Resp struct {
	UserID    int64    `json:"user_id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Age       int32    `json:"age"`
	Tags      []string `json:"tags"`
	Score     float64  `json:"score"`
	CreatedAt int64    `json:"created_at"`
}

const businessLatency = 10 * time.Millisecond

func handleQuery(w http.ResponseWriter, r *http.Request) {
	// 反序列化请求
	var req Req
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()
	_ = json.Unmarshal(body, &req)

	// 模拟业务逻辑：DB 查询 / 调下游服务
	time.Sleep(businessLatency)

	resp := Resp{
		UserID:    req.UserID,
		Name:      "吴嘉晨",
		Email:     "wujiachen@example.com",
		Age:       30,
		Tags:      []string{"go", "rpc", "backend", "architecture", "ai"},
		Score:     98.7,
		CreatedAt: 1717132800,
	}
	data, _ := json.Marshal(resp)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

type breakdown struct {
	ConnectNs  int64
	MarshalNs  int64
	NetworkNs  int64
	UnmarshNs  int64
	TotalNs    int64
}

func callOnce(addr string) breakdown {
	var bd breakdown
	totalStart := time.Now()

	// 1. 序列化请求
	t0 := time.Now()
	reqBody, _ := json.Marshal(&Req{UserID: 12345})
	bd.MarshalNs = time.Since(t0).Nanoseconds()

	// 2. 建立 HTTP 请求并打 trace 钩子
	req, _ := http.NewRequest("POST", "http://"+addr+"/q", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	var connectStart, connectEnd, gotFirstByte, doneRead time.Time
	trace := &httptrace.ClientTrace{
		ConnectStart: func(_, _ string) { connectStart = time.Now() },
		ConnectDone:  func(_, _ string, _ error) { connectEnd = time.Now() },
		GotFirstResponseByte: func() { gotFirstByte = time.Now() },
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))

	// 3. 发送请求并等待响应
	tNet0 := time.Now()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return bd
	}
	respBody, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	doneRead = time.Now()

	// 4. 反序列化响应
	t0 = time.Now()
	var r Resp
	_ = json.Unmarshal(respBody, &r)
	bd.UnmarshNs = time.Since(t0).Nanoseconds()

	// 计算各段
	if !connectStart.IsZero() && !connectEnd.IsZero() {
		bd.ConnectNs = connectEnd.Sub(connectStart).Nanoseconds()
	}
	// 网络耗时 = 从发请求到读完响应的总时间 - 业务耗时
	// 但我们这里用 trace：发出 ↔ GotFirstResponseByte 这段是"网络 + 服务端处理"
	// 简化：网络耗时 ≈ doneRead - tNet0 - businessLatency
	netDur := doneRead.Sub(tNet0) - businessLatency
	if netDur < 0 {
		netDur = 0
	}
	bd.NetworkNs = netDur.Nanoseconds()
	_ = gotFirstByte

	bd.TotalNs = time.Since(totalStart).Nanoseconds()
	return bd
}

func main() {
	addr := "127.0.0.1:18811"

	// 启动服务
	srv := &http.Server{Addr: addr, Handler: http.HandlerFunc(handleQuery)}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Println("listen:", err)
		os.Exit(1)
	}
	go srv.Serve(ln)
	time.Sleep(100 * time.Millisecond)

	// warm up
	for i := 0; i < 5; i++ {
		callOnce(addr)
	}

	// 测量 N 次（Keep-Alive 连接复用，所以连接耗时主要在第一次）
	const N = 1000
	bds := make([]breakdown, 0, N)
	for i := 0; i < N; i++ {
		bds = append(bds, callOnce(addr))
	}

	// 汇总：取中位数
	collect := func(get func(breakdown) int64) (p50, p99 int64) {
		v := make([]int64, len(bds))
		for i, b := range bds {
			v[i] = get(b)
		}
		sort.Slice(v, func(i, j int) bool { return v[i] < v[j] })
		return v[len(v)/2], v[(len(v)*99)/100]
	}

	mP50, _ := collect(func(b breakdown) int64 { return b.MarshalNs })
	uP50, _ := collect(func(b breakdown) int64 { return b.UnmarshNs })
	nP50, nP99 := collect(func(b breakdown) int64 { return b.NetworkNs })
	tP50, tP99 := collect(func(b breakdown) int64 { return b.TotalNs })

	fmt.Println("# 实验 E2：端到端调用延迟分解")
	fmt.Println()
	fmt.Println("> 业务逻辑模拟耗时: 10ms（代表一次中等量级的 DB/下游调用）")
	fmt.Println("> 测量样本: 1000 次调用，HTTP/1.1 + JSON over Keep-Alive")
	fmt.Println()
	fmt.Println("## 各环节耗时（中位数 P50）")
	fmt.Println()
	fmt.Println("| 环节 | 中位耗时 | 占总耗时 |")
	fmt.Println("|------|--------:|--------:|")
	fmt.Printf("| 请求序列化 (JSON Marshal)   | %v | %.2f%% |\n",
		time.Duration(mP50), float64(mP50)*100/float64(tP50))
	fmt.Printf("| 网络栈 + 协议处理            | %v | %.2f%% |\n",
		time.Duration(nP50), float64(nP50)*100/float64(tP50))
	fmt.Printf("| 业务逻辑 (DB/下游调用模拟)   | %v | %.2f%% |\n",
		businessLatency, float64(businessLatency.Nanoseconds())*100/float64(tP50))
	fmt.Printf("| 响应反序列化 (JSON Unmarshal) | %v | %.2f%% |\n",
		time.Duration(uP50), float64(uP50)*100/float64(tP50))
	fmt.Printf("| **端到端 P50 总耗时** | **%v** | 100% |\n",
		time.Duration(tP50))
	fmt.Println()
	fmt.Printf("- 端到端 P99: %v（网络 P99: %v）\n",
		time.Duration(tP99), time.Duration(nP99))

	protocolNs := mP50 + nP50 + uP50
	protocolPct := float64(protocolNs) * 100 / float64(tP50)
	fmt.Println()
	fmt.Println("## 关键推论")
	fmt.Printf("- 「协议+序列化」总占比: **%.2f%%**\n", protocolPct)
	fmt.Printf("- 「业务逻辑」占比:        %.2f%%\n",
		float64(businessLatency.Nanoseconds())*100/float64(tP50))
	fmt.Println()
	fmt.Println("## 假设业务逻辑更重的场景（敏感性分析）")
	for _, bizMs := range []int{20, 50, 100} {
		biz := int64(bizMs) * int64(time.Millisecond)
		total := protocolNs + biz
		pct := float64(protocolNs) * 100 / float64(total)
		fmt.Printf("- 业务 %dms → 协议+序列化占比 %.2f%%\n", bizMs, pct)
	}
}

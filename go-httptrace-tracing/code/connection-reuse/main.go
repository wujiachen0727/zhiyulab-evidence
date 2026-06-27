package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

type traceResult struct {
	reused   bool
	idleTime time.Duration
	total    time.Duration
}

func measure(url string) traceResult {
	var result traceResult
	begin := time.Now()

	var dnsBegin, dnsEnd time.Time
	var dialBegin, dialEnd time.Time
	var tlsBegin, tlsEnd time.Time
	var connReady time.Time

	trace := &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) { dnsBegin = time.Now() },
		DNSDone:  func(_ httptrace.DNSDoneInfo) { dnsEnd = time.Now() },
		ConnectStart: func(_, _ string) { dialBegin = time.Now() },
		ConnectDone:  func(_, _ string, _ error) { dialEnd = time.Now() },
		TLSHandshakeStart: func() { tlsBegin = time.Now() },
		TLSHandshakeDone:  func(_ tls.ConnectionState, _ error) { tlsEnd = time.Now() },
		GotConn: func(info httptrace.GotConnInfo) {
			_ = connReady
			result.reused = info.Reused
			result.idleTime = info.IdleTime
		},
	}

	ctx := httptrace.WithClientTrace(context.Background(), trace)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "httptrace-cli/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return result
	}
	_, _ = io.Copy(io.Discard, res.Body)
	res.Body.Close()
	result.total = time.Since(begin)

	if !dnsBegin.IsZero() {
		fmt.Printf("  DNS 解析: %v\n", dnsEnd.Sub(dnsBegin))
	}
	if !dialBegin.IsZero() {
		fmt.Printf("  TCP 连接: %v\n", dialEnd.Sub(dialBegin))
	}
	if !tlsBegin.IsZero() {
		fmt.Printf("  TLS 握手: %v\n", tlsEnd.Sub(tlsBegin))
	}

	return result
}

func main() {
	url := "https://httpbin.org/get"

	fmt.Println("=== 第一次请求（应新建连接）===")
	r1 := measure(url)
	fmt.Printf("  连接复用: %v\n", r1.reused)
	fmt.Printf("  总耗时: %v\n\n", r1.total)

	time.Sleep(200 * time.Millisecond)

	fmt.Println("=== 第二次请求（应复用连接）===")
	r2 := measure(url)
	fmt.Printf("  连接复用: %v\n", r2.reused)
	if r2.reused {
		fmt.Printf("  空闲时间: %v\n", r2.idleTime)
	}
	fmt.Printf("  总耗时: %v\n\n", r2.total)

	fmt.Println("=== 第三次请求（应复用连接）===")
	r3 := measure(url)
	fmt.Printf("  连接复用: %v\n", r3.reused)
	if r3.reused {
		fmt.Printf("  空闲时间: %v\n", r3.idleTime)
	}
	fmt.Printf("  总耗时: %v\n\n", r3.total)
}

package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"time"
)

// phaseTimes 记录一次 HTTP 请求各阶段的时间点
type phaseTimes struct {
	begin     time.Time
	dnsBegin  time.Time
	dnsEnd    time.Time
	dialBegin time.Time
	dialEnd   time.Time
	tlsBegin  time.Time
	tlsEnd    time.Time
	connReady time.Time
	firstByte time.Time
	finished  time.Time
	reused    bool   // 连接是否复用
	idleTime  time.Duration // 连接空闲时间
}

func buildClientTrace(pt *phaseTimes) *httptrace.ClientTrace {
	return &httptrace.ClientTrace{
		DNSStart: func(_ httptrace.DNSStartInfo) {
			pt.dnsBegin = time.Now()
		},
		DNSDone: func(_ httptrace.DNSDoneInfo) {
			pt.dnsEnd = time.Now()
		},
		ConnectStart: func(_, _ string) {
			pt.dialBegin = time.Now()
		},
		ConnectDone: func(_, _ string, _ error) {
			pt.dialEnd = time.Now()
		},
		TLSHandshakeStart: func() {
			pt.tlsBegin = time.Now()
		},
		TLSHandshakeDone: func(_ tls.ConnectionState, _ error) {
			pt.tlsEnd = time.Now()
		},
		GotConn: func(info httptrace.GotConnInfo) {
			pt.connReady = time.Now()
			pt.reused = info.Reused
			pt.idleTime = info.IdleTime
		},
		GotFirstResponseByte: func() {
			pt.firstByte = time.Now()
		},
	}
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "用法: %s <url>\n", os.Args[0])
		os.Exit(2)
	}

	url := os.Args[1]
	pt := &phaseTimes{begin: time.Now()}

	ctx := httptrace.WithClientTrace(context.Background(), buildClientTrace(pt))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建请求失败: %v\n", err)
		os.Exit(1)
	}

	// 设置 User-Agent 避免被某些服务拒绝
	req.Header.Set("User-Agent", "httptrace-cli/1.0")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			IdleConnTimeout:     90 * time.Second,
			DisableCompression:  false,
			DisableKeepAlives:   false,
		},
	}

	res, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "请求失败: %v\n", err)
		os.Exit(1)
	}

	// 读取并丢弃响应体（确保连接可以复用）
	_, _ = io.Copy(io.Discard, res.Body)
	_ = res.Body.Close()
	pt.finished = time.Now()

	fmt.Printf("=== %s ===\n", url)
	fmt.Printf("状态码: %d\n\n", res.StatusCode)

	fmt.Println("阶段耗时:")
	if !pt.dnsBegin.IsZero() && !pt.dnsEnd.IsZero() {
		fmt.Printf("  DNS 解析:        %v\n", pt.dnsEnd.Sub(pt.dnsBegin))
	} else {
		fmt.Println("  DNS 解析:        未触发（IP 直连或缓存命中）")
	}

	if !pt.dialBegin.IsZero() && !pt.dialEnd.IsZero() {
		fmt.Printf("  TCP 连接:        %v\n", pt.dialEnd.Sub(pt.dialBegin))
	} else {
		fmt.Println("  TCP 连接:        未触发（连接复用）")
	}

	if !pt.tlsBegin.IsZero() && !pt.tlsEnd.IsZero() {
		fmt.Printf("  TLS 握手:        %v\n", pt.tlsEnd.Sub(pt.tlsBegin))
	} else {
		fmt.Println("  TLS 握手:        HTTP 请求，无 TLS")
	}

	if !pt.connReady.IsZero() && !pt.firstByte.IsZero() {
		fmt.Printf("  等待首字节(TTFB): %v\n", pt.firstByte.Sub(pt.connReady))
	}

	if !pt.firstByte.IsZero() {
		fmt.Printf("  Body 传输:       %v\n", pt.finished.Sub(pt.firstByte))
	}

	fmt.Printf("  总耗时:          %v\n\n", pt.finished.Sub(pt.begin))

	fmt.Printf("连接信息:\n")
	fmt.Printf("  连接复用:        %v\n", pt.reused)
	if pt.reused && pt.idleTime > 0 {
		fmt.Printf("  空闲时间:        %v\n", pt.idleTime)
	}
}

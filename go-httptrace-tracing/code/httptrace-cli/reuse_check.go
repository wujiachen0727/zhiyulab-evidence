package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"time"
)

func main() {
	url := "https://httpbin.org/get"

	// 第一次请求——新连接
	pt1 := measure(url)
	fmt.Printf("=== 第一次请求 ===\n")
	fmt.Printf("连接复用: %v\n\n", pt1.reused)

	time.Sleep(100 * time.Millisecond)

	// 第二次请求——应复用连接
	pt2 := measure(url)
	fmt.Printf("=== 第二次请求（应复用连接）===\n")
	fmt.Printf("连接复用: %v\n", pt2.reused)
}

func measure(url string) phaseTimes {
	pt := phaseTimes{begin: time.Now()}
	ctx := httptrace.WithClientTrace(context.Background(), buildClientTrace(&pt))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("User-Agent", "httptrace-cli/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return pt
	}
	_, _ = io.Copy(io.Discard, res.Body)
	res.Body.Close()
	pt.finished = time.Now()
	return pt
}

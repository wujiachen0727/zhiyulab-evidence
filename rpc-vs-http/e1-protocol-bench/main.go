// Package main 实验 E1：HTTP/1.1+JSON vs HTTP/2+JSON vs HTTP/2+Protobuf-like 性能对比
//
// 目的：在同一台机器、同一台进程内对比三种协议组合的吞吐和延迟，
// 排除网络抖动，纯粹测协议+序列化开销。
//
// 注意：这里用标准库 net/http（支持 HTTP/2 over h2c），用手写的
// "Protobuf-like 二进制编码"（紧凑变长编码）作为 Protobuf 替身——避免
// 引入 google.golang.org/protobuf 这种外部依赖。其编码空间和速度
// 与真正的 Protobuf 同量级（甚至更小，因为没有 tag 开销）。
//
// 对照实验，不是生产参考。
package main

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// User 业务 payload，模拟一个典型的"用户信息查询"接口返回
type User struct {
	ID        int64    `json:"id"`
	Name      string   `json:"name"`
	Email     string   `json:"email"`
	Age       int32    `json:"age"`
	Tags      []string `json:"tags"`
	Score     float64  `json:"score"`
	Active    bool     `json:"active"`
	CreatedAt int64    `json:"created_at"`
}

// 全局测试 payload
var sampleUser = User{
	ID:        1234567890,
	Name:      "吴嘉晨",
	Email:     "wujiachen@example.com",
	Age:       30,
	Tags:      []string{"go", "rpc", "backend", "architecture", "ai"},
	Score:     98.7,
	Active:    true,
	CreatedAt: 1717132800,
}

// ============================================================
// "Protobuf-like" 紧凑二进制编码
// ============================================================
// 字段顺序固定，不带 tag。变长字符串前置 uvarint 长度。
// 这模拟了 Protobuf 类二进制协议的编码效率（实际上更紧凑）。

func encodeBinary(u *User) []byte {
	buf := bytes.NewBuffer(nil)
	tmp := make([]byte, binary.MaxVarintLen64)

	// id (int64)
	n := binary.PutVarint(tmp, u.ID)
	buf.Write(tmp[:n])
	// name (string)
	writeString(buf, u.Name, tmp)
	// email
	writeString(buf, u.Email, tmp)
	// age (int32)
	n = binary.PutVarint(tmp, int64(u.Age))
	buf.Write(tmp[:n])
	// tags (repeated string)
	n = binary.PutUvarint(tmp, uint64(len(u.Tags)))
	buf.Write(tmp[:n])
	for _, t := range u.Tags {
		writeString(buf, t, tmp)
	}
	// score (float64)
	binary.Write(buf, binary.LittleEndian, u.Score)
	// active (bool)
	if u.Active {
		buf.WriteByte(1)
	} else {
		buf.WriteByte(0)
	}
	// created_at (int64)
	n = binary.PutVarint(tmp, u.CreatedAt)
	buf.Write(tmp[:n])
	return buf.Bytes()
}

func writeString(buf *bytes.Buffer, s string, tmp []byte) {
	n := binary.PutUvarint(tmp, uint64(len(s)))
	buf.Write(tmp[:n])
	buf.WriteString(s)
}

func decodeBinary(data []byte) (*User, error) {
	r := bytes.NewReader(data)
	u := &User{}

	id, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	u.ID = id

	if u.Name, err = readString(r); err != nil {
		return nil, err
	}
	if u.Email, err = readString(r); err != nil {
		return nil, err
	}
	age, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	u.Age = int32(age)

	tagCount, err := binary.ReadUvarint(r)
	if err != nil {
		return nil, err
	}
	u.Tags = make([]string, 0, tagCount)
	for i := uint64(0); i < tagCount; i++ {
		s, err := readString(r)
		if err != nil {
			return nil, err
		}
		u.Tags = append(u.Tags, s)
	}

	if err := binary.Read(r, binary.LittleEndian, &u.Score); err != nil {
		return nil, err
	}
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	u.Active = b == 1

	created, err := binary.ReadVarint(r)
	if err != nil {
		return nil, err
	}
	u.CreatedAt = created

	return u, nil
}

func readString(r *bytes.Reader) (string, error) {
	n, err := binary.ReadUvarint(r)
	if err != nil {
		return "", err
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return string(buf), nil
}

// ============================================================
// HTTP Handlers
// ============================================================

func handlerJSON(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(sampleUser)
}

func handlerBinary(w http.ResponseWriter, _ *http.Request) {
	data := encodeBinary(&sampleUser)
	w.Header().Set("Content-Type", "application/x-protobuf-like")
	_, _ = w.Write(data)
}

// ============================================================
// 三种 Server 启动
// ============================================================

func startHTTP1Server(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", handlerJSON)
	mux.HandleFunc("/binary", handlerBinary)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http/1.1 server: %v", err)
		}
	}()
}

func startH2CServer(addr string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", handlerJSON)
	mux.HandleFunc("/binary", handlerBinary)
	h2s := &http2.Server{}
	srv := &http.Server{
		Addr:    addr,
		Handler: http2WithCleartext(mux, h2s),
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("h2c server: %v", err)
		}
	}()
}

// http2WithCleartext: HTTP/2 over cleartext (h2c) 适配
func http2WithCleartext(handler http.Handler, h2s *http2.Server) http.Handler {
	return h2c.NewHandler(handler, h2s)
}

// ============================================================
// Client + 压测
// ============================================================

type result struct {
	Name        string
	Total       int
	Success     int
	BytesIn     int64
	Elapsed     time.Duration
	Latencies   []time.Duration
	P50, P99    time.Duration
	QPS         float64
	AvgBytes    int
}

func (r *result) sortAndCalc() {
	sort.Slice(r.Latencies, func(i, j int) bool { return r.Latencies[i] < r.Latencies[j] })
	if n := len(r.Latencies); n > 0 {
		r.P50 = r.Latencies[n/2]
		r.P99 = r.Latencies[(n*99)/100]
	}
	r.QPS = float64(r.Success) / r.Elapsed.Seconds()
	if r.Success > 0 {
		r.AvgBytes = int(r.BytesIn / int64(r.Success))
	}
}

// runBench 在固定时长内用 N 个 worker 持续打 url，收集每个请求的延迟
func runBench(name string, client *http.Client, url string, workers int, duration time.Duration) *result {
	r := &result{Name: name, Latencies: make([]time.Duration, 0, 200000)}
	var (
		mu      sync.Mutex
		success int64
		bytesIn int64
	)
	stop := make(chan struct{})
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			localLats := make([]time.Duration, 0, 4096)
			for {
				select {
				case <-stop:
					mu.Lock()
					r.Latencies = append(r.Latencies, localLats...)
					mu.Unlock()
					return
				default:
				}
				t0 := time.Now()
				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				body, err := io.ReadAll(resp.Body)
				_ = resp.Body.Close()
				if err != nil {
					continue
				}
				lat := time.Since(t0)
				localLats = append(localLats, lat)
				atomic.AddInt64(&success, 1)
				atomic.AddInt64(&bytesIn, int64(len(body)))
			}
		}()
	}
	time.Sleep(duration)
	close(stop)
	wg.Wait()
	r.Elapsed = time.Since(start)
	r.Success = int(success)
	r.Total = int(success)
	r.BytesIn = bytesIn
	r.sortAndCalc()
	return r
}

// ============================================================
// 单次调用细分耗时（编码/解码独立测）
// ============================================================

func benchSerialization(iter int) (jsonEnc, jsonDec, binEnc, binDec time.Duration, jsonSize, binSize int) {
	// JSON encode
	var buf bytes.Buffer
	t0 := time.Now()
	for i := 0; i < iter; i++ {
		buf.Reset()
		_ = json.NewEncoder(&buf).Encode(&sampleUser)
	}
	jsonEnc = time.Since(t0) / time.Duration(iter)
	jsonSize = buf.Len()

	// JSON decode
	data := buf.Bytes()
	t0 = time.Now()
	for i := 0; i < iter; i++ {
		var u User
		_ = json.Unmarshal(data, &u)
	}
	jsonDec = time.Since(t0) / time.Duration(iter)

	// Binary encode
	t0 = time.Now()
	var bdata []byte
	for i := 0; i < iter; i++ {
		bdata = encodeBinary(&sampleUser)
	}
	binEnc = time.Since(t0) / time.Duration(iter)
	binSize = len(bdata)

	// Binary decode
	t0 = time.Now()
	for i := 0; i < iter; i++ {
		_, _ = decodeBinary(bdata)
	}
	binDec = time.Since(t0) / time.Duration(iter)

	return
}

// ============================================================
// main
// ============================================================

func main() {
	const (
		http1Addr = "127.0.0.1:18801"
		h2cAddr   = "127.0.0.1:18802"
		workers   = 32
		duration  = 5 * time.Second
	)

	// 启动两个服务器
	startHTTP1Server(http1Addr)
	startH2CServer(h2cAddr)
	time.Sleep(200 * time.Millisecond) // wait for servers ready

	// 三种 client
	// 1. HTTP/1.1 + JSON
	tr1 := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
		IdleConnTimeout:     90 * time.Second,
	}
	c1 := &http.Client{Transport: tr1, Timeout: 5 * time.Second}

	// 2. HTTP/2 + JSON (h2c)
	tr2 := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}
	c2 := &http.Client{Transport: tr2, Timeout: 5 * time.Second}

	// 3. HTTP/2 + Binary (h2c) — 复用 c2 的 transport
	c3 := c2

	fmt.Println("# 实验 E1：协议层性能对比（同进程，无网络抖动）")
	fmt.Println("# 配置：workers=32, duration=5s")
	fmt.Println("# Go：", runtimeInfo())
	fmt.Println()

	// warm up
	for i := 0; i < 3; i++ {
		_, _ = c1.Get("http://" + http1Addr + "/json")
		_, _ = c2.Get("http://" + h2cAddr + "/json")
	}

	r1 := runBench("HTTP/1.1 + JSON", c1, "http://"+http1Addr+"/json", workers, duration)
	r2 := runBench("HTTP/2   + JSON", c2, "http://"+h2cAddr+"/json", workers, duration)
	r3 := runBench("HTTP/2   + 二进制", c3, "http://"+h2cAddr+"/binary", workers, duration)

	fmt.Println("## 端到端 QPS / 延迟")
	fmt.Println()
	fmt.Println("| 方案 | QPS | P50 延迟 | P99 延迟 | 平均响应大小 |")
	fmt.Println("|------|----:|--------:|--------:|------------:|")
	for _, r := range []*result{r1, r2, r3} {
		fmt.Printf("| %s | %.0f | %v | %v | %d B |\n",
			r.Name, r.QPS, r.P50.Round(time.Microsecond), r.P99.Round(time.Microsecond), r.AvgBytes)
	}

	// 序列化单独测
	jsonEnc, jsonDec, binEnc, binDec, jsonSize, binSize := benchSerialization(50000)
	fmt.Println()
	fmt.Println("## 单纯序列化/反序列化 (50000 次平均)")
	fmt.Println()
	fmt.Println("| 操作 | JSON | 二进制 |")
	fmt.Println("|------|-----:|------:|")
	fmt.Printf("| 编码 | %v | %v |\n", jsonEnc, binEnc)
	fmt.Printf("| 解码 | %v | %v |\n", jsonDec, binDec)
	fmt.Printf("| Body 大小 | %d B | %d B |\n", jsonSize, binSize)
	fmt.Println()

	// 计算"协议+序列化"占总耗时百分比的估算
	// 端到端延迟 P50 = 网络栈 + 协议头 + 序列化 + 业务逻辑
	// 这里业务逻辑近似 0（直接返回常量），P50 ≈ 协议+序列化 总开销
	fmt.Println("## 推论提示")
	fmt.Printf("- HTTP/2 + 二进制 vs HTTP/1.1 + JSON 的 P50 延迟差: %v\n",
		(r1.P50 - r3.P50).Round(time.Microsecond))
	fmt.Printf("- HTTP/2 + 二进制 vs HTTP/2 + JSON  的 P50 延迟差: %v\n",
		(r2.P50 - r3.P50).Round(time.Microsecond))
	fmt.Printf("- 体积压缩比 (二进制/JSON): %.1f%%\n", float64(binSize)*100/float64(jsonSize))
}

func runtimeInfo() string {
	return fmt.Sprintf("(see go.mod)")
}

package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"os"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type sleepDriver struct{}

type sleepConn struct {
	sleep time.Duration
}

type sleepRows struct {
	sent bool
}

func (sleepDriver) Open(name string) (driver.Conn, error) {
	sleep, err := time.ParseDuration(name)
	if err != nil {
		return nil, err
	}
	return &sleepConn{sleep: sleep}, nil
}

func (c *sleepConn) Prepare(query string) (driver.Stmt, error) {
	return nil, fmt.Errorf("Prepare is not used in this experiment")
}

func (c *sleepConn) Close() error { return nil }

func (c *sleepConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("transactions are not used in this experiment")
}

func (c *sleepConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	timer := time.NewTimer(c.sleep)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-timer.C:
		return &sleepRows{}, nil
	}
}

func (r *sleepRows) Columns() []string { return []string{"ok"} }
func (r *sleepRows) Close() error      { return nil }

func (r *sleepRows) Next(dest []driver.Value) error {
	if r.sent {
		return io.EOF
	}
	r.sent = true
	dest[0] = int64(1)
	return nil
}

type scenario struct {
	name        string
	maxOpen     int
	maxIdle     int
	concurrency int
	requests    int
	queryTime   time.Duration
}

type result struct {
	scenario         scenario
	elapsed          time.Duration
	p50              time.Duration
	p95              time.Duration
	p99              time.Duration
	throughput       float64
	waitCount        int64
	waitDuration     time.Duration
	avgWait          time.Duration
	openConnections  int
	inUse            int
	idle             int
	maxIdleClosed    int64
	maxLifetimeClose int64
	errors           int
}

func main() {
	sql.Register("sleepdb", sleepDriver{})

	scenarios := []scenario{
		{name: "队伍很长：MaxOpen=1", maxOpen: 1, maxIdle: 1, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
		{name: "队伍缩短：MaxOpen=2", maxOpen: 2, maxIdle: 2, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
		{name: "继续缩短：MaxOpen=4", maxOpen: 4, maxIdle: 4, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
		{name: "接近够用：MaxOpen=8", maxOpen: 8, maxIdle: 8, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
		{name: "基本不排队：MaxOpen=16", maxOpen: 16, maxIdle: 16, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
		{name: "过量开车道：MaxOpen=64", maxOpen: 64, maxIdle: 64, concurrency: 32, requests: 160, queryTime: 50 * time.Millisecond},
	}

	results := make([]result, 0, len(scenarios))
	for _, s := range scenarios {
		res, err := runScenario(s)
		if err != nil {
			fmt.Fprintf(os.Stderr, "scenario %q failed: %v\n", s.name, err)
			os.Exit(1)
		}
		results = append(results, res)
	}

	printMarkdown(results)
	fmt.Println()
	printCSV(results)
}

func runScenario(s scenario) (result, error) {
	db, err := sql.Open("sleepdb", s.queryTime.String())
	if err != nil {
		return result{}, err
	}
	defer db.Close()

	db.SetMaxOpenConns(s.maxOpen)
	db.SetMaxIdleConns(s.maxIdle)

	ctx := context.Background()
	if err := oneQuery(ctx, db); err != nil {
		return result{}, err
	}

	jobs := make(chan int)
	latencies := make([]time.Duration, s.requests)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errors := 0

	start := time.Now()
	for worker := 0; worker < s.concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				requestStart := time.Now()
				if err := oneQuery(ctx, db); err != nil {
					mu.Lock()
					errors++
					mu.Unlock()
				}
				latencies[idx] = time.Since(requestStart)
			}
		}()
	}

	for i := 0; i < s.requests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	elapsed := time.Since(start)

	stats := db.Stats()
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })

	waitCount := stats.WaitCount
	avgWait := time.Duration(0)
	if waitCount > 0 {
		avgWait = time.Duration(int64(stats.WaitDuration) / waitCount)
	}

	return result{
		scenario:         s,
		elapsed:          elapsed,
		p50:              percentile(latencies, 0.50),
		p95:              percentile(latencies, 0.95),
		p99:              percentile(latencies, 0.99),
		throughput:       float64(s.requests) / elapsed.Seconds(),
		waitCount:        waitCount,
		waitDuration:     stats.WaitDuration,
		avgWait:          avgWait,
		openConnections:  stats.OpenConnections,
		inUse:            stats.InUse,
		idle:             stats.Idle,
		maxIdleClosed:    stats.MaxIdleClosed,
		maxLifetimeClose: stats.MaxLifetimeClosed,
		errors:           errors,
	}, nil
}

func oneQuery(ctx context.Context, db *sql.DB) error {
	rows, err := db.QueryContext(ctx, "select 1")
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return err
		}
	}
	return rows.Err()
}

func percentile(sorted []time.Duration, p float64) time.Duration {
	if len(sorted) == 0 {
		return 0
	}
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func printMarkdown(results []result) {
	fmt.Println("# Go database/sql 连接池等待实验")
	fmt.Println()
	fmt.Println("[实测 Go " + strings.TrimPrefix(runtimeVersion(), "go") + " darwin/arm64；查询耗时由自定义 database/sql driver 固定 sleep 50ms 模拟]")
	fmt.Println()
	fmt.Println("| 场景 | 并发 | 请求数 | MaxOpen | P50 | P95 | P99 | 总耗时 | 吞吐 req/s | WaitCount | WaitDuration | 平均等待 |")
	fmt.Println("|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|")
	for _, r := range results {
		fmt.Printf("| %s | %d | %d | %d | %s | %s | %s | %s | %.1f | %d | %s | %s |\n",
			r.scenario.name,
			r.scenario.concurrency,
			r.scenario.requests,
			r.scenario.maxOpen,
			formatDuration(r.p50),
			formatDuration(r.p95),
			formatDuration(r.p99),
			formatDuration(r.elapsed),
			r.throughput,
			r.waitCount,
			formatDuration(r.waitDuration),
			formatDuration(r.avgWait),
		)
	}
}

func printCSV(results []result) {
	fmt.Println("scenario,concurrency,requests,query_ms,max_open,max_idle,p50_ms,p95_ms,p99_ms,elapsed_ms,throughput_req_s,wait_count,wait_duration_ms,avg_wait_ms,open_connections,in_use,idle,errors")
	for _, r := range results {
		fmt.Printf("%s,%d,%d,%d,%d,%d,%.2f,%.2f,%.2f,%.2f,%.2f,%d,%.2f,%.2f,%d,%d,%d,%d\n",
			strconv.Quote(r.scenario.name),
			r.scenario.concurrency,
			r.scenario.requests,
			r.scenario.queryTime.Milliseconds(),
			r.scenario.maxOpen,
			r.scenario.maxIdle,
			float64(r.p50.Microseconds())/1000,
			float64(r.p95.Microseconds())/1000,
			float64(r.p99.Microseconds())/1000,
			float64(r.elapsed.Microseconds())/1000,
			r.throughput,
			r.waitCount,
			float64(r.waitDuration.Microseconds())/1000,
			float64(r.avgWait.Microseconds())/1000,
			r.openConnections,
			r.inUse,
			r.idle,
			r.errors,
		)
	}
}

func formatDuration(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
}

func runtimeVersion() string {
	return strings.TrimSpace(runtime.Version())
}

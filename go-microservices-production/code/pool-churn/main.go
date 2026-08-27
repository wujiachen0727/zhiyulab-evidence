// 对比 MaxIdle << MaxOpen 时的连接 churn（MaxIdleClosed）与 P99 延迟
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

func (c *sleepConn) Prepare(string) (driver.Stmt, error) {
	return nil, fmt.Errorf("not used")
}

func (c *sleepConn) Close() error { return nil }

func (c *sleepConn) Begin() (driver.Tx, error) {
	return nil, fmt.Errorf("not used")
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

type cfg struct {
	name    string
	maxOpen int
	maxIdle int
}

type benchResult struct {
	cfg           cfg
	p99           time.Duration
	maxIdleClosed int64
	waitCount     int64
}

func main() {
	sql.Register("sleepdb", sleepDriver{})

	const (
		concurrency = 40
		requests    = 200
		queryTime   = 20 * time.Millisecond
		maxOpen     = 20
	)

	configs := []cfg{
		{name: "churn: MaxIdle=2, MaxOpen=20", maxOpen: maxOpen, maxIdle: 2},
		{name: "warm: MaxIdle=20, MaxOpen=20", maxOpen: maxOpen, maxIdle: 20},
	}

	var results []benchResult
	for _, c := range configs {
		res, err := run(c, concurrency, requests, queryTime)
		if err != nil {
			fmt.Fprintf(os.Stderr, "run failed: %v\n", err)
			os.Exit(1)
		}
		results = append(results, res)
	}

	fmt.Println("# 连接池 churn 对比实验")
	fmt.Printf("\n环境：%s；并发 %d，请求 %d，单次查询 sleep %s\n\n", runtime.Version(), concurrency, requests, queryTime)
	fmt.Println("| 配置 | P99 | MaxIdleClosed | WaitCount |")
	fmt.Println("|---|---:|---:|---:|")
	for _, r := range results {
		fmt.Printf("| %s | %s | %d | %d |\n", r.cfg.name, formatDur(r.p99), r.maxIdleClosed, r.waitCount)
	}
}

func run(c cfg, concurrency, requests int, queryTime time.Duration) (benchResult, error) {
	db, err := sql.Open("sleepdb", queryTime.String())
	if err != nil {
		return benchResult{}, err
	}
	defer db.Close()

	db.SetMaxOpenConns(c.maxOpen)
	db.SetMaxIdleConns(c.maxIdle)

	ctx := context.Background()
	if err := queryOnce(ctx, db); err != nil {
		return benchResult{}, err
	}

	jobs := make(chan int)
	latencies := make([]time.Duration, requests)
	var wg sync.WaitGroup

	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				t0 := time.Now()
				_ = queryOnce(ctx, db)
				latencies[idx] = time.Since(t0)
			}
		}()
	}

	for i := 0; i < requests; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()

	stats := db.Stats()
	sort.Slice(latencies, func(i, j int) bool { return latencies[i] < latencies[j] })
	p99Idx := int(float64(len(latencies)-1) * 0.99)

	return benchResult{
		cfg:           c,
		p99:           latencies[p99Idx],
		maxIdleClosed: stats.MaxIdleClosed,
		waitCount:     stats.WaitCount,
	}, nil
}

func queryOnce(ctx context.Context, db *sql.DB) error {
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

func formatDur(d time.Duration) string {
	if d >= time.Second {
		return fmt.Sprintf("%.2fs", d.Seconds())
	}
	return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
}

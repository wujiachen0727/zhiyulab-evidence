// 模拟 3 层服务链共享同一 DB 池上限时的 WaitCount 增长
package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"io"
	"runtime"
	"sort"
	"sync"
	"time"
)

type sleepDriver struct{}

type sleepConn struct{ sleep time.Duration }

type sleepRows struct{ sent bool }

func (sleepDriver) Open(name string) (driver.Conn, error) {
	d, err := time.ParseDuration(name)
	if err != nil {
		return nil, err
	}
	return &sleepConn{sleep: d}, nil
}

func (c *sleepConn) Prepare(string) (driver.Stmt, error) { return nil, fmt.Errorf("n/a") }
func (c *sleepConn) Close() error                         { return nil }
func (c *sleepConn) Begin() (driver.Tx, error)            { return nil, fmt.Errorf("n/a") }

func (c *sleepConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	t := time.NewTimer(c.sleep)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-t.C:
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

func main() {
	sql.Register("sleepdb", sleepDriver{})

	const (
		maxOpen     = 5
		concurrency = 30
		requests    = 120
		hops        = 3
		hopSleep    = 15 * time.Millisecond
	)

	db, err := sql.Open("sleepdb", hopSleep.String())
	if err != nil {
		panic(err)
	}
	defer db.Close()
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxOpen)

	ctx := context.Background()
	latencies := make([]time.Duration, requests)
	jobs := make(chan int)
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for idx := range jobs {
				start := time.Now()
				for hop := 0; hop < hops; hop++ {
					if err := queryOnce(ctx, db); err != nil {
						break
					}
				}
				latencies[idx] = time.Since(start)
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
	p99 := latencies[int(float64(len(latencies)-1)*0.99)]

	fmt.Println("# 三跳链 + 小连接池实验")
	fmt.Printf("\n环境：%s；MaxOpen=%d，并发=%d，每请求 %d 次 DB 查询\n\n", runtime.Version(), maxOpen, concurrency, hops)
	fmt.Printf("| P99 延迟 | WaitCount | WaitDuration | InUse 峰值观察 |\n")
	fmt.Printf("|---|---:|---:|---:|\n")
	fmt.Printf("| %s | %d | %s | %d |\n", formatDur(p99), stats.WaitCount, formatDur(stats.WaitDuration), stats.InUse)
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

package main

import (
	"flag"
	"fmt"
	"runtime"
	"strings"
)

type TracePoint struct {
	AppendIndex int
	Len         int
	Cap         int
	Bytes       int
}

func traceAppendByte(n int) []TracePoint {
	var s []byte
	points := []TracePoint{{AppendIndex: 0, Len: len(s), Cap: cap(s), Bytes: cap(s)}}
	lastCap := cap(s)
	for i := 0; i < n; i++ {
		s = append(s, byte(i))
		if cap(s) != lastCap {
			points = append(points, TracePoint{
				AppendIndex: i + 1,
				Len:         len(s),
				Cap:         cap(s),
				Bytes:       cap(s),
			})
			lastCap = cap(s)
		}
	}
	return points
}

func traceGrowFromCap(startCap int, appends int) []TracePoint {
	s := make([]byte, startCap, startCap)
	points := []TracePoint{{AppendIndex: 0, Len: len(s), Cap: cap(s), Bytes: cap(s)}}
	lastCap := cap(s)
	for i := 0; i < appends; i++ {
		s = append(s, byte(i))
		if cap(s) != lastCap {
			points = append(points, TracePoint{
				AppendIndex: i + 1,
				Len:         len(s),
				Cap:         cap(s),
				Bytes:       cap(s),
			})
			lastCap = cap(s)
		}
	}
	return points
}

func printScanStartCaps(from, to int) {
	fmt.Println("oldcap,newcap,delta,bytes,monotonic_vs_prev")
	prevNewCap := -1
	for oldcap := from; oldcap <= to; oldcap++ {
		s := make([]byte, oldcap, oldcap)
		s = append(s, 1)
		newcap := cap(s)
		status := "-"
		if prevNewCap >= 0 {
			if newcap < prevNewCap {
				status = "down"
			} else if newcap == prevNewCap {
				status = "flat"
			} else {
				status = "up"
			}
		}
		fmt.Printf("%d,%d,%d,%d,%s\n", oldcap, newcap, newcap-oldcap, newcap, status)
		prevNewCap = newcap
	}
}

func printTrace(label string, points []TracePoint) {
	fmt.Printf("# %s\n", label)
	fmt.Println("idx,append_index,len,cap,bytes,growth")
	prev := 0
	for i, p := range points {
		growth := "-"
		if i > 0 && prev > 0 {
			growth = fmt.Sprintf("%.4f", float64(p.Cap)/float64(prev))
		}
		fmt.Printf("%d,%d,%d,%d,%d,%s\n", i, p.AppendIndex, p.Len, p.Cap, p.Bytes, growth)
		prev = p.Cap
	}
}

func printSummary(label string, points []TracePoint) {
	fmt.Printf("# SUMMARY %s\n", label)
	fmt.Printf("go_version=%s\n", runtime.Version())
	fmt.Printf("goos=%s\n", runtime.GOOS)
	fmt.Printf("goarch=%s\n", runtime.GOARCH)
	fmt.Printf("growth_events=%d\n", len(points)-1)
	if len(points) > 0 {
		fmt.Printf("final_len=%d\n", points[len(points)-1].Len)
		fmt.Printf("final_cap=%d\n", points[len(points)-1].Cap)
		fmt.Printf("final_bytes=%d\n", points[len(points)-1].Bytes)
	}
}

func main() {
	mode := flag.String("mode", "append", "append / grow-from-cap / scan-start-caps")
	n := flag.Int("n", 5000, "append 次数")
	startCap := flag.Int("start-cap", 1024, "grow-from-cap 模式的初始 cap")
	fromCap := flag.Int("from-cap", 900, "scan-start-caps 模式的起始 oldcap")
	toCap := flag.Int("to-cap", 1300, "scan-start-caps 模式的结束 oldcap")
	flag.Parse()

	fmt.Printf("# runtime=%s %s/%s mode=%s n=%d start_cap=%d\n", runtime.Version(), runtime.GOOS, runtime.GOARCH, *mode, *n, *startCap)

	switch strings.ToLower(*mode) {
	case "append":
		points := traceAppendByte(*n)
		printTrace("append-byte-growth", points)
		printSummary("append-byte-growth", points)
	case "grow-from-cap":
		points := traceGrowFromCap(*startCap, *n)
		printTrace("grow-from-cap", points)
		printSummary("grow-from-cap", points)
	case "scan-start-caps":
		fmt.Printf("# scan-start-caps from=%d to=%d\n", *fromCap, *toCap)
		printScanStartCaps(*fromCap, *toCap)
	default:
		panic("unknown mode")
	}
}

// 对照：channel 在全局变量上时，goroutineleak 可能漏检
package main

import (
	"fmt"
	"os"
	"runtime/pprof"
	"time"
)

var globalCh = make(chan struct{})

func main() {
	mode := "global-channel"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	switch mode {
	case "local-channel":
		go func() {
			ch := make(chan struct{})
			<-ch
		}()
	case "global-channel":
		go func() {
			<-globalCh
		}()
	default:
		fmt.Fprintf(os.Stderr, "usage: goroutineleak [local-channel|global-channel]\n")
		os.Exit(2)
	}

	time.Sleep(200 * time.Millisecond)

	f, err := os.Create(fmt.Sprintf("goroutineleak-%s.prof", mode))
	if err != nil {
		panic(err)
	}
	defer f.Close()

	if err := pprof.Lookup("goroutineleak").WriteTo(f, 0); err != nil {
		fmt.Fprintf(os.Stderr, "write profile: %v\n", err)
		os.Exit(1)
	}

	// 打印 profile 是否为空
	info, _ := f.Stat()
	fmt.Printf("mode=%s bytes=%d\n", mode, info.Size())
}

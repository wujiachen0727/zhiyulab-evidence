package main

import (
	"fmt"
	"runtime"
	"sync"
)

func main() {
	fmt.Printf("工具链 runtime.Version() = %s\n", runtime.Version())
	fmt.Printf("runtime.NumCPU()         = %d\n", runtime.NumCPU())
	fmt.Printf("runtime.GOMAXPROCS(0)    = %d\n\n", runtime.GOMAXPROCS(0))

	var (
		wg  sync.WaitGroup
		mu  sync.Mutex
		sum int
	)

	for i := 1; i <= 3; i++ {
		// sync.WaitGroup.Go 是 Go 1.25 新增的方法
		wg.Go(func() {
			mu.Lock()
			sum += i
			mu.Unlock()
		})
	}
	wg.Wait()

	fmt.Printf("wg.Go 汇总结果 = %d\n", sum)
}

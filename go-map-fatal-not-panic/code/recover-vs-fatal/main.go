package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: recover-vs-fatal <panic|fatal>")
		os.Exit(64)
	}

	switch os.Args[1] {
	case "panic":
		runRecoverablePanic()
	case "fatal":
		runConcurrentMapWrites()
	default:
		fmt.Fprintln(os.Stderr, "unknown mode")
		os.Exit(64)
	}
}

func runRecoverablePanic() {
	defer func() {
		if recovered := recover(); recovered != nil {
			fmt.Printf("recovered ordinary panic: %v\n", recovered)
		}
	}()

	panic("ordinary panic")
}

func runConcurrentMapWrites() {
	runtime.GOMAXPROCS(8)

	m := map[int]int{}
	start := make(chan struct{})
	var ready sync.WaitGroup

	for workerID := 0; workerID < 16; workerID++ {
		ready.Add(1)
		go func(id int) {
			defer func() {
				if recovered := recover(); recovered != nil {
					fmt.Printf("unexpectedly recovered in worker %d: %v\n", id, recovered)
				}
			}()

			ready.Done()
			<-start
			for i := 0; i < 10_000_000; i++ {
				m[(i+id)%1024] = i
			}
		}(workerID)
	}

	ready.Wait()
	close(start)

	time.Sleep(3 * time.Second)
	fmt.Println("completed_without_fatal_unexpected")
}

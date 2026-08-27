// 演示 gobreaker 未区分 gRPC client error 时误 trip
package main

import (
	"fmt"
	"os"

	"github.com/sony/gobreaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	const calls = 20

	naive := runBreaker("naive (NotFound=失败)", func(err error) bool {
		return err == nil
	}, calls)

	correct := runBreaker("correct (NotFound=成功)", func(err error) bool {
		if err == nil {
			return true
		}
		st, ok := status.FromError(err)
		if !ok {
			return false
		}
		switch st.Code() {
		case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists,
			codes.PermissionDenied, codes.Unauthenticated:
			return true
		default:
			return false
		}
	}, calls)

	fmt.Println("# gobreaker gRPC 状态码误 trip 实验")
	fmt.Println()
	fmt.Println("| 策略 | 20 次 NotFound 后状态 | Trip 次数 |")
	fmt.Println("|---|---|---:|")
	fmt.Printf("| %s | %s | %d |\n", naive.name, naive.finalState, naive.trips)
	fmt.Printf("| %s | %s | %d |\n", correct.name, correct.finalState, correct.trips)
}

type outcome struct {
	name       string
	finalState string
	trips      int
}

func runBreaker(name string, isSuccessful func(error) bool, n int) outcome {
	trips := 0
	cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 1,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(_ string, from, to gobreaker.State) {
			if to == gobreaker.StateOpen {
				trips++
			}
		},
		IsSuccessful: isSuccessful,
	})

	notFound := status.Error(codes.NotFound, "user not found")
	for i := 0; i < n; i++ {
		_, err := cb.Execute(func() (any, error) {
			return nil, notFound
		})
		if err != nil && err != gobreaker.ErrOpenState && err != gobreaker.ErrTooManyRequests {
			// 预期 NotFound 返回给调用方
		}
	}

	return outcome{name: name, finalState: cb.State().String(), trips: trips}
}

func init() {
	if os.Getenv("GOBREAKER_DEMO") == "" {
		return
	}
}

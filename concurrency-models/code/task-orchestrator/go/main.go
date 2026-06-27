package main

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"
)

type Task struct {
	Name  string
	Delay time.Duration
	Fail  bool
}

type Result struct {
	Name  string
	Value string
	Err   error
}

type Summary struct {
	Scenario  string
	Completed []string
	Canceled  []string
	Failed    []string
	Err       error
}

func runTask(ctx context.Context, task Task) Result {
	select {
	case <-time.After(task.Delay):
		if task.Fail {
			return Result{Name: task.Name, Err: fmt.Errorf("%s returned error", task.Name)}
		}
		return Result{Name: task.Name, Value: task.Name + "-ok"}
	case <-ctx.Done():
		return Result{Name: task.Name, Err: fmt.Errorf("%s canceled: %w", task.Name, ctx.Err())}
	}
}

func orchestrate(parent context.Context, scenario string, tasks []Task, timeout time.Duration) Summary {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()

	resultCh := make(chan Result, len(tasks))
	for _, task := range tasks {
		task := task
		go func() {
			resultCh <- runTask(ctx, task)
		}()
	}

	summary := Summary{Scenario: scenario}
	for range tasks {
		result := <-resultCh
		if result.Err != nil {
			if strings.Contains(result.Err.Error(), "canceled") {
				summary.Canceled = append(summary.Canceled, result.Name)
			} else {
				summary.Failed = append(summary.Failed, result.Name)
			}
			if summary.Err == nil {
				summary.Err = result.Err
				cancel()
			}
			continue
		}
		summary.Completed = append(summary.Completed, result.Name)
	}

	sort.Strings(summary.Completed)
	sort.Strings(summary.Canceled)
	sort.Strings(summary.Failed)
	return summary
}

func printSummary(summary Summary) {
	fmt.Printf("scenario=%s\n", summary.Scenario)
	fmt.Printf("state_owner=aggregator goroutine owns the result slice; workers only send immutable Result values\n")
	fmt.Printf("waiting_owner=context.WithTimeout + select make waiting and cancellation explicit in application code\n")
	fmt.Printf("failure_boundary=first worker error cancels sibling work through shared context\n")
	fmt.Printf("completed=%v\n", summary.Completed)
	fmt.Printf("failed=%v\n", summary.Failed)
	fmt.Printf("canceled=%v\n", summary.Canceled)
	if summary.Err != nil {
		fmt.Printf("error=%v\n", summary.Err)
	} else {
		fmt.Println("error=<nil>")
	}
	fmt.Println("---")
}

func main() {
	parent := context.Background()
	successTasks := []Task{
		{Name: "profile", Delay: 30 * time.Millisecond},
		{Name: "billing", Delay: 45 * time.Millisecond},
		{Name: "risk", Delay: 20 * time.Millisecond},
	}
	timeoutTasks := []Task{
		{Name: "profile", Delay: 30 * time.Millisecond},
		{Name: "billing", Delay: 45 * time.Millisecond},
		{Name: "risk", Delay: 120 * time.Millisecond},
	}
	failureTasks := []Task{
		{Name: "profile", Delay: 30 * time.Millisecond},
		{Name: "billing", Delay: 45 * time.Millisecond, Fail: true},
		{Name: "risk", Delay: 80 * time.Millisecond},
	}

	printSummary(orchestrate(parent, "go-success", successTasks, 100*time.Millisecond))
	printSummary(orchestrate(parent, "go-timeout", timeoutTasks, 70*time.Millisecond))
	printSummary(orchestrate(parent, "go-worker-error", failureTasks, 100*time.Millisecond))
}

// 微服务连接预算：replicas × MaxOpenConns vs PostgreSQL max_connections
package main

import (
	"fmt"
	"os"
)

type service struct {
	name     string
	replicas int
	maxOpen  int
}

func main() {
	pgMax := 100
	headroom := 0.8
	limit := int(float64(pgMax) * headroom)

	services := []service{
		{name: "order", replicas: 4, maxOpen: 25},
		{name: "inventory", replicas: 4, maxOpen: 25},
		{name: "payment", replicas: 3, maxOpen: 25},
	}

	total := 0
	fmt.Println("# 连接预算算术实验")
	fmt.Printf("\n假设 PostgreSQL max_connections=%d，预留 headroom=%.0f%% → 可用上限 %d\n\n", pgMax, headroom*100, limit)
	fmt.Println("| 服务 | 副本 | MaxOpen | 小计 | 累计 | 状态 |")
	fmt.Println("|---|---:|---:|---:|---:|---|")

	for _, s := range services {
		sub := s.replicas * s.maxOpen
		total += sub
		status := "OK"
		if total > limit {
			status = "OVER"
		}
		fmt.Printf("| %s | %d | %d | %d | %d | %s |\n", s.name, s.replicas, s.maxOpen, sub, total, status)
	}

	fmt.Printf("\n结论：三服务合计峰值连接 %d，超过 PG 可用上限 %d 共 %d。\n", total, limit, total-limit)
	if total > limit {
		os.Exit(0)
	}
}

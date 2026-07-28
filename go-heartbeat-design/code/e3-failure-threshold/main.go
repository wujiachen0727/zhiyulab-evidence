package main

import (
	"fmt"
	"math/rand"
)

func main() {
	// 模拟不同连续失败阈值下的误判率
	// 假设：网络抖动概率 = 5%（每次心跳有5%概率超时）
	// 心跳间隔 = 30s
	// 模拟 10000 次心跳，统计不同阈值下的误杀次数

	const totalHeartbeats = 10000
	const networkJitterRate = 0.05 // 5% 网络抖动

	rng := rand.New(rand.NewSource(42)) // 固定种子，可复现

	// 生成心跳结果序列：true=正常，false=超时
	results := make([]bool, totalHeartbeats)
	for i := range results {
		results[i] = rng.Float64() > networkJitterRate
	}

	fmt.Printf("=== 连续失败阈值误判率模拟 ===\n")
	fmt.Printf("总心跳次数: %d\n", totalHeartbeats)
	fmt.Printf("网络抖动率: %.0f%%\n", networkJitterRate*100)
	fmt.Printf("实际超时次数: %d\n\n", countFailures(results))

	for threshold := 1; threshold <= 5; threshold++ {
		falseKills := 0
		consecutiveFailures := 0

		for _, ok := range results {
			if !ok {
				consecutiveFailures++
				if consecutiveFailures >= threshold {
					falseKills++ // 误杀：因网络抖动达到阈值，判定连接死亡
					consecutiveFailures = 0 // 重置
				}
			} else {
				consecutiveFailures = 0
			}
		}

		fmt.Printf("阈值=%d: 误杀次数=%d, 误杀率=%.2f%%\n",
			threshold, falseKills, float64(falseKills)/float64(totalHeartbeats)*100)
	}
}

func countFailures(results []bool) int {
	count := 0
	for _, ok := range results {
		if !ok {
			count++
		}
	}
	return count
}

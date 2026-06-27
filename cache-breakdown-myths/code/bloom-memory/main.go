package main

import (
	"fmt"
	"math"
)

// bloomFilterSize 计算布隆过滤器所需的比特数和内存
// n: 元素数量, p: 期望误判率
// 公式: m = -n * ln(p) / (ln2)^2
// 最优哈希函数数量: k = (m/n) * ln2
func bloomFilterSize(n int, p float64) (bits int, bytes int, mb float64, k int) {
	ln2 := math.Log(2)
	ln2sq := ln2 * ln2

	m := -float64(n) * math.Log(p) / ln2sq
	bits = int(math.Ceil(m))
	bytes = (bits + 7) / 8
	mb = float64(bytes) / (1024 * 1024)
	k = int(math.Ceil(m / float64(n) * ln2))
	return
}

func main() {
	fmt.Println("=== 布隆过滤器内存占用实测 ===")
	fmt.Println()

	scales := []int{1_000_000, 10_000_000, 100_000_000}
	fpRates := []float64{0.05, 0.01, 0.001}

	fmt.Printf("%-15s", "数据规模\\误判率")
	for _, p := range fpRates {
		fmt.Printf("%-20s", fmt.Sprintf("%.1f%%", p*100))
	}
	fmt.Println()
	fmt.Println("-----------------------------------------------------------")

	for _, n := range scales {
		label := ""
		switch n {
		case 1_000_000:
			label = "100万 key"
		case 10_000_000:
			label = "1000万 key"
		case 100_000_000:
			label = "1亿 key"
		}
		fmt.Printf("%-15s", label)
		for _, p := range fpRates {
			_, _, mb, k := bloomFilterSize(n, p)
			fmt.Printf("%-20s", fmt.Sprintf("%.1f MB (k=%d)", mb, k))
		}
		fmt.Println()
	}

	fmt.Println()
	fmt.Println("=== 典型 Redis 场景参考 ===")
	fmt.Println()

	// 典型场景：电商商品 ID 防穿透
	n := 10_000_000 // 1000 万商品
	p := 0.01       // 1% 误判率
	bits, byteCount, mb, k := bloomFilterSize(n, p)
	fmt.Printf("场景：1000万商品ID，1%%误判率\n")
	fmt.Printf("  比特数：%d bits\n", bits)
	fmt.Printf("  字节数：%d bytes\n", byteCount)
	fmt.Printf("  内存：%.1f MB\n", mb)
	fmt.Printf("  哈希函数数：%d\n", k)
	fmt.Printf("  Redis 节点 512MB 时占比：%.1f%%\n", mb/512*100)
	fmt.Println()

	// 对比：1亿用户ID
	n2 := 100_000_000
	p2 := 0.01
	_, _, mb2, k2 := bloomFilterSize(n2, p2)
	fmt.Printf("场景：1亿用户ID，1%%误判率\n")
	fmt.Printf("  内存：%.1f MB\n", mb2)
	fmt.Printf("  哈希函数数：%d\n", k2)
	fmt.Printf("  Redis 节点 512MB 时占比：%.1f%%\n", mb2/512*100)
	fmt.Printf("  Redis 节点 2GB 时占比：%.1f%%\n", mb2/2048*100)
}

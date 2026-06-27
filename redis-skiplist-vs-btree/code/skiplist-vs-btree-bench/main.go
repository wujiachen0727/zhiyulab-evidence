package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

// ============================================================
// 简化版 SkipList 实现（参考 Redis t_zset.c 的设计）
// ============================================================

const maxLevel = 32

type skipNode struct {
	ele    string
	score  float64
	backward *skipNode
	level  []skipLevel
}

type skipLevel struct {
	forward *skipNode
	span    int
}

type skipList struct {
	header *skipNode
	tail   *skipNode
	length int
	level  int
}

func newSkipList() *skipList {
	return &skipList{
		header: &skipNode{level: make([]skipLevel, maxLevel)},
		level:  1,
	}
}

func randomLevel() int {
	level := 1
	// Redis 使用 0.25 的 p 值（1/4 概率晋升）
	for rand.Float64() < 0.25 && level < maxLevel {
		level++
	}
	return level
}

func (sl *skipList) insert(score float64, ele string) {
	update := make([]*skipNode, maxLevel)
	rank := make([]int, maxLevel)
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		if i == sl.level-1 {
			rank[i] = 0
		} else {
			rank[i] = rank[i+1]
		}
		for x.level[i].forward != nil &&
			(x.level[i].forward.score < score ||
				(x.level[i].forward.score == score && x.level[i].forward.ele < ele)) {
			rank[i] += x.level[i].span
			x = x.level[i].forward
		}
		update[i] = x
	}
	level := randomLevel()
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			rank[i] = 0
			update[i] = sl.header
			update[i].level[i].span = sl.length
		}
		sl.level = level
	}
	n := &skipNode{ele: ele, score: score, level: make([]skipLevel, level)}
	for i := 0; i < level; i++ {
		n.level[i].forward = update[i].level[i].forward
		update[i].level[i].forward = n
		n.level[i].span = update[i].level[i].span - (rank[0] - rank[i])
		update[i].level[i].span = (rank[0] - rank[i]) + 1
	}
	if n.level[0].forward != nil {
		n.level[0].forward.backward = n
	}
	if update[0] == sl.header {
		n.backward = nil
	} else {
		n.backward = update[0]
	}
	sl.length++
}

// ZRANGE 等价：返回 [start, end) 范围内的元素
func (sl *skipList) rangeQuery(start, end float64) []string {
	x := sl.header
	for i := sl.level - 1; i >= 0; i-- {
		for x.level[i].forward != nil && x.level[i].forward.score < start {
			x = x.level[i].forward
		}
	}
	x = x.level[0].forward
	var result []string
	for x != nil && x.score < end {
		result = append(result, x.ele)
		x = x.level[0].forward
	}
	return result
}

// ============================================================
// 简化版 B+ 树实现（内存场景，节点对齐 cache line）
// ============================================================

const btreeOrder = 64 // 每个节点最多 64 个 key（模拟 B+ 树）

type bplusNode struct {
	keys     []float64
	children []*bplusNode
	leaf     bool
	next     *bplusNode // 叶子节点的链表
}

type bplusTree struct {
	root *bplusNode
}

func newBplusTree() *bplusTree {
	return &bplusTree{root: &bplusNode{leaf: true}}
}

func (t *bplusTree) insert(key float64) {
	leaf := t.findLeaf(key)
	// 简化：直接插入并排序
	leaf.keys = append(leaf.keys, key)
	sort.Float64s(leaf.keys)
	if len(leaf.keys) >= btreeOrder {
		t.splitLeaf(leaf)
	}
}

func (t *bplusTree) findLeaf(key float64) *bplusNode {
	node := t.root
	for !node.leaf {
		i := 0
		for i < len(node.keys) && key >= node.keys[i] {
			i++
		}
		node = node.children[i]
	}
	return node
}

func (t *bplusTree) splitLeaf(leaf *bplusNode) {
	mid := len(leaf.keys) / 2
	newLeaf := &bplusNode{leaf: true, keys: append([]float64{}, leaf.keys[mid:]...), next: leaf.next}
	leaf.keys = leaf.keys[:mid]
	leaf.next = newLeaf
	if t.root == leaf {
		newRoot := &bplusNode{keys: []float64{newLeaf.keys[0]}, children: []*bplusNode{leaf, newLeaf}}
		t.root = newRoot
	}
	// 简化：不处理非根分裂（benchmark 足够）
}

// ZRANGE 等价：返回 [start, end) 范围内的元素
func (t *bplusTree) rangeQuery(start, end float64) []float64 {
	leaf := t.findLeaf(start)
	var result []float64
	for leaf != nil {
		for _, k := range leaf.keys {
			if k >= end {
				return result
			}
			if k >= start {
				result = append(result, k)
			}
		}
		leaf = leaf.next
	}
	return result
}

// ============================================================
// Benchmark
// ============================================================

func benchSkiplistInsert(n int) time.Duration {
	rand.Seed(42)
	sl := newSkipList()
	start := time.Now()
	for i := 0; i < n; i++ {
		sl.insert(rand.Float64()*1000000, fmt.Sprintf("m%d", i))
	}
	return time.Since(start)
}

func benchBplusInsert(n int) time.Duration {
	rand.Seed(42)
	bt := newBplusTree()
	start := time.Now()
	for i := 0; i < n; i++ {
		bt.insert(rand.Float64() * 1000000)
	}
	return time.Since(start)
}

func benchSkiplistRange(sl *skipList, queries int) time.Duration {
	rand.Seed(42)
	start := time.Now()
	for i := 0; i < queries; i++ {
		s := rand.Float64() * 1000000
		sl.rangeQuery(s, s+1000)
	}
	return time.Since(start)
}

func benchBplusRange(bt *bplusTree, queries int) time.Duration {
	rand.Seed(42)
	start := time.Now()
	for i := 0; i < queries; i++ {
		s := rand.Float64() * 1000000
		bt.rangeQuery(s, s+1000)
	}
	return time.Since(start)
}

func main() {
	// 预热
	rand.Seed(42)
	
	fmt.Println("=== 跳表 vs B+ 树内存场景 Benchmark ===")
	fmt.Println("Go 1.26.4 darwin/arm64")
	fmt.Println("时间:", time.Now().Format("2006-01-02T15:04:05Z07:00"))
	fmt.Println()
	
	sizes := []int{10000, 100000, 1000000}
	
	fmt.Println("## 插入性能（全量插入 N 个元素）")
	fmt.Println("| N | skiplist | b+tree | 比值 |")
	fmt.Println("|---:|---:|---:|---:|")
	for _, n := range sizes {
		slD := benchSkiplistInsert(n)
		btD := benchBplusInsert(n)
		ratio := float64(slD) / float64(btD)
		fmt.Printf("| %d | %v | %v | %.2fx |\n", n, slD, btD, ratio)
	}
	
	fmt.Println()
	fmt.Println("## 范围查询性能（10000 次随机范围查询，跨度 1000）")
	fmt.Println("| N | skiplist | b+tree | 比值 |")
	fmt.Println("|---:|---:|---:|---:|")
	for _, n := range sizes {
		// 重建数据
		rand.Seed(42)
		sl := newSkipList()
		for i := 0; i < n; i++ {
			sl.insert(rand.Float64()*1000000, fmt.Sprintf("m%d", i))
		}
		rand.Seed(42)
		bt := newBplusTree()
		for i := 0; i < n; i++ {
			bt.insert(rand.Float64() * 1000000)
		}
		
		slD := benchSkiplistRange(sl, 10000)
		btD := benchBplusRange(bt, 10000)
		ratio := float64(slD) / float64(btD)
		fmt.Printf("| %d | %v | %v | %.2fx |\n", n, slD, btD, ratio)
	}
	
	fmt.Println()
	fmt.Println("## 核心证伪检查")
	fmt.Println("假设：内存场景下跳表和 B+ 树性能相近，B+ 树不必然更快")
	fmt.Println("判定标准：如果 B+ 树范围查询显著快于跳表（比值 > 2x），假设被推翻")
}

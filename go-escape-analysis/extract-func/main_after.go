package main

import (
	"fmt"
	"strings"
)

// 提取的轻量函数：trimAndUpper 会被内联
// 预期：内联后调用处不再有函数边界，data 不逃逸
func trimAndUpper(s string) string {
	return strings.TrimSpace(strings.ToUpper(s))
}

// 提取后 processData 变得更轻，但仍可能超出预算
// 关键变化：trimAndUpper 被内联后，函数边界减少
func processData(data string) string {
	parts := strings.Split(data, ",")
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "-"
		}
		result += trimAndUpper(p)
	}
	transformed := strings.ReplaceAll(result, "A", "4")
	transformed = strings.ReplaceAll(transformed, "E", "3")
	transformed = strings.ReplaceAll(transformed, "I", "1")
	transformed = strings.ReplaceAll(transformed, "O", "0")
	transformed = strings.ReplaceAll(transformed, "S", "5")
	return transformed
}

func main() {
	input := "hello, world, golang, escape, analysis"
	output := processData(input)
	fmt.Println(output)
}

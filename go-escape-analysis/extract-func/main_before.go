package main

import (
	"fmt"
	"strings"
)

// 大函数：超过内联预算，不会被编译器内联
// 预期：processData 不会被内联，data 参数逃逸到堆上
func processData(data string) string {
	parts := strings.Split(data, ",")
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += "-"
		}
		result += strings.TrimSpace(strings.ToUpper(p))
	}
	// 增加函数复杂度，确保超出内联预算
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

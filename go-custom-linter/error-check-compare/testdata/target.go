// 测试目标文件：10 种场景，对比 go/ast 和 go/types 的检测精度
// [实测 Go 1.26.2]
package testdata

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ========== 真阳性：确实忽略了危险的 error ==========

// 场景1：完全丢弃 os.Remove 的 error（真阳性）
func discardRemove() {
	os.Remove("temp.txt")
}

// 场景2：用 _ 丢弃 os.Open 的 error（真阳性）
func discardOpen() {
	f, _ := os.Open("data.txt")
	_ = f
}

// 场景3：丢弃 strconv.Atoi 的 error（真阳性）
func discardAtoi() {
	v, _ := strconv.Atoi("not-a-number")
	_ = v
}

// 场景4：丢弃 os.MkdirAll 的 error（真阳性）
func discardMkdir() {
	os.MkdirAll("/tmp/test/dir", 0755)
}

// ========== AST 假阳性：不返回 error，AST 仍会标记 ==========

// 场景5：fmt.Println 返回 (int, error)，但 AST 不知道类型
// AST 看到 ExprStmt + CallExpr → 标记；types 知道它返回 error
// 但在工程实践中这是"可接受的忽略"，不是 bug
func printlnCall() {
	fmt.Println("hello world")
}

// 场景6：strings.Contains 返回 bool，不返回 error
// AST 看到函数调用结果被丢弃 → 标记（但实际无 error）
func containsCall() {
	strings.Contains("hello", "lo")
}

// 场景7：_ = f 赋值，f 是 *os.File 不是 error
// AST 看到 _ 赋值 → 标记（但被丢弃的不是 error）
func blankAssignNonError() {
	f, err := os.Open("file.txt")
	if err != nil {
		return
	}
	_ = f
}

// 场景8：doNothing() 无返回值
// AST 看到 ExprStmt + CallExpr → 标记（但函数无返回值）
func callNoReturn() {
	doNothing()
}

// ========== 真阴性：两种方法都不应标记 ==========

// 场景9：正确处理了 error
func properHandling() {
	f, err := os.Open("config.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
}

// 场景10：赋值给具名变量（不是 _）
func namedAssign() {
	n, err := fmt.Println("test")
	if err != nil {
		return
	}
	_ = n
}

func doNothing() {
	// 无返回值的辅助函数
}

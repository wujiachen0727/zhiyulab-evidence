// singlechecker 入口：一行代码把 Analyzer 变成可执行的命令行工具
// [实测 Go 1.26.2]
package main

import "golang.org/x/tools/go/analysis/singlechecker"

func main() {
	singlechecker.Main(Analyzer)
}

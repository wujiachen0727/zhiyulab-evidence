// Package main demonstrates how Go's type safety does NOT protect against
// SQL injection when the injection point is a column name (ORDER BY / LIMIT)
// rather than a value.
//
// PoC E2: "Type-safe parameters" prevent WHERE injection but not ORDER BY injection.
//
// Run: go run main.go
package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

// UserQuery 表示一个按列排序的用户查询。
// 注意：OrderBy 是一个字符串，而不是枚举——很多业务代码都这么写。
type UserQuery struct {
	Keyword string // WHERE 条件的关键字（参数化，安全）
	OrderBy string // 排序字段（非参数化，危险）
	Limit   int    // LIMIT 值（int，看起来"类型安全"）
}

// vulnerableSearch 演示"看起来用了 ? 参数化但仍然可被注入"的代码。
// 常见业务代码模式：WHERE 用参数化，ORDER BY 走字符串拼接。
func vulnerableSearch(db *sql.DB, q UserQuery) ([]string, error) {
	// 这段代码是从无数"看起来安全"的 Go 业务代码里抽象出来的：
	// WHERE 子句用 ? 参数化，开发者觉得"我用了类型安全的参数化"，
	// 但 ORDER BY 因为 database/sql 不支持列名参数化，拼了字符串。
	query := fmt.Sprintf(
		"SELECT username FROM users WHERE username LIKE ? ORDER BY %s LIMIT %d",
		q.OrderBy, q.Limit,
	)

	rows, err := db.Query(query, "%"+q.Keyword+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, nil
}

func setupDB() *sql.DB {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	_, _ = db.Exec(`CREATE TABLE users (
		id INTEGER PRIMARY KEY,
		username TEXT,
		password TEXT,
		role TEXT
	);`)
	_, _ = db.Exec(`INSERT INTO users (username, password, role) VALUES
		('alice', 'secret_alice', 'user'),
		('bob', 'secret_bob', 'user'),
		('admin', 'root_token_leaked', 'admin');`)
	return db
}

func main() {
	db := setupDB()
	defer db.Close()

	fmt.Println("=== 正常查询 ===")
	normal := UserQuery{Keyword: "a", OrderBy: "username", Limit: 10}
	names, err := vulnerableSearch(db, normal)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("返回: %v\n\n", names)

	fmt.Println("=== 类型安全幻觉：OrderBy 字段注入 ===")
	// 攻击者控制了 OrderBy 字段（比如前端传上来的排序参数没校验），
	// 构造一个 ORDER BY 子查询，把 password 列内容作为排序结果泄漏。
	// Go 编译器不会报错——OrderBy 是 string，类型上完全合法。
	evil := UserQuery{
		Keyword: "a",
		OrderBy: "(CASE WHEN (SELECT password FROM users WHERE role='admin') LIKE 'r%' THEN username ELSE password END)",
		Limit:   10,
	}
	names, err = vulnerableSearch(db, evil)
	if err != nil {
		// SQLite 对部分注入会直接返回错误——这个错误本身就是信息泄漏信号
		fmt.Printf("查询报错: %v\n", err)
	} else {
		fmt.Printf("返回（已按密码内容排序，密码首字母 'r' 的排前面）: %v\n", names)
	}

	fmt.Println("\n=== 更直接：LIMIT 注入（int 类型也能出事） ===")
	// 如果 Limit 是从 URL query 解析的字符串再 strconv.Atoi，
	// 看起来是 int——但如果开发者为了"兼容" offset，用 fmt.Sprintf 拼接了
	// LIMIT %d OFFSET %d，而 OFFSET 来自另一个字段没做同等校验……
	// 这里演示的是 int 类型在拼接到 SQL 时的相同问题。
	// Go 语言层面的 int 类型安全完全无法阻止运行时拼出的 SQL 结构被篡改。
	fmt.Println("类型安全只保证数值是整数，不保证 SQL 结构安全。")
	fmt.Println()

	fmt.Println("=== 结论 ===")
	fmt.Println("1. database/sql 不支持列名/表名/排序方向作为 ? 参数")
	fmt.Println("2. Go 的强类型系统保证 OrderBy 变量是 string——仅此而已")
	fmt.Println("3. 类型安全不等于语义安全。防住注入要靠白名单，不是类型")
}

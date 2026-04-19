// [实测结构，代码逻辑为演示]
// E1: 无分层 Gin 项目 —— 所有代码塞在一个文件里
// 用途：展示"能跑"的最小 Gin 项目长什么样，以及它为什么难维护
package main

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

// 配置用全局变量 —— 任何地方都能改，任何地方都在用
var db *sql.DB

func main() {
	var err error
	// 数据库连接写死在 main 里
	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/myapp")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/users", listUsers)
	r.GET("/users/:id", getUser)
	r.POST("/users", createUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)

	r.Run(":8080")
}

// ============================================================
// 所有 handler 直接操作数据库，没有中间层
// ============================================================

func listUsers(c *gin.Context) {
	// 数据库操作直接写在 handler 里
	rows, err := db.Query("SELECT id, name, email FROM users")
	if err != nil {
		// 错误处理：每个 handler 各写各的，格式不统一
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		log.Printf("查询用户列表失败: %v", err) // 日志也是各写各的
		return
	}
	defer rows.Close()

	var users []gin.H
	for rows.Next() {
		var id int
		var name, email string
		if err := rows.Scan(&id, &name, &email); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "数据解析失败"})
			return
		}
		users = append(users, gin.H{"id": id, "name": name, "email": email})
	}
	c.JSON(http.StatusOK, users)
}

func getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var name, email string
	// 数据库操作直接写在 handler 里 —— 和 listUsers 的 SQL 有重复
	err = db.QueryRow("SELECT name, email FROM users WHERE id = ?", id).Scan(&name, &email)
	if err == sql.ErrNoRows {
		// 同样的"用户不存在"判断，每个 handler 写一遍
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		log.Printf("查询用户 %d 失败: %v", id, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "name": name, "email": email})
}

func createUser(c *gin.Context) {
	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 业务校验也混在 handler 里
	if req.Name == "" || req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "姓名和邮箱不能为空"})
		return
	}

	// 数据库操作直接写在 handler 里
	result, err := db.Exec(
		"INSERT INTO users (name, email) VALUES (?, ?)",
		req.Name, req.Email,
	)
	if err != nil {
		// 错误处理没有区分"邮箱重复"和"数据库故障"
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建失败"})
		log.Printf("创建用户失败: %v", err)
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, gin.H{"id": id, "name": req.Name, "email": req.Email})
}

func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	var req struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数格式错误"})
		return
	}

	// 数据库操作直接写在 handler 里
	result, err := db.Exec(
		"UPDATE users SET name = ?, email = ? WHERE id = ?",
		req.Name, req.Email, id,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新失败"})
		log.Printf("更新用户 %d 失败: %v", id, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"id": id, "name": req.Name, "email": req.Email})
}

func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的用户ID"})
		return
	}

	// 数据库操作直接写在 handler 里
	result, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除失败"})
		log.Printf("删除用户 %d 失败: %v", id, err)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "用户不存在"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ============================================================
// 问题总结（这个文件的问题，在项目小时感觉不到，长大后全是债）：
//
// 1. handler 直接持有 sql.DB —— 写单测时没法 mock 数据库
// 2. SQL 散落在各个 handler —— 改表结构要改 5 个地方
// 3. 错误处理各写各的 —— 返回格式不统一，日志级别随缘
// 4. 业务校验和 HTTP 处理混在一起 —— 想复用"创建用户"逻辑？复制粘贴
// 5. 全局变量 db —— 并发安全靠 sql.DB 自身，但测试隔离无解
// ============================================================

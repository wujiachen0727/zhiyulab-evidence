// [实测结构，代码逻辑为演示]
// E4: 错误处理 —— 散落模式
// 用途：展示 error 处理散落在每个 handler 的典型问题
package errors_demo

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 问题：每个 handler 各自处理错误，格式和行为不统一
// ============================================================

var db *sql.DB

// 问题 1: getUser 返回 "error" 字段
func getUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id 无效"})
		return
	}

	var name string
	err = db.QueryRow("SELECT name FROM users WHERE id = ?", id).Scan(&name)
	if err == sql.ErrNoRows {
		// 用字符串判断错误类型 —— 其他 handler 可能写法不一样
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到用户"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "服务器错误"})
		log.Printf("getUser 失败: %v", err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"name": name})
}

// 问题 2: createUser 返回 "message" 字段（和 getUser 的 "error" 不统一）
func createUser(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// 有的 handler 用 "message"，有的用 "error"，有的用 "msg"
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数错误"})
		return
	}
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "名字不能为空"})
		return
	}

	_, err := db.Exec("INSERT INTO users (name) VALUES (?)", req.Name)
	if err != nil {
		// 日志级别随缘 —— 有的用 Printf，有的用 Println，有的压根不打
		log.Println("插入失败", err)
		c.JSON(500, gin.H{"message": "创建失败"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "成功"})
}

// 问题 3: deleteUser 直接用数字状态码，还把内部错误暴露给客户端
func deleteUser(c *gin.Context) {
	id := c.Param("id")
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	if err != nil {
		// 把数据库错误直接返回给前端 —— 安全隐患
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"ok": true})
}

// ============================================================
// 散落模式的典型问题：
// 1. 错误响应字段不统一：error / message / msg，前端解析像猜谜
// 2. 状态码使用不一致：有人用 http.StatusXxx 常量，有人直接写数字
// 3. 日志格式随缘：Printf / Println / 不打，出了问题难以排查
// 4. 内部错误泄漏：err.Error() 直接返回，暴露数据库信息
// 5. 没有错误分类：调用方无法用程序判断是"用户不存在"还是"服务器故障"
// ============================================================

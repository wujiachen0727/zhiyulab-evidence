// [实测结构，代码逻辑为演示]
// E4: 错误处理 —— 统一中间件模式
// 用途：展示自定义错误类型 + 中间件统一处理的方案
package errors_demo

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ============================================================
// 第一步：定义统一的业务错误类型
// ============================================================

// AppError 业务错误 —— 携带 HTTP 状态码和错误码
type AppError struct {
	Code    string // 业务错误码，前端用这个做逻辑判断
	Message string // 用户可见的错误信息
	Status  int    // HTTP 状态码
	Err     error  // 原始错误（不暴露给客户端，只进日志）
}

func (e *AppError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

// 预定义错误 —— 整个项目复用，不用每个 handler 自己编
var (
	ErrNotFound = func(msg string) *AppError {
		return &AppError{Code: "NOT_FOUND", Message: msg, Status: http.StatusNotFound}
	}
	ErrBadRequest = func(msg string) *AppError {
		return &AppError{Code: "BAD_REQUEST", Message: msg, Status: http.StatusBadRequest}
	}
	ErrInternal = func(err error) *AppError {
		return &AppError{
			Code: "INTERNAL_ERROR", Message: "服务器内部错误",
			Status: http.StatusInternalServerError, Err: err,
		}
	}
)

// ============================================================
// 第二步：错误处理中间件 —— 所有 handler 的错误都在这里统一处理
// ============================================================

// ErrorHandler 统一错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误被设置
		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err
		var appErr *AppError
		if errors.As(err, &appErr) {
			// 业务错误 —— 返回结构化响应
			resp := gin.H{
				"code":    appErr.Code,
				"message": appErr.Message,
			}
			// 内部错误只进日志，不暴露给客户端
			if appErr.Err != nil {
				log.Printf("[%s] %s | 原因: %v", appErr.Code, appErr.Message, appErr.Err)
			}
			c.JSON(appErr.Status, resp)
		} else {
			// 未分类错误 —— 兜底处理
			log.Printf("[UNKNOWN] 未分类错误: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "服务器内部错误",
			})
		}
	}
}

// ============================================================
// 第三步：handler 变得很干净 —— 只管业务逻辑，不管错误格式
// ============================================================

func getUserLayered(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		// 不需要自己写 c.JSON —— 把错误丢给中间件
		_ = c.Error(ErrBadRequest("无效的用户ID"))
		return
	}

	// 假设从 service 层拿到了业务错误
	user, err := mockGetUserByID(id)
	if err != nil {
		_ = c.Error(err) // 直接把业务错误往上抛
		return
	}
	c.JSON(http.StatusOK, user)
}

func createUserLayered(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(ErrBadRequest("参数格式错误"))
		return
	}
	if req.Name == "" {
		_ = c.Error(ErrBadRequest("名字不能为空"))
		return
	}

	user, err := mockCreateUser(req.Name)
	if err != nil {
		_ = c.Error(err)
		return
	}
	c.JSON(http.StatusCreated, user)
}

// ============================================================
// mock 函数（模拟 service 层）
// ============================================================

func mockGetUserByID(id int64) (map[string]any, error) {
	if id == 999 {
		return nil, ErrNotFound("用户不存在")
	}
	return map[string]any{"id": id, "name": "张三"}, nil
}

func mockCreateUser(name string) (map[string]any, error) {
	return map[string]any{"id": 1, "name": name}, nil
}

// ============================================================
// 统一模式的收益：
// 1. 响应格式统一：前端只需解析 code + message，写一次就够
// 2. 内部错误不泄漏：原始 error 只进日志，客户端看不到数据库信息
// 3. handler 更干净：不需要每个 handler 重复写 c.JSON + log.Printf
// 4. 新增错误类型只改一处：加一个 ErrXxx 构造函数，所有 handler 自动可用
// 5. 可测试：AppError 是普通 struct，测试时直接 errors.As 判断
// ============================================================

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// ========================================
// 版本1：无分层的错误处理——底层错误直接暴露到HTTP响应
// ========================================

// 模拟数据库错误
var errDBTimeout = errors.New("pq: connection refused (SQLSTATE 08006), database=users_db, query=SELECT id, email FROM users WHERE id=$1")

type UserHandlerV1 struct {
	db *sql.DB
}

func (h *UserHandlerV1) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	// 直接返回数据库错误到HTTP响应——无分层
	var user struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	err := h.db.QueryRow("SELECT id, email FROM users WHERE id = ?", id).Scan(&user.ID, &user.Email)
	if err != nil {
		// 错误1：把底层错误直接暴露给客户端
		http.Error(w, fmt.Sprintf("query failed: %v", errDBTimeout), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ========================================
// 版本2：三层错误分层——底层错误在Service层被翻译
// ========================================

// 领域错误定义（Service层）
var (
	ErrUserNotFound      = errors.New("user not found")
	ErrServiceUnavailable = errors.New("service temporarily unavailable")
)

type UserService struct {
	db *sql.DB
}

func (s *UserService) GetUser(id string) (*struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}, error) {
	var user struct {
		ID    int    `json:"id"`
		Email string `json:"email"`
	}
	err := s.db.QueryRow("SELECT id, email FROM users WHERE id = ?", id).Scan(&user.ID, &user.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		// Infra层错误翻译为领域错误——不暴露底层细节
		return nil, ErrServiceUnavailable
	}
	return &user, nil
}

type UserHandlerV2 struct {
	service *UserService
}

func (h *UserHandlerV2) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	user, err := h.service.GetUser(id)
	if err != nil {
		switch {
		case errors.Is(err, ErrUserNotFound):
			http.Error(w, `{"error":"user not found","code":"NOT_FOUND"}`, http.StatusNotFound)
		case errors.Is(err, ErrServiceUnavailable):
			// 只返回通用错误+traceID，不暴露内部信息
			http.Error(w, `{"error":"service temporarily unavailable","code":"SERVICE_UNAVAILABLE","trace_id":"abc-123"}`, http.StatusServiceUnavailable)
		default:
			http.Error(w, `{"error":"internal error","code":"INTERNAL"}`, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// ========================================
// 测试：对比两个版本的HTTP响应
// ========================================

func main() {
	// 用SQLite模拟（会因文件不存在而报错，正好用于演示）
	db, _ := sql.Open("sqlite3", ":memory:")

	fmt.Println("=== 版本1：无分层——底层错误直接暴露 ===")
	handlerV1 := &UserHandlerV1{db: db}
	req := httptest.NewRequest("GET", "/user?id=1", nil)
	w := httptest.NewRecorder()
	handlerV1.GetUser(w, req)
	fmt.Printf("HTTP Status: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	// 检查泄露信息
	body := w.Body.String()
	leaked := []string{"pq:", "SQLSTATE", "database=", "SELECT", "users"}
	fmt.Println("\n--- 信息泄露检测 ---")
	for _, keyword := range leaked {
		if strings.Contains(body, keyword) {
			fmt.Printf("⚠️  泄露了: '%s'\n", keyword)
		} else {
			fmt.Printf("✅ 未泄露: '%s'\n", keyword)
		}
	}

	fmt.Println("\n=== 版本2：三层分层——底层错误被翻译 ===")
	service := &UserService{db: db}
	handlerV2 := &UserHandlerV2{service: service}
	req2 := httptest.NewRequest("GET", "/user?id=1", nil)
	w2 := httptest.NewRecorder()
	handlerV2.GetUser(w2, req2)
	fmt.Printf("HTTP Status: %d\n", w2.Code)
	fmt.Printf("Response Body: %s\n", w2.Body.String())

	body2 := w2.Body.String()
	fmt.Println("\n--- 信息泄露检测 ---")
	for _, keyword := range leaked {
		if strings.Contains(body2, keyword) {
			fmt.Printf("⚠️  泄露了: '%s'\n", keyword)
		} else {
			fmt.Printf("✅ 未泄露: '%s'\n", keyword)
		}
	}
}

// 简单三层架构：用户注册
// 三个文件搞定：handler → service → repo
package main

// ============================================================
// 文件 1: handler.go — HTTP 入口
// ============================================================

/*
package handler

import (
	"encoding/json"
	"net/http"

	"app/service"
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	user, err := service.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(RegisterResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
*/

// ============================================================
// 文件 2: service.go — 业务逻辑
// ============================================================

/*
package service

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"unicode"

	"app/repo"
)

type User struct {
	ID    string
	Email string
	Name  string
}

func Register(ctx context.Context, email, password, name string) (*User, error) {
	// 校验邮箱格式
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, errors.New("invalid email format")
	}

	// 校验密码强度（>=8位，含大小写+数字）
	if err := validatePassword(password); err != nil {
		return nil, err
	}

	// 检查邮箱是否已注册
	exists, err := repo.EmailExists(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email: %w", err)
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// 创建用户
	user, err := repo.CreateUser(ctx, email, password, name)
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 发送欢迎邮件（异步，不阻塞注册流程）
	go sendWelcomeEmail(user.Email, user.Name)

	return &User{ID: user.ID, Email: user.Email, Name: user.Name}, nil
}

func validatePassword(pwd string) error {
	if len(pwd) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range pwd {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasDigit = true
		}
	}
	if !hasUpper || !hasLower || !hasDigit {
		return errors.New("password must contain upper, lower case and digit")
	}
	return nil
}

func sendWelcomeEmail(email, name string) {
	// 实际项目中调用邮件服务
	fmt.Printf("Sending welcome email to %s (%s)\n", name, email)
}
*/

// ============================================================
// 文件 3: repo.go — 数据访问
// ============================================================

/*
package repo

import (
	"context"
	"database/sql"
)

type UserRecord struct {
	ID    string
	Email string
	Name  string
}

func EmailExists(ctx context.Context, email string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email = ?", email).Scan(&count)
	return count > 0, err
}

func CreateUser(ctx context.Context, email, password, name string) (*UserRecord, error) {
	hashedPwd := hashPassword(password)
	result, err := db.ExecContext(ctx,
		"INSERT INTO users (email, password_hash, name) VALUES (?, ?, ?)",
		email, hashedPwd, name)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return &UserRecord{ID: fmt.Sprintf("%d", id), Email: email, Name: name}, nil
}
*/

// ============================================================
// 统计
// ============================================================
// 文件数：3（handler.go, service.go, repo.go）
// 总代码行数（不含空行和注释）：约 95 行
// 依赖层数：2（handler → service → repo）
// 新人理解路径：handler.RegisterHandler → service.Register → repo.CreateUser
// 跳转次数：2 次

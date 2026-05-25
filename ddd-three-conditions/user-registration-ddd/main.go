// DDD 战术模式架构：用户注册
// 七个文件：controller → app service → domain service → entity → repository(interface) → repository(impl) → event
package main

// ============================================================
// 文件 1: controller.go — HTTP 入口（表现层）
// ============================================================

/*
package controller

import (
	"encoding/json"
	"net/http"

	"app/application"
	"app/application/dto"
)

type UserController struct {
	appService *application.UserApplicationService
}

func (c *UserController) Register(w http.ResponseWriter, r *http.Request) {
	var req dto.RegisterCommand
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	result, err := c.appService.Register(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(result)
}
*/

// ============================================================
// 文件 2: dto.go — 数据传输对象（应用层）
// ============================================================

/*
package dto

type RegisterCommand struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type RegisterResult struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

// Assembler: Command → Domain 参数
func (cmd RegisterCommand) ToRegistrationParams() (email, password, name string) {
	return cmd.Email, cmd.Password, cmd.Name
}

// Assembler: Domain Entity → Result DTO
func FromUser(id, email string) RegisterResult {
	return RegisterResult{ID: id, Email: email}
}
*/

// ============================================================
// 文件 3: user_application_service.go — 应用服务（编排层）
// ============================================================

/*
package application

import (
	"context"
	"fmt"

	"app/application/dto"
	"app/domain/service"
	"app/domain/event"
)

type UserApplicationService struct {
	domainService  *service.UserDomainService
	eventPublisher event.Publisher
}

func (s *UserApplicationService) Register(ctx context.Context, cmd dto.RegisterCommand) (*dto.RegisterResult, error) {
	email, password, name := cmd.ToRegistrationParams()

	user, err := s.domainService.RegisterUser(ctx, email, password, name)
	if err != nil {
		return nil, fmt.Errorf("register user: %w", err)
	}

	// 发布领域事件
	s.eventPublisher.Publish(event.UserRegistered{
		UserID: user.ID().String(),
		Email:  user.Email().Value(),
		Name:   user.Name(),
	})

	result := dto.FromUser(user.ID().String(), user.Email().Value())
	return &result, nil
}
*/

// ============================================================
// 文件 4: user_domain_service.go — 领域服务
// ============================================================

/*
package service

import (
	"context"
	"errors"

	"app/domain/entity"
	"app/domain/repository"
	"app/domain/valueobject"
)

type UserDomainService struct {
	userRepo repository.UserRepository
}

func (s *UserDomainService) RegisterUser(ctx context.Context, email, password, name string) (*entity.User, error) {
	// 构造值对象（校验在值对象内部）
	emailVO, err := valueobject.NewEmail(email)
	if err != nil {
		return nil, err
	}

	passwordVO, err := valueobject.NewPassword(password)
	if err != nil {
		return nil, err
	}

	// 唯一性检查（领域规则）
	exists, err := s.userRepo.ExistsByEmail(ctx, emailVO)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("email already registered")
	}

	// 创建聚合根
	user := entity.NewUser(emailVO, passwordVO, name)

	// 持久化
	if err := s.userRepo.Save(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}
*/

// ============================================================
// 文件 5: user_entity.go — 聚合根（领域层）
// ============================================================

/*
package entity

import (
	"app/domain/valueobject"
	"github.com/google/uuid"
)

type UserID struct {
	value string
}

func NewUserID() UserID {
	return UserID{value: uuid.New().String()}
}

func (id UserID) String() string { return id.value }

type User struct {
	id       UserID
	email    valueobject.Email
	password valueobject.Password
	name     string
}

func NewUser(email valueobject.Email, password valueobject.Password, name string) *User {
	return &User{
		id:       NewUserID(),
		email:    email,
		password: password,
		name:     name,
	}
}

func (u *User) ID() UserID                  { return u.id }
func (u *User) Email() valueobject.Email     { return u.email }
func (u *User) Password() valueobject.Password { return u.password }
func (u *User) Name() string                 { return u.name }
*/

// ============================================================
// 文件 5b: value_objects.go — 值对象
// ============================================================

/*
package valueobject

import (
	"errors"
	"net/mail"
	"unicode"
)

// Email 值对象——校验封装在构造函数中
type Email struct {
	value string
}

func NewEmail(raw string) (Email, error) {
	if _, err := mail.ParseAddress(raw); err != nil {
		return Email{}, errors.New("invalid email format")
	}
	return Email{value: raw}, nil
}

func (e Email) Value() string { return e.value }

// Password 值对象
type Password struct {
	hashed string
}

func NewPassword(raw string) (Password, error) {
	if len(raw) < 8 {
		return Password{}, errors.New("password must be at least 8 characters")
	}
	var hasUpper, hasLower, hasDigit bool
	for _, c := range raw {
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
		return Password{}, errors.New("password must contain upper, lower case and digit")
	}
	return Password{hashed: hashPassword(raw)}, nil
}
*/

// ============================================================
// 文件 6: user_repository.go — Repository 接口（领域层）
// ============================================================

/*
package repository

import (
	"context"

	"app/domain/entity"
	"app/domain/valueobject"
)

type UserRepository interface {
	Save(ctx context.Context, user *entity.User) error
	ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error)
	FindByID(ctx context.Context, id entity.UserID) (*entity.User, error)
}
*/

// ============================================================
// 文件 7: user_repository_mysql.go — Repository 实现（基础设施层）
// ============================================================

/*
package infrastructure

import (
	"context"
	"database/sql"

	"app/domain/entity"
	"app/domain/repository"
	"app/domain/valueobject"
)

type MySQLUserRepository struct {
	db *sql.DB
}

func (r *MySQLUserRepository) Save(ctx context.Context, user *entity.User) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO users (id, email, password_hash, name) VALUES (?, ?, ?, ?)",
		user.ID().String(),
		user.Email().Value(),
		user.Password().HashedValue(),
		user.Name(),
	)
	return err
}

func (r *MySQLUserRepository) ExistsByEmail(ctx context.Context, email valueobject.Email) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM users WHERE email = ?", email.Value()).Scan(&count)
	return count > 0, err
}

func (r *MySQLUserRepository) FindByID(ctx context.Context, id entity.UserID) (*entity.User, error) {
	// ... 省略重建逻辑
	return nil, nil
}
*/

// ============================================================
// 文件 8: events.go — 领域事件（领域层）
// ============================================================

/*
package event

type UserRegistered struct {
	UserID string
	Email  string
	Name   string
}

type Publisher interface {
	Publish(event interface{})
}
*/

// ============================================================
// 文件 9: event_handler.go — 事件处理器（应用/基础设施层）
// ============================================================

/*
package handler

import "fmt"

type WelcomeEmailHandler struct{}

func (h *WelcomeEmailHandler) Handle(evt event.UserRegistered) {
	fmt.Printf("Sending welcome email to %s (%s)\n", evt.Name, evt.Email)
}
*/

// ============================================================
// 统计
// ============================================================
// 文件数：7-9（controller, dto, app_service, domain_service, entity+value_objects, repository_interface, repository_impl, event, event_handler）
// 核心文件：7（不含 value_objects 和 event_handler 拆分）
// 总代码行数（不含空行和注释）：约 210 行
// 依赖层数：4（controller → app_service → domain_service → repository）
// 新人理解路径：
//   controller.Register
//   → dto.RegisterCommand (了解参数结构)
//   → application.UserApplicationService.Register (编排逻辑)
//   → service.UserDomainService.RegisterUser (业务规则)
//   → entity.NewUser + valueobject.NewEmail + valueobject.NewPassword (领域模型)
//   → repository.UserRepository.Save (接口)
//   → infrastructure.MySQLUserRepository.Save (实现)
// 跳转次数：6 次（含理解值对象和 Repository 接口/实现分离）

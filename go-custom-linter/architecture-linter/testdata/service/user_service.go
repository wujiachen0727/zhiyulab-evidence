// Package service 提供业务逻辑层
// [实测 Go 1.26.2]
package service

import "example.com/testdata/repository"

// UserService 用户业务逻辑
type UserService struct {
	repo repository.UserRepository // 依赖接口，不依赖具体实现
}

// NewUserService 构造函数注入
func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

// GetUser 通过 service 层获取用户——合规路径
func (s *UserService) GetUser(id int) (string, error) {
	return s.repo.FindByID(id)
}

// RemoveUser 通过 service 层删除用户——合规路径
func (s *UserService) RemoveUser(id int) error {
	return s.repo.DeleteByID(id)
}

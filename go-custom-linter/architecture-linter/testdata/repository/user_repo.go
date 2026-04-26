// Package repository 提供数据访问层
// [实测 Go 1.26.2]
package repository

import "fmt"

// UserRepo 用户数据仓库
type UserRepo struct{}

// FindByID 根据 ID 查找用户
func (r *UserRepo) FindByID(id int) (string, error) {
	return fmt.Sprintf("user-%d", id), nil
}

// DeleteByID 根据 ID 删除用户
func (r *UserRepo) DeleteByID(id int) error {
	return nil
}

// UserRepository 接口——供 service 层依赖注入
type UserRepository interface {
	FindByID(id int) (string, error)
	DeleteByID(id int) error
}

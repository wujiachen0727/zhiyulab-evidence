// [实测结构，代码逻辑为演示]
// E2: 业务逻辑层 —— 不知道 HTTP，不知道数据库
package service

import (
	"errors"

	"go-web-demo/model"
	"go-web-demo/repository"
)

// 业务层预定义错误 —— handler 层用 errors.Is() 判断
var (
	ErrUserNotFound = errors.New("用户不存在")
	ErrInvalidInput = errors.New("输入参数无效")
	ErrEmailExists  = errors.New("邮箱已被注册")
)

// UserService 用户业务逻辑
// 依赖 UserRepository 接口，不依赖 MySQL
type UserService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) List() (users []model.User, err error) {
	users, err = s.repo.List()
	return
}

func (s *UserService) GetByID(id int64) (user model.User, err error) {
	user, err = s.repo.GetByID(id)
	if err != nil {
		// 业务逻辑层翻译数据层错误 —— handler 不需要知道 sql.ErrNoRows
		err = ErrUserNotFound
	}
	return
}

func (s *UserService) Create(name, email string) (user model.User, err error) {
	// 业务校验集中在 service 层
	if name == "" || email == "" {
		err = ErrInvalidInput
		return
	}

	user = model.User{Name: name, Email: email}
	err = s.repo.Create(&user)
	return
}

func (s *UserService) Update(id int64, name, email string) (user model.User, err error) {
	if name == "" || email == "" {
		err = ErrInvalidInput
		return
	}

	user = model.User{ID: id, Name: name, Email: email}
	err = s.repo.Update(&user)
	return
}

func (s *UserService) Delete(id int64) (err error) {
	err = s.repo.Delete(id)
	return
}

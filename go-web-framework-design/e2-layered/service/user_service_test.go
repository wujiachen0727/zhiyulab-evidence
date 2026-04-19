// [实测结构，代码逻辑为演示]
// E2: 单元测试示例 —— 展示分层后如何轻松 mock 测试
package service

import (
	"testing"

	"go-web-demo/model"
)

// mockUserRepo 测试用的 mock 实现
// 关键：因为 service 依赖的是 UserRepository 接口，
// 不需要任何 mock 框架，手写一个 struct 就能替换
type mockUserRepo struct {
	users  []model.User
	lastID int64
}

func (m *mockUserRepo) List() (users []model.User, err error) {
	users = m.users
	return
}

func (m *mockUserRepo) GetByID(id int64) (user model.User, err error) {
	for _, u := range m.users {
		if u.ID == id {
			user = u
			return
		}
	}
	err = ErrUserNotFound
	return
}

func (m *mockUserRepo) Create(user *model.User) (err error) {
	m.lastID++
	user.ID = m.lastID
	m.users = append(m.users, *user)
	return
}

func (m *mockUserRepo) Update(user *model.User) (err error) {
	for i, u := range m.users {
		if u.ID == user.ID {
			m.users[i] = *user
			return
		}
	}
	err = ErrUserNotFound
	return
}

func (m *mockUserRepo) Delete(id int64) (err error) {
	for i, u := range m.users {
		if u.ID == id {
			m.users = append(m.users[:i], m.users[i+1:]...)
			return
		}
	}
	err = ErrUserNotFound
	return
}

// TestCreateUser_Success 创建用户成功
// 注意：这个测试不需要数据库、不需要网络、不需要 Docker
// 跑完只要几毫秒
func TestCreateUser_Success(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewUserService(repo)

	user, err := svc.Create("张三", "zhangsan@example.com")
	if err != nil {
		t.Fatalf("期望成功，实际报错: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("期望 ID 非零")
	}
	if user.Name != "张三" {
		t.Fatalf("期望姓名为张三，实际: %s", user.Name)
	}
}

// TestCreateUser_EmptyName 空姓名应返回业务错误
func TestCreateUser_EmptyName(t *testing.T) {
	repo := &mockUserRepo{}
	svc := NewUserService(repo)

	_, err := svc.Create("", "test@example.com")
	if err != ErrInvalidInput {
		t.Fatalf("期望 ErrInvalidInput，实际: %v", err)
	}
}

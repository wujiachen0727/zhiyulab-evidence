// Package handler 提供 HTTP 处理层
// [实测 Go 1.26.2]
package handler

import (
	"fmt"

	repo "example.com/testdata/repository" // 案例1: import alias
	"example.com/testdata/service"
)

// UserHandler HTTP 处理器
type UserHandler struct {
	svc      *service.UserService
	repoObj  *repo.UserRepo          // 直接持有 repository 对象——架构违规
	repoIntf repo.UserRepository     // 案例3: 通过接口变量持有——go/ast 检测不到
}

// NewUserHandler 构造函数
func NewUserHandler(svc *service.UserService, r *repo.UserRepo) *UserHandler {
	return &UserHandler{
		svc:      svc,
		repoObj:  r,
		repoIntf: r, // 接口变量指向 repository 实现
	}
}

// HandleGetUser_Compliant 合规: 通过 service 层调用
func (h *UserHandler) HandleGetUser_Compliant(id int) {
	name, err := h.svc.GetUser(id)
	if err != nil {
		fmt.Println("error:", err)
		return
	}
	fmt.Println("user:", name)
}

// HandleDeleteUser_ViolationDirect 违规案例1: 直接调用 repository（带 import alias）
// go/ast 只能看到 "repo.UserRepo"，如果 alias 换成别的名字就检测不到
func (h *UserHandler) HandleDeleteUser_ViolationDirect(id int) {
	r := &repo.UserRepo{}
	_ = r.DeleteByID(id) // handler → repository 直接调用
	fmt.Println("deleted user", id)
}

// HandleGetUser_ViolationField 违规案例2: 通过字段直接调用 repository 方法
func (h *UserHandler) HandleGetUser_ViolationField(id int) {
	name, _ := h.repoObj.FindByID(id) // handler → repository 直接调用
	fmt.Println("user:", name)
}

// HandleGetUser_ViolationInterface 违规案例3: 通过接口变量调用 repository 方法
// 这是 go/ast 的盲区——ast 只能看到 h.repoIntf.FindByID()，
// 无法知道 repoIntf 的类型来自 repository 包
func (h *UserHandler) HandleGetUser_ViolationInterface(id int) {
	name, _ := h.repoIntf.FindByID(id) // handler → repository（经接口）
	fmt.Println("user:", name)
}

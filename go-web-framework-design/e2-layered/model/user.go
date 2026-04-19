// [实测结构，代码逻辑为演示]
// E2: 领域模型 —— 和框架无关的纯数据结构
package model

// User 用户实体 —— 不依赖任何框架
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// [实测结构，代码逻辑为演示]
// E2: 数据访问层接口 —— 定义"做什么"，不关心"怎么做"
package repository

import "go-web-demo/model"

// UserRepository 用户数据访问接口
// 关键：这是一个接口，不是实现。
// service 层依赖这个接口，不依赖 MySQL/Redis/文件 等具体实现。
// 换数据源 = 写一个新 struct 实现这个接口，main.go 里换一行注入。
type UserRepository interface {
	List() ([]model.User, error)
	GetByID(id int64) (user model.User, err error)
	Create(user *model.User) (err error)
	Update(user *model.User) (err error)
	Delete(id int64) (err error)
}

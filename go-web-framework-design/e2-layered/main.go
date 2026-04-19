// [实测结构，代码逻辑为演示]
// E2: 分层版 Gin 项目 —— 入口文件
// 用途：展示依赖注入 + 接口分层的项目结构
package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"go-web-demo/handler"
	"go-web-demo/repository"
	"go-web-demo/service"
)

func main() {
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/myapp")
	if err != nil {
		log.Fatal(err)
	}

	// 依赖注入链：db → repository → service → handler
	// 每一层只知道自己的上游接口，不知道具体实现
	userRepo := repository.NewMySQLUserRepo(db)
	userSvc := service.NewUserService(userRepo)
	userHandler := handler.NewUserHandler(userSvc)

	r := gin.Default()

	// 路由注册 —— handler 只负责 HTTP 协议层
	users := r.Group("/users")
	{
		users.GET("", userHandler.List)
		users.GET("/:id", userHandler.Get)
		users.POST("", userHandler.Create)
		users.PUT("/:id", userHandler.Update)
		users.DELETE("/:id", userHandler.Delete)
	}

	r.Run(":8080")
}

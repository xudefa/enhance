// Package main 演示如何使用 enhance 框架构建完整的 REST API
//
// 该示例展示：
// - RESTful 路由设计
// - 请求验证
// - 统一错误处理
// - 中间件使用
// - 依赖注入
package main

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance/web/mvc"
	"github.com/xudefa/enhance/web/server"
)

// User 用户模型
type User struct {
	ID    int    `json:"id"`
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=0,max=150"`
}

// UserController 用户控制器
type UserController struct {
	users  []User
	nextID int
}

// NewUserController 创建用户控制器
func NewUserController() *UserController {
	return &UserController{
		users:  make([]User, 0),
		nextID: 1,
	}
}

// Routes 注册路由
func (c *UserController) Routes(router mvc.Router) {
	router.GET("/api/users", c.ListUsers)
	router.GET("/api/users/{id}", c.GetUser)
	router.POST("/api/users", c.CreateUser)
	router.PUT("/api/users/{id}", c.UpdateUser)
	router.DELETE("/api/users/{id}", c.DeleteUser)
}

// ListUsers 获取用户列表
func (c *UserController) ListUsers(ctx mvc.Context) {
	ctx.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"data":  c.users,
		"total": len(c.users),
	})
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx mvc.Context) {
	id := ctx.PathParam("id")

	for _, user := range c.users {
		if fmt.Sprintf("%d", user.ID) == id {
			ctx.JSON(http.StatusOK, map[string]any{
				"code": 0,
				"data": user,
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, map[string]any{
		"code":    404,
		"message": "用户不存在",
	})
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name  string `json:"name" validate:"required,min=2,max=50"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"min=0,max=150"`
}

// CreateUser 创建用户
func (c *UserController) CreateUser(ctx mvc.Context) {
	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	user := User{
		ID:    c.nextID,
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}
	c.nextID++
	c.users = append(c.users, user)

	ctx.JSON(http.StatusCreated, map[string]any{
		"code": 0,
		"data": user,
	})
}

// UpdateUser 更新用户
func (c *UserController) UpdateUser(ctx mvc.Context) {
	id := ctx.PathParam("id")

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	for i, user := range c.users {
		if fmt.Sprintf("%d", user.ID) == id {
			c.users[i].Name = req.Name
			c.users[i].Email = req.Email
			c.users[i].Age = req.Age

			ctx.JSON(http.StatusOK, map[string]any{
				"code": 0,
				"data": c.users[i],
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, map[string]any{
		"code":    404,
		"message": "用户不存在",
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx mvc.Context) {
	id := ctx.PathParam("id")

	for i, user := range c.users {
		if fmt.Sprintf("%d", user.ID) == id {
			c.users = append(c.users[:i], c.users[i+1:]...)
			ctx.JSON(http.StatusOK, map[string]any{
				"code":    0,
				"message": "删除成功",
			})
			return
		}
	}

	ctx.JSON(http.StatusNotFound, map[string]any{
		"code":    404,
		"message": "用户不存在",
	})
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(ctx mvc.Context) {
	fmt.Printf("[%s] %s\n", ctx.RequestMethod(), ctx.RequestURI())
	ctx.Next()
}

func main() {
	// 创建路由器
	router := server.NewRouter()

	// 注册中间件
	router.Use(LoggingMiddleware)

	// 创建并注册控制器
	userController := NewUserController()
	userController.Routes(router)

	// 创建服务器
	srv := server.NewHTTPServer(
		server.WithHost(":8080"),
		server.WithReadTimeout(30),
		server.WithWriteTimeout(30),
	)
	srv.SetHandler(router)

	fmt.Println("REST API 服务器启动: http://localhost:8080")
	fmt.Println("API 端点:")
	fmt.Println("  GET    /api/users         - 获取用户列表")
	fmt.Println("  GET    /api/users/{id}    - 获取用户详情")
	fmt.Println("  POST   /api/users         - 创建用户")
	fmt.Println("  PUT    /api/users/{id}    - 更新用户")
	fmt.Println("  DELETE /api/users/{id}    - 删除用户")

	if err := srv.Start(); err != nil {
		fmt.Printf("服务器启动失败: %v\n", err)
	}
}

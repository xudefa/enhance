package main

import (
	"net/http"

	"github.com/xudefa/enhance/web/mvc"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
	DB *gorm.DB
}

// NewUserController 创建用户控制器
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{DB: db}
}

// Routes 注册路由
// 使用 mvc.Router 接口而非具体实现，方便后续替换为 gin/hertz 等其他框架
func (c *UserController) Routes(router mvc.Router) {
	router.GET("/api/health", c.HealthCheck)
	router.GET("/api/users", c.ListUsers)
	router.GET("/api/users/{id}", c.GetUser)
	router.POST("/api/users", c.CreateUser)
	router.PUT("/api/users/{id}", c.UpdateUser)
	router.DELETE("/api/users/{id}", c.DeleteUser)
}

// HealthCheck 健康检查接口 - 验证数据库连接
func (c *UserController) HealthCheck(ctx mvc.Context) {
	// 查询数据库连接
	var count int64
	if err := c.DB.Table("users").Count(&count).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "数据库连接正常",
		"status":  "ok",
		"data": map[string]any{
			"user_count": count,
			"database":   "connected",
		},
	})
}

// ListUsers 获取用户列表
func (c *UserController) ListUsers(ctx mvc.Context) {
	var users []User
	if err := c.DB.Find(&users).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"data":  users,
		"total": len(users),
	})
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx mvc.Context) {
	id := ctx.PathParam("id")

	var user User
	if err := c.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": user,
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
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	if err := c.DB.Create(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

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

	var user User
	if err := c.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			ctx.JSON(http.StatusNotFound, map[string]any{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Age = req.Age

	if err := c.DB.Save(&user).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": user,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx mvc.Context) {
	id := ctx.PathParam("id")

	if err := c.DB.Delete(&User{}, id).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "删除成功",
	})
}

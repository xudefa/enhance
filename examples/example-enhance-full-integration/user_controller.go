package main

import (
	"net/http"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/mvc"
	"gorm.io/gorm"
)

// UserController 用户控制器
type UserController struct {
	DB     *gorm.DB
	Logger log.Logger
}

// NewUserController 创建用户控制器
func NewUserController(db *gorm.DB, logger log.Logger) *UserController {
	return &UserController{DB: db, Logger: logger}
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
	c.Logger.Info(ctx.Context(), "查询用户列表")

	var users []User
	if err := c.DB.Find(&users).Error; err != nil {
		c.Logger.Error(ctx.Context(), "查询用户失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	c.Logger.Info(ctx.Context(), "查询用户列表成功", log.KeyValue{Key: "count", Value: len(users)})
	ctx.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"data":  users,
		"total": len(users),
	})
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx mvc.Context) {
	id := ctx.PathParam("id")
	c.Logger.Info(ctx.Context(), "查询用户", log.KeyValue{Key: "id", Value: id})

	var user User
	if err := c.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Logger.Warn(ctx.Context(), "用户不存在", log.KeyValue{Key: "id", Value: id})
			ctx.JSON(http.StatusNotFound, map[string]any{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.Logger.Error(ctx.Context(), "查询用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
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
		c.Logger.Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	c.Logger.Info(ctx.Context(), "创建用户",
		log.KeyValue{Key: "name", Value: req.Name},
		log.KeyValue{Key: "email", Value: req.Email},
	)

	user := User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	if err := c.DB.Create(&user).Error; err != nil {
		c.Logger.Error(ctx.Context(), "创建用户失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

	c.Logger.Info(ctx.Context(), "用户创建成功", log.KeyValue{Key: "id", Value: user.ID})
	ctx.JSON(http.StatusCreated, map[string]any{
		"code": 0,
		"data": user,
	})
}

// UpdateUser 更新用户
func (c *UserController) UpdateUser(ctx mvc.Context) {
	id := ctx.PathParam("id")
	c.Logger.Info(ctx.Context(), "更新用户", log.KeyValue{Key: "id", Value: id})

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.Logger.Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	var user User
	if err := c.DB.First(&user, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.Logger.Warn(ctx.Context(), "用户不存在", log.KeyValue{Key: "id", Value: id})
			ctx.JSON(http.StatusNotFound, map[string]any{
				"code":    404,
				"message": "用户不存在",
			})
			return
		}
		c.Logger.Error(ctx.Context(), "查询用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
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
		c.Logger.Error(ctx.Context(), "更新用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	c.Logger.Info(ctx.Context(), "用户更新成功", log.KeyValue{Key: "id", Value: id})
	ctx.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": user,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx mvc.Context) {
	id := ctx.PathParam("id")
	c.Logger.Info(ctx.Context(), "删除用户", log.KeyValue{Key: "id", Value: id})

	if err := c.DB.Delete(&User{}, id).Error; err != nil {
		c.Logger.Error(ctx.Context(), "删除用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	c.Logger.Info(ctx.Context(), "用户删除成功", log.KeyValue{Key: "id", Value: id})
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "删除成功",
	})
}

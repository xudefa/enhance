package main

import (
	"net/http"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/web/mvc"
	"xorm.io/xorm"
)

// UserController 用户控制器
type UserController struct {
	Engine *xorm.Engine
	Logger log.Logger
}

// logger 获取日志记录器，如果未注入则使用默认实现
func (c *UserController) logger() log.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return log.NewSlogLogger()
}

// NewUserController 创建用户控制器
func NewUserController(engine *xorm.Engine, logger log.Logger) *UserController {
	return &UserController{Engine: engine, Logger: logger}
}

// Routes 注册路由
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
	c.logger().Info(ctx.Context(), "执行健康检查")

	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
			"status":  "error",
		})
		return
	}

	// 查询数据库连接
	count, err := c.Engine.Count(new(User))
	if err != nil {
		c.logger().Error(ctx.Context(), "数据库连接失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "数据库连接失败: " + err.Error(),
			"status":  "error",
		})
		return
	}

	c.logger().Info(ctx.Context(), "数据库连接正常", log.KeyValue{Key: "user_count", Value: count})
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
	c.logger().Info(ctx.Context(), "查询用户列表")

	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
		})
		return
	}

	var users []User
	if err := c.Engine.Find(&users); err != nil {
		c.logger().Error(ctx.Context(), "查询用户失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "查询用户列表成功", log.KeyValue{Key: "count", Value: len(users)})
	ctx.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"data":  users,
		"total": len(users),
	})
}

// GetUser 获取单个用户
func (c *UserController) GetUser(ctx mvc.Context) {
	id := ctx.PathParam("id")
	c.logger().Info(ctx.Context(), "查询用户", log.KeyValue{Key: "id", Value: id})

	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
		})
		return
	}

	var user User
	has, err := c.Engine.ID(id).Get(&user)
	if err != nil {
		c.logger().Error(ctx.Context(), "查询用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	if !has {
		c.logger().Warn(ctx.Context(), "用户不存在", log.KeyValue{Key: "id", Value: id})
		ctx.JSON(http.StatusNotFound, map[string]any{
			"code":    404,
			"message": "用户不存在",
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
	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
		})
		return
	}

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger().Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "创建用户",
		log.KeyValue{Key: "name", Value: req.Name},
		log.KeyValue{Key: "email", Value: req.Email},
	)

	user := User{
		Name:  req.Name,
		Email: req.Email,
		Age:   req.Age,
	}

	_, err := c.Engine.Insert(&user)
	if err != nil {
		c.logger().Error(ctx.Context(), "创建用户失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "创建用户失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "用户创建成功", log.KeyValue{Key: "id", Value: user.ID})
	ctx.JSON(http.StatusCreated, map[string]any{
		"code": 0,
		"data": user,
	})
}

// UpdateUser 更新用户
func (c *UserController) UpdateUser(ctx mvc.Context) {
	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
		})
		return
	}

	id := ctx.PathParam("id")
	c.logger().Info(ctx.Context(), "更新用户", log.KeyValue{Key: "id", Value: id})

	var req CreateUserRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger().Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	var user User
	has, err := c.Engine.ID(id).Get(&user)
	if err != nil {
		c.logger().Error(ctx.Context(), "查询用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询用户失败: " + err.Error(),
		})
		return
	}

	if !has {
		c.logger().Warn(ctx.Context(), "用户不存在", log.KeyValue{Key: "id", Value: id})
		ctx.JSON(http.StatusNotFound, map[string]any{
			"code":    404,
			"message": "用户不存在",
		})
		return
	}

	user.Name = req.Name
	user.Email = req.Email
	user.Age = req.Age

	_, err = c.Engine.ID(id).Update(&user)
	if err != nil {
		c.logger().Error(ctx.Context(), "更新用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "更新用户失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "用户更新成功", log.KeyValue{Key: "id", Value: id})
	ctx.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": user,
	})
}

// DeleteUser 删除用户
func (c *UserController) DeleteUser(ctx mvc.Context) {
	if c.Engine == nil {
		c.logger().Error(ctx.Context(), "XORM Engine 未初始化")
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "XORM Engine 未初始化",
		})
		return
	}

	id := ctx.PathParam("id")
	c.logger().Info(ctx.Context(), "删除用户", log.KeyValue{Key: "id", Value: id})

	_, err := c.Engine.ID(id).Delete(new(User))
	if err != nil {
		c.logger().Error(ctx.Context(), "删除用户失败",
			log.KeyValue{Key: "id", Value: id},
			log.KeyValue{Key: "error", Value: err.Error()},
		)
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "删除用户失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "用户删除成功", log.KeyValue{Key: "id", Value: id})
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "删除成功",
	})
}

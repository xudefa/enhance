// Package main 演示完整集成示例：zerolog + jwt + casbin + security + gorm
//
// 该示例展示了如何集成以下模块：
// - zerolog: 高性能结构化日志
// - jwt: JSON Web Token 认证
// - casbin: 基于角色的访问控制（RBAC）
// - security: 安全过滤器链
// - gorm: 数据库 ORM
package main

import (
	"fmt"
	"net/http"

	"github.com/xudefa/enhance"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/starter/jwt"
	"github.com/xudefa/enhance/web/mvc"

	// 导入 security 包以启用安全自动配置
	_ "github.com/xudefa/enhance/security"
	_ "github.com/xudefa/enhance/starter/casbin"

	// 导入 casbin-gorm 包以启用 Casbin GORM 自动配置（策略存储到数据库）
	_ "github.com/xudefa/enhance/starter/casbin-gorm"
	_ "github.com/xudefa/enhance/starter/gorm"

	// 导入 zerolog 包以启用 Zerolog 日志自动配置
	_ "github.com/xudefa/enhance/starter/zerolog"
)

// LoginRequest 登录请求。
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthController 认证控制器。
type AuthController struct {
	TokenProvider *jwt.DefaultTokenProvider
}

// Routes 注册路由。
func (c *AuthController) Routes(router mvc.Router) {
	router.POST("/login", c.Login)
}

// Login 登录接口。
func (c *AuthController) Login(ctx mvc.Context) {
	var req LoginRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if req.Username == "" || req.Password == "" {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "username and password required",
		})
		return
	}

	// 根据用户名分配角色
	var roles []string
	if req.Username == "alice" {
		roles = []string{"admin"}
	} else {
		roles = []string{"user"}
	}

	// 生成 Token
	token, err := c.TokenProvider.GenerateToken(ctx.Context(), req.Username, roles)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "failed to generate token",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "login success",
		"data": map[string]any{
			"token": token,
		},
	})
}

// ProfileController 用户资料控制器。
type ProfileController struct {
}

// Routes 注册路由。
func (c *ProfileController) Routes(router mvc.Router) {
	router.GET("/api/profile", c.GetProfile)
}

// GetProfile 获取用户资料。
func (c *ProfileController) GetProfile(ctx mvc.Context) {
	auth := security.GetAuthenticationFromContext(ctx.Context())
	if auth == nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{
			"code":    401,
			"message": "unauthorized",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "success",
		"data": map[string]any{
			"username":    auth.Principal(),
			"authorities": auth.Authorities(),
		},
	})
}

// AdminController 管理员控制器。
type AdminController struct {
}

// Routes 注册路由。
func (c *AdminController) Routes(router mvc.Router) {
	router.GET("/api/admin/users", c.GetUsers)
}

// GetUsers 获取用户列表（仅管理员）。
func (c *AdminController) GetUsers(ctx mvc.Context) {
	auth := security.GetAuthenticationFromContext(ctx.Context())
	if auth == nil {
		ctx.JSON(http.StatusUnauthorized, map[string]any{
			"code":    401,
			"message": "unauthorized",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "success",
		"data": map[string]any{
			"users":       []string{"alice", "bob"},
			"requestedBy": auth.Principal(),
		},
	})
}

func init() {
	// 注册控制器
	fmt.Println("[DEBUG] 开始注册控制器...")
	mvc.RegisterController(&AuthController{})
	mvc.RegisterController(&ProfileController{})
	mvc.RegisterController(&AdminController{})
	mvc.RegisterController(&UserController{})
	mvc.RegisterController(&CasbinController{})
	fmt.Println("[DEBUG] 控制器已注册，总数:", len(mvc.GetControllers()))
}

func main() {
	enhance.Run(
		boot.WithAppName("enhance-full-demo"),
		boot.WithVersion("1.0.0"),
	)
}

// Package main 演示 XORM + Casbin-XORM 集成示例
//
// 该示例展示了如何集成以下模块：
// - xorm: 数据库 ORM
// - casbin-xorm: 基于 XORM 的 Casbin 权限管理（策略存储到数据库）
// - jwt: JSON Web Token 认证
// - security: 安全过滤器链
package main

import (
	"net/http"

	"github.com/xudefa/enhance"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/starter/jwt"
	"github.com/xudefa/enhance/web/mvc"

	// 导入 security 包以启用安全自动配置
	_ "github.com/xudefa/enhance/security"
	// 导入 casbin-xorm 包以启用 Casbin XORM 自动配置（策略存储到数据库）
	_ "github.com/xudefa/enhance/starter/casbin-xorm"
	// 导入 xorm 包以启用 XORM 数据库自动配置
	_ "github.com/xudefa/enhance/starter/xorm"
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
			"username":    auth.Name(),
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
			"requestedBy": auth.Name(),
		},
	})
}

// CasbinPolicyController Casbin 策略管理控制器。
type CasbinPolicyController struct {
	Enforcer security.CasbinEnforcer
}

// Routes 注册路由。
func (c *CasbinPolicyController) Routes(router mvc.Router) {
	router.GET("/api/casbin/policies", c.GetPolicies)
	router.POST("/api/casbin/policy", c.AddPolicy)
	router.DELETE("/api/casbin/policy", c.RemovePolicy)
}

// GetPolicies 获取所有策略。
func (c *CasbinPolicyController) GetPolicies(ctx mvc.Context) {
	policies, err := c.Enforcer.GetPolicy(ctx.Context())
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "获取策略失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code": 0,
		"data": policies,
	})
}

// AddPolicyRequest 添加策略请求。
type AddPolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// AddPolicy 添加策略。
func (c *CasbinPolicyController) AddPolicy(ctx mvc.Context) {
	var req AddPolicyRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := c.Enforcer.AddPolicy(ctx.Context(), req.Subject, req.Object, req.Action); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "添加策略失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "策略添加成功",
	})
}

// RemovePolicyRequest 移除策略请求。
type RemovePolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// RemovePolicy 移除策略。
func (c *CasbinPolicyController) RemovePolicy(ctx mvc.Context) {
	var req RemovePolicyRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	if err := c.Enforcer.RemovePolicy(ctx.Context(), req.Subject, req.Object, req.Action); err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "移除策略失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "策略移除成功",
	})
}

func init() {
	// 注册控制器
	mvc.RegisterController(&AuthController{})
	mvc.RegisterController(&ProfileController{})
	mvc.RegisterController(&AdminController{})
	mvc.RegisterController(&UserController{})
	mvc.RegisterController(&CasbinPolicyController{})
}

func main() {
	enhance.Run(
		boot.WithAppName("enhance-xorm-casbin-demo"),
		boot.WithVersion("1.0.0"),
	)
}

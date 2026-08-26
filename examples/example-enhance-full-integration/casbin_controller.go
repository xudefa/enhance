package main

import (
	"net/http"

	"github.com/xudefa/enhance/log"
	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/web/mvc"
)

// CasbinController Casbin 策略管理控制器
type CasbinController struct {
	Enforcer security.CasbinEnforcer `inject:"true"`
	Logger   log.Logger              `inject:"true"`
}

func (c *CasbinController) logger() log.Logger {
	if c.Logger != nil {
		return c.Logger
	}
	return log.NewSlogLogger()
}

// NewCasbinController 创建 Casbin 控制器
func NewCasbinController(enforcer security.CasbinEnforcer, logger log.Logger) *CasbinController {
	return &CasbinController{Enforcer: enforcer, Logger: logger}
}

// Routes 注册路由
func (c *CasbinController) Routes(router mvc.Router) {
	router.GET("/api/casbin/policies", c.ListPolicies)
	router.POST("/api/casbin/policy", c.AddPolicy)
	router.DELETE("/api/casbin/policy", c.RemovePolicy)
	router.POST("/api/casbin/check", c.CheckPermission)
}

// AddPolicyRequest 添加策略请求
type AddPolicyRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// CheckPermissionRequest 权限检查请求
type CheckPermissionRequest struct {
	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

// ListPolicies 获取所有策略
func (c *CasbinController) ListPolicies(ctx mvc.Context) {
	c.logger().Info(ctx.Context(), "查询 Casbin 策略列表")

	policies, err := c.Enforcer.GetPolicy(ctx.Context())
	if err != nil {
		c.logger().Error(ctx.Context(), "查询策略失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "查询策略失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":  0,
		"data":  policies,
		"total": len(policies),
	})
}

// AddPolicy 添加策略
func (c *CasbinController) AddPolicy(ctx mvc.Context) {
	var req AddPolicyRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger().Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "添加 Casbin 策略",
		log.KeyValue{Key: "subject", Value: req.Subject},
		log.KeyValue{Key: "object", Value: req.Object},
		log.KeyValue{Key: "action", Value: req.Action},
	)

	if err := c.Enforcer.AddPolicy(ctx.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.logger().Error(ctx.Context(), "添加策略失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "添加策略失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "策略添加成功",
		log.KeyValue{Key: "subject", Value: req.Subject},
	)
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "策略添加成功",
	})
}

// RemovePolicy 移除策略
func (c *CasbinController) RemovePolicy(ctx mvc.Context) {
	var req AddPolicyRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger().Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "移除 Casbin 策略",
		log.KeyValue{Key: "subject", Value: req.Subject},
		log.KeyValue{Key: "object", Value: req.Object},
		log.KeyValue{Key: "action", Value: req.Action},
	)

	if err := c.Enforcer.RemovePolicy(ctx.Context(), req.Subject, req.Object, req.Action); err != nil {
		c.logger().Error(ctx.Context(), "移除策略失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "移除策略失败: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "策略移除成功")
	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"message": "策略移除成功",
	})
}

// CheckPermission 检查权限
func (c *CasbinController) CheckPermission(ctx mvc.Context) {
	var req CheckPermissionRequest
	if err := ctx.BindJSON(&req); err != nil {
		c.logger().Warn(ctx.Context(), "请求参数错误", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"code":    400,
			"message": "请求参数错误: " + err.Error(),
		})
		return
	}

	c.logger().Info(ctx.Context(), "检查权限",
		log.KeyValue{Key: "subject", Value: req.Subject},
		log.KeyValue{Key: "object", Value: req.Object},
		log.KeyValue{Key: "action", Value: req.Action},
	)

	allowed, err := c.Enforcer.Enforce(ctx.Context(), req.Subject, req.Object, req.Action)
	if err != nil {
		c.logger().Error(ctx.Context(), "权限检查失败", log.KeyValue{Key: "error", Value: err.Error()})
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"code":    500,
			"message": "权限检查失败: " + err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]any{
		"code":    0,
		"allowed": allowed,
		"data": map[string]any{
			"subject": req.Subject,
			"object":  req.Object,
			"action":  req.Action,
			"result":  allowed,
		},
	})
}

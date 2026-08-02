package security

import (
	"context"

	"github.com/xudefa/enhance/security/authorization"
)

// ==================== 配置键常量 ====================

const (
	// Casbin 配置
	CasbinEnabled          = "security.casbin.enabled"
	CasbinModelType        = "security.casbin.model-type"
	CasbinModelPath        = "security.casbin.model-path"
	CasbinModelText        = "security.casbin.model-text"
	CasbinPolicyType       = "security.casbin.policy-type"
	CasbinPolicyPath       = "security.casbin.policy-path"
	CasbinPolicyText       = "security.casbin.policy-text"
	CasbinAutoLoad         = "security.casbin.auto-load"
	CasbinAutoLoadInterval = "security.casbin.auto-load-interval"

	// casbin 字段常量
	CasbinLogFieldModel  = "model-path"
	CasbinLogFieldPolicy = "policy-path"
)

// ==================== 默认值常量 ====================

const (
	DefaultCasbinModelType        = "file"
	DefaultCasbinModelPath        = "config/casbin_model.conf"
	DefaultCasbinPolicyType       = "file"
	DefaultCasbinPolicyPath       = "config/casbin_policy.csv"
	DefaultCasbinAutoLoad         = false
	DefaultCasbinAutoLoadInterval = 5
)

// CasbinEnforcer Casbin 执行器接口。
type CasbinEnforcer interface {
	Enforce(ctx context.Context, subject, object, action string) (bool, error)
	AddPolicy(ctx context.Context, sub, obj, act string) error
	RemovePolicy(ctx context.Context, sub, obj, act string) error
	GetPolicy(ctx context.Context) ([][]string, error)
	LoadPolicy(ctx context.Context) error
	SavePolicy(ctx context.Context) error
}

// CasbinVoter Casbin 投票者实现。
type CasbinVoter struct {
	enforcer CasbinEnforcer
}

func NewCasbinVoter(enforcer CasbinEnforcer) *CasbinVoter {
	if enforcer == nil {
		panic("casbin: enforcer must not be nil")
	}
	return &CasbinVoter{
		enforcer: enforcer,
	}
}

// Vote 投票决定是否授予访问权限。
// resource 格式为 "METHOD:URI"（由 FilterSecurityInterceptor 生成）
func (v *CasbinVoter) Vote(ctx context.Context, authentication authorization.Authentication, resource string, attributes []string) int {
	if authentication == nil || !authentication.Authenticated() {
		return ACCESS_ABSTAIN
	}

	// 从 resource 解析 HTTP 方法和 URI（格式: "METHOD:URI"）
	method := ""
	uri := resource
	if idx := len(resource); idx > 0 {
		for i, c := range resource {
			if c == ':' {
				method = resource[:i]
				uri = resource[i+1:]
				break
			}
		}
	}

	// 如果无法解析出方法，使用全部 resource 作为 URI
	if method == "" {
		uri = resource
	}

	subject := extractPrincipalName(authentication)

	allowed, err := v.enforcer.Enforce(ctx, subject, uri, method)
	if err != nil {
		// 执行器返回错误时采用失败关闭策略，拒绝访问
		return ACCESS_DENIED
	}

	if allowed {
		return ACCESS_GRANTED
	}

	return ACCESS_DENIED
}

// Supports 是否支持该属性。
func (v *CasbinVoter) Supports(attribute string) bool {
	return true
}

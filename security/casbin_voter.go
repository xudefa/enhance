package security

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/security/authorization"
)

// ==================== 配置键常量 ====================

const (
	// Casbin 配置
	CasbinEnabled = "security.casbin.enabled"
	// CasbinModelType Casbin 模型类型配置键。
	CasbinModelType = "security.casbin.model-type"
	// CasbinModelPath Casbin 模型文件路径配置键。
	CasbinModelPath = "security.casbin.model-path"
	// CasbinModelText Casbin 模型文本配置键。
	CasbinModelText = "security.casbin.model-text"
	// CasbinPolicyType Casbin 策略类型配置键。
	CasbinPolicyType = "security.casbin.policy-type"
	// CasbinPolicyPath Casbin 策略文件路径配置键。
	CasbinPolicyPath = "security.casbin.policy-path"
	// CasbinPolicyText Casbin 策略文本配置键。
	CasbinPolicyText = "security.casbin.policy-text"
	// CasbinAutoLoad 是否自动加载策略的配置键。
	CasbinAutoLoad = "security.casbin.auto-load"
	// CasbinAutoLoadInterval 自动加载间隔（秒）配置键。
	CasbinAutoLoadInterval = "security.casbin.auto-load-interval"

	// casbin 字段常量
	CasbinLogFieldModel = "model-path"
	// CasbinLogFieldPolicy 策略路径日志字段名。
	CasbinLogFieldPolicy = "policy-path"
)

// ==================== 默认值常量 ====================

const (
	// DefaultCasbinModelType 默认 Casbin 模型类型。
	DefaultCasbinModelType = "file"
	// DefaultCasbinModelPath 默认 Casbin 模型文件路径。
	DefaultCasbinModelPath = "config/casbin_model.conf"
	// DefaultCasbinPolicyType 默认 Casbin 策略类型。
	DefaultCasbinPolicyType = "file"
	// DefaultCasbinPolicyPath 默认 Casbin 策略文件路径。
	DefaultCasbinPolicyPath = "config/casbin_policy.csv"
	// DefaultCasbinAutoLoad 默认是否自动加载策略。
	DefaultCasbinAutoLoad = false
	// DefaultCasbinAutoLoadInterval 默认自动加载间隔（秒）。
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

// NewCasbinVoter 创建 Casbin 投票者，enforcer 为空时返回错误。
func NewCasbinVoter(enforcer CasbinEnforcer) (*CasbinVoter, error) {
	if enforcer == nil {
		return nil, fmt.Errorf("casbin: enforcer must not be nil")
	}
	return &CasbinVoter{
		enforcer: enforcer,
	}, nil
}

// MustNewCasbinVoter 创建 Casbin 投票者，失败则 panic。
func MustNewCasbinVoter(enforcer CasbinEnforcer) *CasbinVoter {
	voter, err := NewCasbinVoter(enforcer)
	if err != nil {
		panic(err)
	}
	return voter
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

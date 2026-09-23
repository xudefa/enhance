package security

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/security/authorization"
)

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

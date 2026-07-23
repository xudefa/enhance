// Package authorization 提供授权功能支持，用于 enhance 框架。
//
// 该模块提供独立的授权抽象层，将授权逻辑与认证分离。
// 参考 Spring Security Authorization 模块的设计理念。
//
// # 架构设计
//
//   - AccessDecisionManager: 访问决策管理器接口，基于投票结果做最终决策
//   - AccessDecisionVoter: 访问决策投票者接口，对访问请求进行投票
//   - AuthorizeRequests: 授权请求配置接口，注册 URL 匹配规则
//   - ExpressionInterceptUrlRegistry: 表达式拦截 URL 注册表接口
//   - UrlAuthorizationRule: URL 授权规则结构体
//
// # 核心功能
//
//   - 投票决策: 支持多种访问决策策略（肯定优先、一致通过、共识优先）
//   - URL 授权: 支持基于表达式的 URL 访问控制规则配置
//   - 权限检查: 支持角色和权限的细粒度检查
//
// # 使用方式
//
// 创建访问决策管理器：
//
//	voter := authorization.NewWebExpressionVoter()
//	manager := authorization.NewAffirmativeBased(voter)
//	err := manager.Decide(ctx, auth, "/api/users", []string{"hasRole('ADMIN')"})
//
// 配置 URL 授权规则：
//
//	registry := authorization.NewExpressionBasedUrlRegistry()
//	registry.RequestMatchers("/api/public/**").
//	    PermitAll().
//	    And().
//	    RequestMatchers("/api/admin/**").
//	    HasRole("ADMIN").
//	    And()
//	rules := registry.Get()
package authorization

import (
	"context"
	"errors"
)

// 错误定义。
var (
	// ErrAccessDenied 访问被拒绝。
	ErrAccessDenied = errors.New("access denied")
)

// 投票结果常量。
const (
	// AccessGranted 允许访问。
	AccessGranted = 1
	// AccessAbstain 投票弃权。
	AccessAbstain = 0
	// AccessDenied 拒绝访问。
	AccessDenied = -1
)

// Authentication 认证信息接口。
//
// 代表一个已认证主体的信息，包括身份标识、凭证和权限。
// 用于授权决策时获取当前用户的身份和权限信息。
type Authentication interface {
	// Principal 获取主体标识，通常是用户名或用户对象。
	Principal() any
	// Credentials 获取凭证信息。
	Credentials() any
	// Authorities 返回权限列表。
	Authorities() []string
	// Authenticated 是否已认证。
	Authenticated() bool
}

// AccessDecisionManager 访问决策管理器接口。
//
// 决定是否允许访问受保护的资源。
// 基于多个投票者的投票结果做最终决策。
type AccessDecisionManager interface {
	// Decide 决策是否允许访问。
	Decide(ctx context.Context, authentication Authentication, resource string, attributes []string) error
	// Supports 是否支持该决策属性。
	Supports(attribute string) bool
}

// AccessDecisionVoter 访问决策投票者接口。
//
// 对访问请求进行投票，返回允许、拒绝或弃权。
type AccessDecisionVoter interface {
	// Vote 投票决定是否允许访问。
	Vote(ctx context.Context, authentication Authentication, resource string, attributes []string) int
	// Supports 是否支持该属性。
	Supports(attribute string) bool
}

// AuthorizeRequests 授权请求配置接口。
//
// 配置 URL 路径的访问规则，支持链式调用。
type AuthorizeRequests interface {
	// RequestMatchers 注册 URL 匹配规则。
	RequestMatchers(patterns ...string) UrlAuthorizationRuleBuilder
	// AnyRequest 配置所有请求。
	AnyRequest() UrlAuthorizationRuleBuilder
}

// UrlAuthorizationRuleBuilder URL 授权规则构建器接口。
//
// 配置特定 URL 的访问控制规则，支持链式调用。
type UrlAuthorizationRuleBuilder interface {
	// HasAnyAuthority 要求指定权限（传入单个或多个）。
	HasAnyAuthority(authorities ...string) UrlAuthorizationRuleBuilder
	// HasRole 要求特定角色。
	HasRole(role string) UrlAuthorizationRuleBuilder
	// HasAnyRole 要求任意角色。
	HasAnyRole(roles ...string) UrlAuthorizationRuleBuilder
	// PermitAll 允许所有。
	PermitAll() UrlAuthorizationRuleBuilder
	// DenyAll 拒绝所有。
	DenyAll() UrlAuthorizationRuleBuilder
	// Authenticated 要求认证。
	Authenticated() UrlAuthorizationRuleBuilder
	// And 添加下一个规则。
	And() UrlAuthorizationRuleBuilder
}

// UrlAuthorizationRuleRegistry URL 授权规则注册表接口。
//
// 用于获取已构建的授权规则列表。
type UrlAuthorizationRuleRegistry interface {
	// Get 获取所有规则。
	Get() []UrlAuthorizationRule
}

// ExpressionInterceptUrlRegistry 表达式拦截 URL 注册表接口。
//
// 组合规则构建和规则获取功能。
type ExpressionInterceptUrlRegistry interface {
	UrlAuthorizationRuleBuilder
	UrlAuthorizationRuleRegistry
}

// UrlAuthorizationRule URL 授权规则。
//
// 定义一组 URL 模式及其对应的访问控制属性。
type UrlAuthorizationRule struct {
	// Patterns URL 模式列表。
	Patterns []string
	// Attributes 访问控制属性列表。
	Attributes []string
}

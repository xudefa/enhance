// Package security 提供安全功能支持，用于 enhance 框架。
//
// 该模块提供认证、授权、Web 安全过滤器、速率限制等安全机制，保护应用安全。
// 参考 Spring Security 的设计理念。
//
// # 架构设计
//
//   - SecurityContext: 安全上下文接口，管理认证信息
//   - Authentication: 认证信息接口（别名 → authentication.Authentication）
//   - AuthenticationManager: 认证管理器接口（别名 → authentication.AuthenticationManager）
//   - AccessDecisionManager: 访问决策管理器接口（别名 → authorization.AccessDecisionManager）
//   - UserDetailsService: 用户详情服务接口
//   - UserDetails: 用户详情接口
//   - PasswordEncoder: 密码编码器接口
//   - SecurityFilter: 安全过滤器接口（别名 → filter.Filter）
//   - SecurityFilterChain: 安全过滤器链接口（别名 → filter.SecurityFilterChain）
//   - AccessDeniedHandler: 访问拒绝处理器接口
//   - AuthenticationEntryPoint: 认证入口点接口
//   - SecurityMetadataSource: 安全元数据源接口
//   - SecurityRequest: 安全请求接口
//   - SecurityResponse: 安全响应接口
//   - GrantedAuthority: 授予权限接口
//   - HttpSecurity: HTTP 安全配置接口
//   - AuthorizeRequests: 授权请求配置接口
//   - ExpressionInterceptUrlRegistry: URL 拦截注册接口
//   - SecurityConfig: 安全配置接口
//   - LogoutSuccessHandler: 登出成功处理器接口
//   - RateLimiter: 速率限制器接口
//   - CsrfTokenManager: CSRF 令牌管理器接口
//
// # 核心功能
//
//   - 认证: 支持用户名/密码、JWT、OAuth2 等认证方式
//   - 授权: 支持基于角色和权限的授权检查
//   - 安全过滤器: 支持请求拦截和安全检查
//   - 速率限制: 防止恶意请求和暴力破解
//   - CSRF 防护: 防止跨站请求伪造攻击
//
// # 使用方式
//
// 配置安全策略：
//
//	httpSec := security.NewHttpSecurity()
//	httpSec.FormLogin("/login").
//	    Logout("/logout").
//	    Csrf().
//	    AuthorizeRequests(func(authz security.AuthorizeRequests) {
//	        authz.AntMatchers("/api/**").HasRole("ROLE_API")
//	        authz.AnyRequest().Authenticated()
//	    })
//
// 认证用户：
//
//	auth := security.Authenticate(username, password)
//	if auth != nil {
//	    // 认证成功
//	}
//
// 检查权限：
//
//	if security.HasPermission("user:read") {
//	    // 有权限访问
//	}
//
// # 安全过滤器链
//
// 请求经过的过滤器链：
//
//   - AuthenticationFilter: 认证过滤器
//   - AuthorizationFilter: 授权过滤器
//   - RateLimitFilter: 速率限制过滤器
//   - CSRFTokenFilter: CSRF 防护过滤器
package security

import (
	"context"
	"errors"

	"github.com/xudefa/enhance/security/authentication"
	"github.com/xudefa/enhance/security/authorization"
	"github.com/xudefa/enhance/security/filter"
)

// 错误定义。
var (
	// ErrAuthenticationFailed 认证失败。
	ErrAuthenticationFailed = errors.New("authentication failed")
	// ErrAccessDenied 访问被拒绝。
	ErrAccessDenied = errors.New("access denied")
	// ErrUserNotFound 用户未找到。
	ErrUserNotFound = errors.New("user not found")
	// ErrBadCredentials 凭证无效。
	ErrBadCredentials = errors.New("bad credentials")
)

// ==================== 子包类型别名 ====================
//
// 以下类型已迁移到子包中，此处通过类型别名保持包级访问。

type (
	// authentication 包的类型别名
	Authentication         = authentication.Authentication
	AuthenticationToken    = authentication.AuthenticationToken
	AuthenticationManager  = authentication.AuthenticationManager
	AuthenticationProvider = authentication.AuthenticationProvider
	UserDetails            = authentication.UserDetails
	UserDetailsService     = authentication.UserDetailsService
	PasswordEncoder        = authentication.PasswordEncoder

	// authorization 包的类型别名
	AccessDecisionManager = authorization.AccessDecisionManager
	AccessDecisionVoter   = authorization.AccessDecisionVoter

	// filter 包的类型别名
	SecurityFilter      = filter.Filter
	SecurityFilterChain = filter.SecurityFilterChain
)

// SecurityContext 安全上下文接口。
//
// 提供对当前认证信息的访问和管理。
// 通常与请求生命周期绑定，用于存储和检索认证主体信息。
//
// 使用示例：
//
//	ctx := security.NewSecurityContext()
//	ctx.SetAuthentication(auth)
//	auth := ctx.Authentication()
type SecurityContext interface {
	// Authentication 获取当前认证信息。
	Authentication() Authentication
	// SetAuthentication 设置认证信息。
	SetAuthentication(auth Authentication)
	// ClearAuthentication 清除认证信息。
	ClearAuthentication()
}

// AccessDeniedHandler 访问拒绝处理器接口。
//
// 处理访问被拒绝的情况，如权限不足。
// 可以自定义响应格式或重定向到错误页面。
type AccessDeniedHandler interface {
	// Handle 处理访问拒绝。
	Handle(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error
}

// AuthenticationEntryPoint 认证入口点接口。
//
// 处理未认证用户访问受保护资源的情况。
// 通常返回 401 状态码或重定向到登录页面。
type AuthenticationEntryPoint interface {
	// Commence 开始认证流程。
	Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error
}

// SecurityMetadataSource 安全元数据源接口。
//
// 提供访问资源所需的安全属性（如角色要求）。
// 用于动态访问控制决策。
type SecurityMetadataSource interface {
	// GetAttributes 获取安全属性。
	GetAttributes(ctx context.Context, request SecurityRequest) ([]string, error)
}

// SecurityRequest 安全请求接口。
//
// 抽象 HTTP 请求，提供安全相关的访问方法。
// 用于获取请求方法、URI、头部等信息。
type SecurityRequest interface {
	// GetMethod 获取请求方法。
	GetMethod() string
	// GetURI 获取请求 URI。
	GetURI() string
	// GetHeader 获取请求头。
	GetHeader(key string) string
	// RemoteAddress 获取直连对端的地址，格式为 "host:port"。
	RemoteAddress() string
	// SetAttribute 设置请求属性。
	SetAttribute(key string, value any)
	// GetAttribute 获取请求属性。
	GetAttribute(key string) (any, bool)
}

// SecurityResponse 安全响应接口。
//
// 抽象 HTTP 响应，提供安全相关的修改方法。
// 用于设置状态码、头部和响应体。
type SecurityResponse interface {
	// SetStatusCode 设置状态码。
	SetStatusCode(code int)
	// SetHeader 设置响应头。
	SetHeader(key, value string)
	// Write 写入响应数据。
	Write(data []byte) error
}

// GrantedAuthority 授予权限接口。
//
// 代表一个具体的权限，如 "ROLE_ADMIN"、"ROLE_USER"。
type GrantedAuthority interface {
	// Authority 返回权限字符串。
	Authority() string
}

// HttpSecurity HTTP 安全配置接口。
//
// 提供链式 API 配置 HTTP 安全规则。
// 支持认证管理器、用户详情服务、过滤器、CSRF、登出等配置。
//
// 注意：此接口有 16 个方法，违反小接口原则。
// 未来版本将拆分为多个独立的 Configurer 接口（CsrfConfigurer、CorsConfigurer 等）。
//
// 使用示例：
//
//	httpSec := security.NewHttpSecurity()
//	httpSec.FormLogin("/login").
//	    Logout("/logout").
//	    Csrf().
//	    AuthorizeRequests(func(authz security.AuthorizeRequests) {
//	        authz.AntMatchers("/api/**").HasRole("ROLE_API")
//	        authz.AnyRequest().Authenticated()
//	    })
type HttpSecurity interface {
	// AuthenticationManager 设置认证管理器。
	AuthenticationManager(authManager AuthenticationManager) HttpSecurity
	// UserDetailsService 设置用户详情服务。
	UserDetailsService(userDetailsService UserDetailsService) HttpSecurity
	// PasswordEncoder 设置密码编码器。
	PasswordEncoder(encoder PasswordEncoder) HttpSecurity
	// AccessDecisionManager 设置访问决策管理器。
	AccessDecisionManager(manager AccessDecisionManager) HttpSecurity
	// SecurityMetadataSource 设置安全元数据源。
	SecurityMetadataSource(source SecurityMetadataSource) HttpSecurity
	// AuthorizeRequests 配置授权规则。
	AuthorizeRequests(config func(authorizer AuthorizeRequests)) HttpSecurity
	// AddFilter 添加过滤器。
	AddFilter(filter SecurityFilter) HttpSecurity
	// AddFilterBefore 在指定过滤器之前添加过滤器。
	AddFilterBefore(filter SecurityFilter, beforeFilter SecurityFilter) HttpSecurity
	// AddFilterAfter 在指定过滤器之后添加过滤器。
	AddFilterAfter(filter SecurityFilter, afterFilter SecurityFilter) HttpSecurity
	// Anonymous 启用匿名访问。
	Anonymous() HttpSecurity
	// ExceptionHandling 配置异常处理。
	ExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) HttpSecurity
	// Csrf 启用 CSRF 防护。
	Csrf() HttpSecurity
	// Logout 配置登出。
	Logout(logoutUrl string, successHandler ...LogoutSuccessHandler) HttpSecurity
	// FormLogin 配置表单登录。
	FormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) HttpSecurity
	// HttpBasic 启用 HTTP Basic 认证。
	HttpBasic() HttpSecurity
	// Build 构建安全过滤器链。
	Build() (SecurityFilterChain, error)
}

// AuthorizeRequests 授权请求配置接口。
//
// 配置 URL 路径的访问规则。
type AuthorizeRequests interface {
	// AntMatchers 配置 Ant 风格的路径匹配器。
	AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry
	// AnyRequest 配置所有请求。
	AnyRequest() ExpressionInterceptUrlRegistry
}

// ExpressionInterceptUrlRegistry URL 拦截注册接口。
//
// 配置特定 URL 的访问表达式。
type ExpressionInterceptUrlRegistry interface {
	// PermitAll 允许所有访问。
	PermitAll() HttpSecurity
	// Authenticated 需要认证。
	Authenticated() HttpSecurity
	// HasRole 需要指定角色。
	HasRole(role string) HttpSecurity
	// HasAnyRole 需要指定角色之一。
	HasAnyRole(roles ...string) HttpSecurity
	// HasAuthority 需要指定权限。
	HasAuthority(authority string) HttpSecurity
	// HasAnyAuthority 需要指定权限之一。
	HasAnyAuthority(authorities ...string) HttpSecurity
	// DenyAll 拒绝所有访问。
	DenyAll() HttpSecurity
}

// SecurityConfig 安全配置接口。
//
// 提供自定义安全配置的入口。
type SecurityConfig interface {
	// Configure 配置 HTTP 安全。
	Configure(http HttpSecurity) error
}

// LogoutHandler 登出处理器接口。
type LogoutHandler interface {
	// Logout 处理登出。
	Logout(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication)
}

// LogoutSuccessHandler 登出成功处理器接口。
type LogoutSuccessHandler interface {
	// OnLogoutSuccess 处理登出成功。
	OnLogoutSuccess(ctx context.Context, request SecurityRequest, response SecurityResponse, authentication Authentication)
}

// RateLimiter 限流器接口。
//
// 提供请求频率限制功能，防止恶意请求和暴力攻击。
type RateLimiter interface {
	// Allow 判断是否允许请求。
	Allow(key string) bool
	// Cleanup 清理过期数据。
	Cleanup()
}

// RateLimitStrategy 限流策略接口。
type RateLimitStrategy interface {
	// Allow 判断是否允许请求。
	Allow(key string) bool
}

// CsrfTokenRepository CSRF 令牌仓库接口。
//
// 管理 CSRF 令牌的生成、验证、保存和清除。
type CsrfTokenRepository interface {
	// GenerateToken 生成新的 CSRF 令牌。
	GenerateToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error)
	// LoadToken 加载当前会话已存在的 CSRF 令牌，会话中不存在时返回 nil。
	LoadToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error)
	// ValidateToken 验证 CSRF 令牌。
	ValidateToken(ctx context.Context, request SecurityRequest, token string) bool
	// SaveToken 保存 CSRF 令牌到响应。
	SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken)
	// ClearToken 清除 CSRF 令牌。
	ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse)
}

// CsrfToken CSRF 令牌结构。
type CsrfToken struct {
	Identifier string // 令牌标识符
	Value      string // 令牌值
}

// ==================== 配置键常量 ====================

const (
	// Security 配置
	SecurityEnabled = "security.enabled"
)

// ==================== 条件值常量 ====================

const (
	ConditionTrue = "true"
)

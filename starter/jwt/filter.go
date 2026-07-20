package jwt

import (
	"context"
	"fmt"
	"strings"

	"github.com/xudefa/enhance/security"
)

// JwtAuthenticationFilter JWT 认证过滤器。
//
// 实现 security.SecurityFilter 接口，从 HTTP 请求的 Authorization 头中提取 Bearer Token，
// 验证 Token 有效性并设置安全上下文。
//
// 工作流程：
//  1. 检查请求路径是否在排除列表中（如 /login, /public/**）
//  2. 从 Authorization 头提取 Bearer Token
//  3. 使用 TokenProvider 解析和验证 Token
//  4. 可选：通过 UserDetailsService 加载完整的用户详情
//  5. 创建 Authentication 对象并设置到 SecurityContext
//
// 注意：该过滤器只处理携带 Bearer Token 的请求，
// 对于没有 Token 的请求，会直接传递给下一个过滤器（由后续的 AnonymousAuthenticationFilter 处理）。
type JwtAuthenticationFilter struct {
	tokenProvider      TokenProvider               // Token 提供者，负责解析和验证 JWT
	excludePaths       []string                    // 不需要 JWT 认证的路径列表
	userDetailsService security.UserDetailsService // 用户详情服务（可选）
}

// NewJwtAuthenticationFilter 创建 JWT 认证过滤器。
func NewJwtAuthenticationFilter(tokenProvider TokenProvider, opts ...JwtFilterOption) *JwtAuthenticationFilter {
	f := &JwtAuthenticationFilter{
		tokenProvider: tokenProvider,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

// JwtFilterOption JWT 过滤器选项函数类型。
type JwtFilterOption func(*JwtAuthenticationFilter)

// WithExcludePaths 设置排除路径。
func WithExcludePaths(paths ...string) JwtFilterOption {
	return func(f *JwtAuthenticationFilter) {
		f.excludePaths = paths
	}
}

// WithUserDetailsService 设置用户详情服务。
func WithUserDetailsService(service security.UserDetailsService) JwtFilterOption {
	return func(f *JwtAuthenticationFilter) {
		f.userDetailsService = service
	}
}

// DoFilter 执行 JWT 认证过滤器。
//
// 该方法实现 security.SecurityFilter 接口，是过滤器链的核心执行逻辑。
//
// 执行流程：
//  1. 检查请求路径是否在排除列表中，如果是则跳过认证
//  2. 从 Authorization 头提取 Bearer Token（格式：Bearer <token>）
//  3. 使用 TokenProvider 解析和验证 Token
//  4. 如果配置了 UserDetailsService，则加载完整的用户详情
//  5. 创建 Authentication 对象并设置到 SecurityContext
//  6. 继续执行过滤器链
//
// 注意：
//   - 对于没有 Authorization 头的请求，直接跳过（由后续过滤器处理）
//   - 对于 Token 无效的请求，返回错误（不会被后续过滤器处理）
//   - 认证成功后，用户信息可通过 security.GetAuthentication() 获取
func (f *JwtAuthenticationFilter) DoFilter(ctx context.Context, request security.SecurityRequest, response security.SecurityResponse, chain security.SecurityFilterChain) error {
	// 检查是否在排除路径中
	uri := request.GetURI()
	if f.isExcluded(uri) {
		return chain.DoFilter(ctx, request, response)
	}

	// 从 Authorization 头提取 Token
	authHeader := request.GetHeader(HeaderAuthorization)
	if authHeader == "" {
		return chain.DoFilter(ctx, request, response)
	}

	token := extractBearerToken(authHeader)
	if token == "" {
		return chain.DoFilter(ctx, request, response)
	}

	// 验证并解析 Token
	claims, err := f.tokenProvider.ParseToken(ctx, token)
	if err != nil {
		// Token 无效时，设置 401 响应并返回错误以停止过滤器链
		response.SetStatusCode(401)
		response.SetHeader("Content-Type", "application/json")
		errorMsg := fmt.Sprintf(`{"code":401,"message":"jwt authentication failed: %v"}`, err)
		if writeErr := response.Write([]byte(errorMsg)); writeErr != nil {
			return fmt.Errorf("jwt authentication failed: %w", err)
		}
		return fmt.Errorf("invalid jwt token: %w", err)
	}

	// 加载用户详情（可选）
	var userDetails security.UserDetails
	if f.userDetailsService != nil {
		userDetails, err = f.userDetailsService.LoadUserByUsername(ctx, claims.Subject)
		if err != nil {
			// UserDetailsService 找不到用户时，使用 Token 中的信息
			userDetails = NewSimpleUserDetails(claims.Subject, claims.Authorities)
		}
	} else {
		// 使用 Token 中的信息创建用户详情
		userDetails = NewSimpleUserDetails(claims.Subject, claims.Authorities)
	}

	// 创建已认证的 Authentication
	auth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(userDetails, userDetails.Authorities())

	// 设置到安全上下文
	security.SetAuthentication(auth)
	ctx = security.ContextWithAuthentication(ctx, auth)

	return chain.DoFilter(ctx, request, response)
}

// IsJwtFilter 标识这是一个 JWT 认证过滤器。
//
// 该方法用于 security 模块自动检测和集成 JWT 过滤器。
// 返回 true 表示此过滤器是 JWT 认证过滤器，security 模块可以据此进行特殊处理。
func (f *JwtAuthenticationFilter) IsJwtFilter() bool {
	return true
}

// isExcluded 检查 URI 是否在排除路径中。
func (f *JwtAuthenticationFilter) isExcluded(uri string) bool {
	for _, path := range f.excludePaths {
		if uri == path || matchPattern(path, uri) {
			return true
		}
	}
	return false
}

// extractBearerToken 从 Authorization 头提取 Bearer Token。
func extractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, HeaderBearerPrefix) {
		return ""
	}
	return strings.TrimPrefix(authHeader, HeaderBearerPrefix)
}

// matchPattern 简单的路径匹配（支持 * 通配符）。
func matchPattern(pattern, uri string) bool {
	if pattern == "/**" {
		return true
	}
	if prefix, found := strings.CutSuffix(pattern, "/**"); found {
		return strings.HasPrefix(uri, prefix)
	}
	if prefix, found := strings.CutSuffix(pattern, "/*"); found {
		if remaining, ok := strings.CutPrefix(uri, prefix); ok {
			return len(remaining) > 0 && remaining[0] == '/' && !strings.Contains(remaining[1:], "/")
		}
		return false
	}
	return pattern == uri
}

// SimpleUserDetails 简单用户详情实现。
type SimpleUserDetails struct {
	username    string
	authorities []string
}

// NewSimpleUserDetails 创建简单用户详情。
func NewSimpleUserDetails(username string, authorities []string) *SimpleUserDetails {
	return &SimpleUserDetails{
		username:    username,
		authorities: authorities,
	}
}

func (u *SimpleUserDetails) Username() string {
	return u.username
}

func (u *SimpleUserDetails) Password() string {
	return ""
}

func (u *SimpleUserDetails) Authorities() []string {
	return u.authorities
}

func (u *SimpleUserDetails) Enabled() bool {
	return true
}

func (u *SimpleUserDetails) AccountNonExpired() bool {
	return true
}

func (u *SimpleUserDetails) CredentialsNonExpired() bool {
	return true
}

func (u *SimpleUserDetails) AccountNonLocked() bool {
	return true
}

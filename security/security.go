// Package security 提供安全功能支持，用于 enhance 框架。
// 设计灵感来源于 Spring Security，采用接口设计模式，支持灵活的安全配置。
//
// 核心功能：
// - 认证（Authentication）：验证用户身份
// - 授权（Authorization）：控制资源访问权限
// - 安全上下文（SecurityContext）：线程安全的认证信息存储
// - 过滤器链（FilterChain）：可配置的安全过滤器执行链
package security

import (
	"context"
	"sync"
	"sync/atomic"
)

// contextKey 定义上下文键类型。
type requestContextKey string

const authContextKey requestContextKey = "security.authentication"

// globalSecurityContext 全局安全上下文实例。
// 用于存储当前请求的认证信息，线程安全。
var globalSecurityContext atomic.Pointer[securityContext]

func init() {
	globalSecurityContext.Store(&securityContext{})
}

// GetAuthenticationFromContext 从 context.Context 获取认证信息。
func GetAuthenticationFromContext(ctx context.Context) Authentication {
	if val, ok := ctx.Value(authContextKey).(Authentication); ok {
		return val
	}
	return nil
}

// ContextWithAuthentication 将认证信息存入 context.Context。
func ContextWithAuthentication(ctx context.Context, auth Authentication) context.Context {
	return context.WithValue(ctx, authContextKey, auth)
}

// GetSecurityContext 获取全局安全上下文。
func GetSecurityContext() SecurityContext {
	return globalSecurityContext.Load()
}

// SetSecurityContext 设置全局安全上下文。
func SetSecurityContext(ctx SecurityContext) {
	if sc, ok := ctx.(*securityContext); ok {
		auth := sc.Authentication()
		globalSecurityContext.Load().SetAuthentication(auth)
	}
}

// GetAuthentication 获取当前认证信息。
//
// 优先从请求上下文获取，如果不存在则从全局安全上下文获取。
func GetAuthentication() Authentication {
	// 注意：这个函数无法访问请求上下文，只能从全局上下文获取
	// 在 HTTP 请求处理中，应该使用 GetAuthenticationFromContext(ctx.Context())
	return globalSecurityContext.Load().Authentication()
}

// SetAuthentication 设置当前认证信息到全局安全上下文。
func SetAuthentication(auth Authentication) {
	globalSecurityContext.Load().SetAuthentication(auth)
}

// ClearAuthentication 清除当前全局安全上下文中的认证信息。
func ClearAuthentication() {
	globalSecurityContext.Load().ClearAuthentication()
}

// InitSecurityContext 初始化安全上下文。
//
// 用于测试环境重置安全状态。
func InitSecurityContext() {
	globalSecurityContext.Store(&securityContext{})
}

// NewSecurityContext 创建新的安全上下文。
func NewSecurityContext() SecurityContext {
	return &securityContext{}
}

// securityContext 安全上下文实现。
//
// 线程安全的认证信息存储。
type securityContext struct {
	authentication Authentication
	mu             sync.RWMutex
}

func (s *securityContext) Authentication() Authentication {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.authentication
}

func (s *securityContext) SetAuthentication(auth Authentication) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authentication = auth
}

func (s *securityContext) ClearAuthentication() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.authentication = nil
}

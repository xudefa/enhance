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
)

// contextKey 定义上下文键类型。
type requestContextKey string

const authContextKey requestContextKey = "security.authentication"

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

// NewSecurityContext 创建新的安全上下文。
func NewSecurityContext() SecurityContext {
	return &securityContext{}
}

// securityContext 安全上下文实现。
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

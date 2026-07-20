// Package security 提供安全功能支持，用于 enhance 框架。
package security

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// CsrfFilter CSRF防护过滤器
// 职责：防止跨站请求伪造攻击
// 执行流程：
// 1. GET/HEAD/OPTIONS/TRACE请求：生成并保存CSRF Token到Cookie
// 2. POST/PUT/DELETE等请求：验证请求头中的CSRF Token
// 3. Token验证失败则拒绝请求
type CsrfFilter struct {
	tokenRepository CsrfTokenRepository
	excludePaths    []string
}

// NewCsrfFilter 创建 CSRF 过滤器。
func NewCsrfFilter(tokenRepository CsrfTokenRepository) *CsrfFilter {
	return &CsrfFilter{
		tokenRepository: tokenRepository,
		excludePaths:    []string{},
	}
}

// AddExcludePath 添加排除路径。
func (f *CsrfFilter) AddExcludePath(paths ...string) error {
	for _, path := range paths {
		if path == "" {
			return fmt.Errorf("exclude path cannot be empty")
		}
		if !strings.HasPrefix(path, "/") {
			return fmt.Errorf("exclude path must start with '/': %s", path)
		}
	}
	f.excludePaths = append(f.excludePaths, paths...)
	return nil
}

// DoFilter 执行 CSRF 检查。
func (f *CsrfFilter) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain SecurityFilterChain) error {
	uri := request.GetURI()

	for _, path := range f.excludePaths {
		if strings.HasPrefix(uri, path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	method := request.GetMethod()

	if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
		token, err := f.tokenRepository.GenerateToken(ctx, request)
		if err == nil && token != nil {
			f.tokenRepository.SaveToken(ctx, request, response, token)
			request.SetAttribute("csrf.token", token.Value)
		}
		return chain.DoFilter(ctx, request, response)
	}

	token := request.GetHeader("X-CSRF-Token")
	if token == "" {
		token = request.GetHeader("X-XSRF-Token")
	}
	if token == "" {
		return fmt.Errorf("missing CSRF token")
	}

	if !f.tokenRepository.ValidateToken(ctx, request, token) {
		return fmt.Errorf("invalid CSRF token")
	}

	return chain.DoFilter(ctx, request, response)
}

// CookieCsrfTokenRepository 基于Cookie的CSRF令牌仓库
// 职责：生成、保存和验证CSRF Token
// 存储策略：将Token保存到Cookie中，客户端需在请求头中携带Token
type CookieCsrfTokenRepository struct {
	cookieName     string
	headerName     string
	cookieHttpOnly bool
	secure         bool
	sameSite       http.SameSite
}

// NewCookieCsrfTokenRepository 创建基于 Cookie 的 CSRF 令牌仓库。
func NewCookieCsrfTokenRepository() *CookieCsrfTokenRepository {
	return &CookieCsrfTokenRepository{
		cookieName:     "_csrf_token",
		headerName:     "X-CSRF-Token",
		cookieHttpOnly: false,
		secure:         false,
		sameSite:       http.SameSiteLaxMode,
	}
}

// GenerateToken 生成新的 CSRF 令牌。
func (r *CookieCsrfTokenRepository) GenerateToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error) {
	token := generateSecureToken(32)
	return &CsrfToken{
		Identifier: request.GetURI(),
		Value:      token,
	}, nil
}

// ValidateToken 验证 CSRF 令牌。
func (r *CookieCsrfTokenRepository) ValidateToken(ctx context.Context, request SecurityRequest, token string) bool {
	savedToken, exists := request.GetAttribute("_csrf_token")
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken.(string))) == 1
}

// SaveToken 保存 CSRF 令牌到 Cookie。
func (r *CookieCsrfTokenRepository) SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=%s; Path=/; HttpOnly=%v; SameSite=%v",
		r.cookieName, token.Value, r.cookieHttpOnly, r.sameSite))
}

// ClearToken 清除 CSRF 令牌。
func (r *CookieCsrfTokenRepository) ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", r.cookieName))
}

// generateSecureToken 生成密码学安全的随机令牌。
//
// 使用 crypto/rand 确保生成的 token 不可预测，
// 防止攻击者通过推测随机数种子来伪造 CSRF token。
func generateSecureToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		// 在极罕见情况下（系统熵池耗尽），panic 是合理的选择
		panic(fmt.Sprintf("failed to generate secure token: %v", err))
	}
	// 使用 URL 安全的 base64 编码
	return base64.URLEncoding.EncodeToString(b)
}

// CsrfTokenManager CSRF 令牌管理器。
type CsrfTokenManager struct {
	tokens map[string]string
	mu     sync.RWMutex
}

// NewCsrfTokenManager 创建 CSRF 令牌管理器。
func NewCsrfTokenManager() *CsrfTokenManager {
	return &CsrfTokenManager{
		tokens: make(map[string]string),
	}
}

// GenerateToken 为用户生成 CSRF 令牌。
func (m *CsrfTokenManager) GenerateToken(principal string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	token := generateSecureToken(32)
	m.tokens[principal] = token
	return token
}

// ValidateToken 验证 CSRF 令牌。
func (m *CsrfTokenManager) ValidateToken(principal, token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	savedToken, exists := m.tokens[principal]
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken)) == 1
}

// RemoveToken 移除用户的 CSRF 令牌。
func (m *CsrfTokenManager) RemoveToken(principal string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, principal)
}

// CsrfAuthenticationStrategy CSRF 令牌会话认证策略。
type CsrfAuthenticationStrategy struct {
	mu           sync.Mutex
	tokenManager *CsrfTokenManager
}

// NewCsrfAuthenticationStrategy 创建 CSRF 认证策略。
func NewCsrfAuthenticationStrategy() *CsrfAuthenticationStrategy {
	return &CsrfAuthenticationStrategy{
		tokenManager: NewCsrfTokenManager(),
	}
}

// OnAuthentication 认证成功后生成新的 CSRF 令牌。
func (s *CsrfAuthenticationStrategy) OnAuthentication(ctx context.Context, authentication Authentication, request SecurityRequest, response SecurityResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token := s.tokenManager.GenerateToken(authentication.Name())
	request.SetAttribute("csrf.token", token)
}

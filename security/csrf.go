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

	"github.com/xudefa/enhance/security/filter"
)

// CsrfFilter CSRF防护过滤器。
//
// 在表单提交时验证 CSRF Token，防止跨站请求伪造攻击。
type CsrfFilter struct {
	tokenRepository CsrfTokenRepository
	excludePaths    []string
}

// NewCsrfFilter 创建 CSRF 防护过滤器。
//
// 参数:
//   - tokenRepository: CSRF Token 存储仓库，用于生成和验证 Token
//
// 返回:
//   - *CsrfFilter: CSRF 过滤器实例
//   - error: 参数错误
func NewCsrfFilter(tokenRepository CsrfTokenRepository) (*CsrfFilter, error) {
	if tokenRepository == nil {
		return nil, fmt.Errorf("csrf: tokenRepository must not be nil")
	}
	return &CsrfFilter{
		tokenRepository: tokenRepository,
		excludePaths:    []string{},
	}, nil
}

// MustNewCsrfFilter 创建 CSRF 防护过滤器，失败则 panic。
func MustNewCsrfFilter(tokenRepository CsrfTokenRepository) *CsrfFilter {
	filter, err := NewCsrfFilter(tokenRepository)
	if err != nil {
		panic(err)
	}
	return filter
}

// AddExcludePath 添加不需要 CSRF 防护的路径。
//
// 参数:
//   - paths: 要排除的路径列表，必须以 "/" 开头
//
// 返回:
//   - error: 路径格式错误时返回错误
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

// DoFilter 实现 filter.Filter 接口
func (f *CsrfFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("CsrfFilter: ctx must be context.Context")
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("CsrfFilter: request must be SecurityRequest")
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("CsrfFilter: response must be SecurityResponse")
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *CsrfFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	uri := request.GetURI()

	for _, path := range f.excludePaths {
		if strings.HasPrefix(uri, path) {
			return chain.DoFilter(ctx, request, response)
		}
	}

	method := request.GetMethod()

	if method == "GET" || method == "HEAD" || method == "OPTIONS" || method == "TRACE" {
		// 仅在会话尚未持有令牌时生成，保持令牌在多次 GET 间稳定，
		// 避免页面加载与表单提交之间令牌变化导致校验失败。
		existing, err := f.tokenRepository.LoadToken(ctx, request)
		if err == nil && existing == nil {
			token, err := f.tokenRepository.GenerateToken(ctx, request)
			if err == nil && token != nil {
				f.tokenRepository.SaveToken(ctx, request, response, token)
				request.SetAttribute("csrf.token", token.Value)
			}
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

// Order 实现 filter.Filter 接口
func (f *CsrfFilter) Order() int { return 0 }

// CookieCsrfTokenRepository 基于Cookie的CSRF令牌仓库
type CookieCsrfTokenRepository struct {
	cookieName     string
	headerName     string
	cookieHttpOnly bool
	secure         bool
	sameSite       http.SameSite
}

func NewCookieCsrfTokenRepository() *CookieCsrfTokenRepository {
	return &CookieCsrfTokenRepository{
		cookieName:     "_csrf_token",
		headerName:     "X-CSRF-Token",
		cookieHttpOnly: false,
		secure:         false,
		sameSite:       http.SameSiteLaxMode,
	}
}

func (r *CookieCsrfTokenRepository) GenerateToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error) {
	token, err := generateSecureToken(32)
	if err != nil {
		return nil, err
	}
	return &CsrfToken{
		Identifier: request.GetURI(),
		Value:      token,
	}, nil
}

// LoadToken 从 Cookie 中加载已存在的 CSRF 令牌，未找到时返回 nil。
func (r *CookieCsrfTokenRepository) LoadToken(ctx context.Context, request SecurityRequest) (*CsrfToken, error) {
	value, exists := r.loadCookieValue(request)
	if !exists {
		return nil, nil
	}
	return &CsrfToken{
		Identifier: request.GetURI(),
		Value:      value,
	}, nil
}

func (r *CookieCsrfTokenRepository) ValidateToken(ctx context.Context, request SecurityRequest, token string) bool {
	savedToken, exists := request.GetAttribute("csrf.token")
	if !exists {
		savedTokenStr, ok := r.loadCookieValue(request)
		if !ok {
			return false
		}
		return subtle.ConstantTimeCompare([]byte(token), []byte(savedTokenStr)) == 1
	}
	savedTokenStr, ok := savedToken.(string)
	if !ok {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedTokenStr)) == 1
}

// loadCookieValue 从 Cookie 请求头中解析令牌值。
func (r *CookieCsrfTokenRepository) loadCookieValue(request SecurityRequest) (string, bool) {
	cookieHeader := request.GetHeader("Cookie")
	if cookieHeader == "" {
		return "", false
	}
	prefix := r.cookieName + "="
	for _, part := range strings.Split(cookieHeader, ";") {
		part = strings.TrimSpace(part)
		if strings.HasPrefix(part, prefix) {
			return strings.TrimPrefix(part, prefix), true
		}
	}
	return "", false
}

func sameSiteString(s http.SameSite) string {
	switch s {
	case http.SameSiteLaxMode:
		return "Lax"
	case http.SameSiteStrictMode:
		return "Strict"
	case http.SameSiteNoneMode:
		return "None"
	default:
		return "Lax"
	}
}

func (r *CookieCsrfTokenRepository) SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=%s; Path=/; HttpOnly=%t; Secure=%t; SameSite=%s",
		r.cookieName, token.Value, r.cookieHttpOnly, r.secure, sameSiteString(r.sameSite)))
}

func (r *CookieCsrfTokenRepository) ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", r.cookieName))
}

func generateSecureToken(length int) (string, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate secure token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// CsrfTokenManager CSRF 令牌管理器
type CsrfTokenManager struct {
	tokens map[string]string
	mu     sync.RWMutex
}

func NewCsrfTokenManager() *CsrfTokenManager {
	return &CsrfTokenManager{
		tokens: make(map[string]string),
	}
}

func (m *CsrfTokenManager) GenerateToken(principal string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	token, err := generateSecureToken(32)
	if err != nil {
		return "", err
	}
	m.tokens[principal] = token
	return token, nil
}

func (m *CsrfTokenManager) ValidateToken(principal, token string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	savedToken, exists := m.tokens[principal]
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken)) == 1
}

func (m *CsrfTokenManager) RemoveToken(principal string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, principal)
}

// CsrfAuthenticationStrategy CSRF 令牌会话认证策略
type CsrfAuthenticationStrategy struct {
	mu           sync.Mutex
	tokenManager *CsrfTokenManager
}

func NewCsrfAuthenticationStrategy() *CsrfAuthenticationStrategy {
	return &CsrfAuthenticationStrategy{
		tokenManager: NewCsrfTokenManager(),
	}
}

// OnAuthentication 认证成功后生成新的 CSRF 令牌
func (s *CsrfAuthenticationStrategy) OnAuthentication(ctx context.Context, authentication Authentication, request SecurityRequest, response SecurityResponse) {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, err := s.tokenManager.GenerateToken(extractPrincipalName(authentication))
	if err != nil {
		// 令牌生成失败，记录错误但不阻止认证
		return
	}
	request.SetAttribute("csrf.token", token)
}

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

// CsrfFilter CSRF防护过滤器
type CsrfFilter struct {
	tokenRepository CsrfTokenRepository
	excludePaths    []string
}

func NewCsrfFilter(tokenRepository CsrfTokenRepository) *CsrfFilter {
	return &CsrfFilter{
		tokenRepository: tokenRepository,
		excludePaths:    []string{},
	}
}

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
	ctxVal, _ := ctx.(context.Context)
	req, _ := request.(SecurityRequest)
	resp, _ := response.(SecurityResponse)
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
	token := generateSecureToken(32)
	return &CsrfToken{
		Identifier: request.GetURI(),
		Value:      token,
	}, nil
}

func (r *CookieCsrfTokenRepository) ValidateToken(ctx context.Context, request SecurityRequest, token string) bool {
	savedToken, exists := request.GetAttribute("csrf.token")
	if !exists {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(token), []byte(savedToken.(string))) == 1
}

func (r *CookieCsrfTokenRepository) SaveToken(ctx context.Context, request SecurityRequest, response SecurityResponse, token *CsrfToken) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=%s; Path=/; HttpOnly=%v; SameSite=%v",
		r.cookieName, token.Value, r.cookieHttpOnly, r.sameSite))
}

func (r *CookieCsrfTokenRepository) ClearToken(ctx context.Context, request SecurityRequest, response SecurityResponse) {
	response.SetHeader("Set-Cookie", fmt.Sprintf("%s=; Path=/; Max-Age=0", r.cookieName))
}

func generateSecureToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		panic(fmt.Sprintf("failed to generate secure token: %v", err))
	}
	return base64.URLEncoding.EncodeToString(b)
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

func (m *CsrfTokenManager) GenerateToken(principal string) string {
	m.mu.Lock()
	defer m.mu.Unlock()
	token := generateSecureToken(32)
	m.tokens[principal] = token
	return token
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

	token := s.tokenManager.GenerateToken(extractPrincipalName(authentication))
	request.SetAttribute("csrf.token", token)
}

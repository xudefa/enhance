package jwt

import (
	"context"
	"fmt"
	"strings"

	"github.com/xudefa/enhance/security"
	"github.com/xudefa/enhance/security/filter"
)

// JwtAuthenticationFilter JWT 认证过滤器。
type JwtAuthenticationFilter struct {
	tokenProvider      TokenProvider
	excludePaths       []string
	userDetailsService security.UserDetailsService
}

func NewJwtAuthenticationFilter(tokenProvider TokenProvider, opts ...JwtFilterOption) *JwtAuthenticationFilter {
	f := &JwtAuthenticationFilter{
		tokenProvider: tokenProvider,
	}
	for _, opt := range opts {
		opt(f)
	}
	return f
}

type JwtFilterOption func(*JwtAuthenticationFilter)

func WithExcludePaths(paths ...string) JwtFilterOption {
	return func(f *JwtAuthenticationFilter) {
		f.excludePaths = paths
	}
}

func WithUserDetailsService(service security.UserDetailsService) JwtFilterOption {
	return func(f *JwtAuthenticationFilter) {
		f.userDetailsService = service
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *JwtAuthenticationFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(security.SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(security.SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *JwtAuthenticationFilter) doFilter(ctx context.Context, request security.SecurityRequest, response security.SecurityResponse, chain filter.FilterChain) error {
	uri := request.GetURI()
	if f.isExcluded(uri) {
		return chain.DoFilter(ctx, request, response)
	}

	authHeader := request.GetHeader(HeaderAuthorization)
	if authHeader == "" {
		return chain.DoFilter(ctx, request, response)
	}

	token := extractBearerToken(authHeader)
	if token == "" {
		return chain.DoFilter(ctx, request, response)
	}

	claims, err := f.tokenProvider.ParseToken(ctx, token)
	if err != nil {
		response.SetStatusCode(401)
		response.SetHeader("Content-Type", "application/json")
		errorMsg := fmt.Sprintf(`{"code":401,"message":"jwt authentication failed: %v"}`, err)
		if writeErr := response.Write([]byte(errorMsg)); writeErr != nil {
			return fmt.Errorf("jwt authentication failed: %w", err)
		}
		return fmt.Errorf("invalid jwt token: %w", err)
	}

	var userDetails security.UserDetails
	if f.userDetailsService != nil {
		userDetails, err = f.userDetailsService.LoadUserByUsername(ctx, claims.Subject)
		if err != nil {
			userDetails = NewSimpleUserDetails(claims.Subject, claims.Authorities)
		}
	} else {
		userDetails = NewSimpleUserDetails(claims.Subject, claims.Authorities)
	}

	auth := security.NewAuthenticatedUsernamePasswordAuthenticationToken(userDetails, userDetails.Authorities())

	ctx = security.ContextWithAuthentication(ctx, auth)
	request.SetAttribute("security.currentAuthentication", auth)

	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *JwtAuthenticationFilter) Order() int { return -500 }

// IsJwtFilter 标识这是一个 JWT 认证过滤器。
func (f *JwtAuthenticationFilter) IsJwtFilter() bool {
	return true
}

func (f *JwtAuthenticationFilter) isExcluded(uri string) bool {
	for _, path := range f.excludePaths {
		if uri == path || matchPattern(path, uri) {
			return true
		}
	}
	return false
}

func extractBearerToken(authHeader string) string {
	if !strings.HasPrefix(authHeader, HeaderBearerPrefix) {
		return ""
	}
	return strings.TrimPrefix(authHeader, HeaderBearerPrefix)
}

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

func NewSimpleUserDetails(username string, authorities []string) *SimpleUserDetails {
	return &SimpleUserDetails{
		username:    username,
		authorities: authorities,
	}
}

func (u *SimpleUserDetails) Username() string            { return u.username }
func (u *SimpleUserDetails) Password() string            { return "" }
func (u *SimpleUserDetails) Authorities() []string       { return u.authorities }
func (u *SimpleUserDetails) Enabled() bool               { return true }
func (u *SimpleUserDetails) AccountNonExpired() bool     { return true }
func (u *SimpleUserDetails) CredentialsNonExpired() bool { return true }
func (u *SimpleUserDetails) AccountNonLocked() bool      { return true }

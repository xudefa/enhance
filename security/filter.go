package security

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/xudefa/enhance/security/filter"
)

// Security filter order constants
const (
	AuthContextFilterOrder             = -1000
	AnonymousAuthenticationFilterOrder = 0
	ExceptionTranslationFilterOrder    = 0
	FilterSecurityInterceptorOrder     = 100
)

// Specificity score constants for pattern matching
const (
	specificityExactMatch       = 1000
	specificitySuffixDoubleStar = 100
	specificitySuffixSingleStar = 50
)

// Anonymous authentication default values
const (
	defaultAnonymousKey       = "anonymousKey"
	defaultAnonymousUser      = "anonymousUser"
	defaultAnonymousAuthority = "ROLE_ANONYMOUS"
)

// filterAppliedKey is the request attribute key used to prevent double filter execution.
const filterAppliedKey = "FILTER_APPLIED"

// AuthContextFilter 认证上下文过滤器。
//
// 在过滤器链执行完成后，将最终认证信息保存到请求属性中，
// 供下游 HTTP 处理器使用。
type AuthContextFilter struct{}

func NewAuthContextFilter() *AuthContextFilter {
	return &AuthContextFilter{}
}

// DoFilter 实现 filter.Filter 接口
func (f *AuthContextFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *AuthContextFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	err := chain.DoFilter(ctx, request, response)

	if _, exists := request.GetAttribute("security.currentAuthentication"); !exists {
		if currentAuth := GetAuthenticationFromContext(ctx); currentAuth != nil {
			request.SetAttribute("security.currentAuthentication", currentAuth)
		}
	}

	return err
}

// Order 实现 filter.Filter 接口
func (f *AuthContextFilter) Order() int { return AuthContextFilterOrder }

// AnonymousAuthenticationFilter 匿名认证过滤器
type AnonymousAuthenticationFilter struct {
	key         string
	principal   any
	authorities []string
}

func NewAnonymousAuthenticationFilter() *AnonymousAuthenticationFilter {
	return &AnonymousAuthenticationFilter{
		key:         defaultAnonymousKey,
		principal:   defaultAnonymousUser,
		authorities: []string{defaultAnonymousAuthority},
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *AnonymousAuthenticationFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *AnonymousAuthenticationFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	auth := GetAuthenticationFromContext(ctx)
	if auth == nil {
		auth = NewAnonymousAuthenticationToken(f.key, f.principal, f.authorities)
		ctx = ContextWithAuthentication(ctx, auth)
	}
	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *AnonymousAuthenticationFilter) Order() int { return AnonymousAuthenticationFilterOrder }

// AnonymousAuthenticationToken 匿名认证令牌
type AnonymousAuthenticationToken struct {
	principal     any
	authorities   []string
	authenticated bool
}

func NewAnonymousAuthenticationToken(key string, principal any, authorities []string) *AnonymousAuthenticationToken {
	return &AnonymousAuthenticationToken{
		principal:     principal,
		authorities:   authorities,
		authenticated: false,
	}
}

// Principal 返回匿名认证主体的身份信息。
func (t *AnonymousAuthenticationToken) Principal() any { return t.principal }

// Credentials 返回匿名认证的凭据（始终为 nil）。
func (t *AnonymousAuthenticationToken) Credentials() any { return nil }

// Authorities 返回匿名认证的授权列表。
func (t *AnonymousAuthenticationToken) Authorities() []string { return t.authorities }

// Authenticated 返回匿名认证是否已通过验证。
func (t *AnonymousAuthenticationToken) Authenticated() bool { return t.authenticated }

// Name 返回匿名认证主体的名称字符串。
func (t *AnonymousAuthenticationToken) Name() string {
	if name, ok := t.principal.(string); ok {
		return name
	}
	return ""
}

// ExceptionTranslationFilter 异常转换过滤器
type ExceptionTranslationFilter struct {
	accessDeniedHandler      AccessDeniedHandler
	authenticationEntryPoint AuthenticationEntryPoint
}

func NewExceptionTranslationFilter(accessDeniedHandler AccessDeniedHandler, authenticationEntryPoint AuthenticationEntryPoint) *ExceptionTranslationFilter {
	return &ExceptionTranslationFilter{
		accessDeniedHandler:      accessDeniedHandler,
		authenticationEntryPoint: authenticationEntryPoint,
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *ExceptionTranslationFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *ExceptionTranslationFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	err := chain.DoFilter(ctx, request, response)
	if err != nil {
		if errors.Is(err, ErrAccessDenied) {
			auth := GetAuthenticationFromContext(ctx)
			if auth == nil || !auth.Authenticated() {
				if f.authenticationEntryPoint != nil {
					return f.authenticationEntryPoint.Commence(ctx, request, response, err)
				}
				return err
			}
			if f.accessDeniedHandler != nil {
				return f.accessDeniedHandler.Handle(ctx, request, response, err)
			}
		}
		return err
	}
	return nil
}

// Order 实现 filter.Filter 接口
func (f *ExceptionTranslationFilter) Order() int { return ExceptionTranslationFilterOrder }

// FilterSecurityInterceptor 过滤器安全拦截器
type FilterSecurityInterceptor struct {
	securityMetadataSource SecurityMetadataSource
	accessDecisionManager  AccessDecisionManager
	authenticationManager  AuthenticationManager
	observeOncePerRequest  bool
}

func NewFilterSecurityInterceptor(securityMetadataSource SecurityMetadataSource, accessDecisionManager AccessDecisionManager, authenticationManager AuthenticationManager) *FilterSecurityInterceptor {
	return &FilterSecurityInterceptor{
		securityMetadataSource: securityMetadataSource,
		accessDecisionManager:  accessDecisionManager,
		authenticationManager:  authenticationManager,
		observeOncePerRequest:  true,
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *FilterSecurityInterceptor) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return f.doFilter(ctxVal, req, resp, chain)
}

func (f *FilterSecurityInterceptor) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	attributes, err := f.securityMetadataSource.GetAttributes(ctx, request)
	if err != nil {
		return err
	}

	if f.observeOncePerRequest {
		attrKey := filterAppliedKey
		if _, exists := request.GetAttribute(attrKey); exists {
			return chain.DoFilter(ctx, request, response)
		}
		request.SetAttribute(attrKey, true)
	}

	auth := GetAuthenticationFromContext(ctx)

	if len(attributes) == 0 {
		return chain.DoFilter(ctx, request, response)
	}

	resource := fmt.Sprintf("%s:%s", request.GetMethod(), request.GetURI())
	if err := f.accessDecisionManager.Decide(ctx, auth, resource, attributes); err != nil {
		return err
	}

	return chain.DoFilter(ctx, request, response)
}

// Order 实现 filter.Filter 接口
func (f *FilterSecurityInterceptor) Order() int { return FilterSecurityInterceptorOrder }

// SetSecurityMetadataSource 设置安全元数据源。
func (f *FilterSecurityInterceptor) SetSecurityMetadataSource(source SecurityMetadataSource) {
	f.securityMetadataSource = source
}

// SetAccessDecisionManager 设置访问决策管理器。
func (f *FilterSecurityInterceptor) SetAccessDecisionManager(manager AccessDecisionManager) {
	f.accessDecisionManager = manager
}

// SetAuthenticationManager 设置认证管理器。
func (f *FilterSecurityInterceptor) SetAuthenticationManager(manager AuthenticationManager) {
	f.authenticationManager = manager
}

// ExpressionBasedFilterInvocationSecurityMetadataSource 基于表达式的过滤器调用安全元数据源
type ExpressionBasedFilterInvocationSecurityMetadataSource struct {
	mu         sync.RWMutex
	requestMap map[string][]string
}

// NewExpressionBasedFilterInvocationSecurityMetadataSource 创建基于表达式的过滤器调用安全元数据源实例。
func NewExpressionBasedFilterInvocationSecurityMetadataSource() *ExpressionBasedFilterInvocationSecurityMetadataSource {
	return &ExpressionBasedFilterInvocationSecurityMetadataSource{
		requestMap: make(map[string][]string),
	}
}

// AddMapping 添加 URL 模式与安全属性的映射关系。
func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) AddMapping(pattern string, attributes []string) {
	if !strings.Contains(pattern, "::") {
		pattern = "**::" + pattern
	}
	s.mu.Lock()
	s.requestMap[pattern] = attributes
	s.mu.Unlock()
}

// GetAttributes 根据请求获取匹配的安全属性列表。
func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) GetAttributes(ctx context.Context, request SecurityRequest) ([]string, error) {
	uri := request.GetURI()
	method := request.GetMethod()

	type matchedRule struct {
		pattern     string
		attributes  []string
		specificity int
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	matchedRules := make([]matchedRule, 0, len(s.requestMap))
	for pattern, attributes := range s.requestMap {
		if s.matches(pattern, uri, method) {
			specificity := s.calculateSpecificity(pattern)
			matchedRules = append(matchedRules, matchedRule{
				pattern:     pattern,
				attributes:  attributes,
				specificity: specificity,
			})
		}
	}

	if len(matchedRules) == 0 {
		return []string{}, nil
	}

	bestMatch := matchedRules[0]
	for _, rule := range matchedRules[1:] {
		if rule.specificity > bestMatch.specificity {
			bestMatch = rule
		}
	}

	return bestMatch.attributes, nil
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) calculateSpecificity(pattern string) int {
	parts := strings.Split(pattern, "::")
	if len(parts) != 2 {
		return 0
	}
	pathPattern := parts[1]
	specificity := len(pathPattern)
	if !strings.Contains(pathPattern, "*") {
		specificity += specificityExactMatch
	}
	if strings.HasSuffix(pathPattern, "/**") {
		specificity += specificitySuffixDoubleStar
	} else if strings.HasSuffix(pathPattern, "/*") {
		specificity += specificitySuffixSingleStar
	}
	return specificity
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) matches(pattern, uri, method string) bool {
	parts := strings.Split(pattern, "::")
	if len(parts) != 2 {
		return false
	}
	methodPattern := parts[0]
	pathPattern := parts[1]

	if methodPattern != method && methodPattern != "**" {
		return false
	}
	if pathPattern == "/**" {
		return true
	}
	if pathPattern == uri {
		return true
	}
	if strings.HasSuffix(pathPattern, "/**") {
		prefix := strings.TrimSuffix(pathPattern, "/**")
		return strings.HasPrefix(uri, prefix) || uri == strings.TrimSuffix(prefix, "/")
	}
	if strings.Contains(pathPattern, "*") {
		return s.matchWildcard(pathPattern, uri)
	}
	return false
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) matchWildcard(pattern, str string) bool {
	patternParts := strings.Split(pattern, "/")
	strParts := strings.Split(str, "/")
	if len(patternParts) != len(strParts) {
		return false
	}
	for i, patternPart := range patternParts {
		if patternPart != "*" && patternPart != strParts[i] {
			return false
		}
	}
	return true
}

// Http403ForbiddenEntryPoint 403禁止访问入口点
type Http403ForbiddenEntryPoint struct{}

func NewHttp403ForbiddenEntryPoint() *Http403ForbiddenEntryPoint {
	return &Http403ForbiddenEntryPoint{}
}

// Commence 发送 403 Forbidden 响应给客户端。
func (e *Http403ForbiddenEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(http.StatusForbidden)
	if writeErr := response.Write([]byte(http.StatusText(http.StatusForbidden))); writeErr != nil {
		fmt.Printf("[enhance] failed to write 403 response: %v\n", writeErr)
	}
	return nil
}

// Http401UnauthorizedEntryPoint 401未认证入口点
type Http401UnauthorizedEntryPoint struct{}

func NewHttp401UnauthorizedEntryPoint() *Http401UnauthorizedEntryPoint {
	return &Http401UnauthorizedEntryPoint{}
}

// Commence 发送 401 Unauthorized 响应给客户端。
func (e *Http401UnauthorizedEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(http.StatusUnauthorized)
	if writeErr := response.Write([]byte(http.StatusText(http.StatusUnauthorized))); writeErr != nil {
		fmt.Printf("[enhance] failed to write 401 response: %v\n", writeErr)
	}
	return nil
}

// Http403ForbiddenAccessDeniedHandler 403禁止访问拒绝处理器
type Http403ForbiddenAccessDeniedHandler struct{}

func NewHttp403ForbiddenAccessDeniedHandler() *Http403ForbiddenAccessDeniedHandler {
	return &Http403ForbiddenAccessDeniedHandler{}
}

// Handle 处理访问被拒绝的情况，发送 403 Forbidden 响应给客户端。
func (e *Http403ForbiddenAccessDeniedHandler) Handle(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(http.StatusForbidden)
	if writeErr := response.Write([]byte(http.StatusText(http.StatusForbidden))); writeErr != nil {
		fmt.Printf("[enhance] failed to write 403 access denied response: %v\n", writeErr)
	}
	return nil
}

// LoginUrlAuthenticationEntryPoint 登录URL认证入口点
type LoginUrlAuthenticationEntryPoint struct {
	loginFormUrl string
}

func NewLoginUrlAuthenticationEntryPoint(loginFormUrl string) *LoginUrlAuthenticationEntryPoint {
	return &LoginUrlAuthenticationEntryPoint{
		loginFormUrl: loginFormUrl,
	}
}

// Commence 重定向客户端到登录页面。
func (e *LoginUrlAuthenticationEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(http.StatusFound)
	response.SetHeader("Location", e.loginFormUrl)
	return nil
}

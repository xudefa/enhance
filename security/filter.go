package security

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/xudefa/enhance/security/filter"
)

// Security filter order constants
const (
	SecurityContextHolderFilterOrder = -1000
	AnonymousAuthenticationFilterOrder = 0
	ExceptionTranslationFilterOrder    = 0
	FilterSecurityInterceptorOrder     = 100
)

// Specificity score constants for pattern matching
const (
	specificityExactMatch      = 1000
	specificitySuffixDoubleStar = 100
	specificitySuffixSingleStar = 50
)

// Anonymous authentication default values
const (
	defaultAnonymousKey      = "anonymousKey"
	defaultAnonymousUser      = "anonymousUser"
	defaultAnonymousAuthority = "ROLE_ANONYMOUS"
)

// filterAppliedKey is the request attribute key used to prevent double filter execution.
const filterAppliedKey = "FILTER_APPLIED"

// securityFilterChainAdapter 将内部 typed FilterChainProxy 适配为 filter.SecurityFilterChain 接口。
type securityFilterChainAdapter struct {
	proxy *FilterChainProxy
}

func (a *securityFilterChainAdapter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
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
	return a.proxy.doFilterWithChain(ctxVal, req, resp, &filterChainAdapter{vfc: &VirtualFilterChain{proxy: a.proxy, index: 0}})
}

func (a *securityFilterChainAdapter) Matches(request interface{}) bool {
	_, ok := request.(SecurityRequest)
	return ok
}

func (a *securityFilterChainAdapter) GetFilters() []filter.Filter {
	result := make([]filter.Filter, len(a.proxy.filters))
	for i, f := range a.proxy.filters {
		result[i] = f
	}
	return result
}

// filterChainAdapter 将 VirtualFilterChain 适配为 filter.FilterChain 接口。
type filterChainAdapter struct {
	vfc *VirtualFilterChain
}

func (a *filterChainAdapter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
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
	nextChain := &filterChainAdapter{vfc: &VirtualFilterChain{proxy: a.vfc.proxy, index: a.vfc.index}}
	return a.vfc.proxy.doFilterWithChain(ctxVal, req, resp, nextChain)
}

func (a *filterChainAdapter) AddFilter(f filter.Filter) {}

func (a *filterChainAdapter) GetFilters() []filter.Filter {
	return nil
}

// FilterChainProxy 过滤器链代理
type FilterChainProxy struct {
	filters []SecurityFilter
	chain   SecurityFilterChain
}

func NewFilterChainProxy(filters []SecurityFilter, chain SecurityFilterChain) *FilterChainProxy {
	return &FilterChainProxy{
		filters: filters,
		chain:   chain,
	}
}

// DoFilter 实现 filter.SecurityFilterChain 接口
func (p *FilterChainProxy) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
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
	return p.doFilterWithChain(ctxVal, req, resp, &filterChainAdapter{vfc: &VirtualFilterChain{proxy: p, index: 0}})
}

// Matches 实现 filter.SecurityFilterChain 接口（FilterChainProxy 匹配所有请求）
func (p *FilterChainProxy) Matches(request interface{}) bool {
	return true
}

// GetFilters 实现 filter.SecurityFilterChain 接口
func (p *FilterChainProxy) GetFilters() []filter.Filter {
	result := make([]filter.Filter, len(p.filters))
	for i, f := range p.filters {
		result[i] = f
	}
	return result
}

// doFilterWithChain 以类型安全方式执行过滤器链
func (p *FilterChainProxy) doFilterWithChain(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	if p == nil || chain == nil {
		return nil
	}
	return chain.DoFilter(ctx, request, response)
}

// doFilterInternal 执行过滤器链中的指定索引过滤器
func (p *FilterChainProxy) doFilterInternal(ctx context.Context, request SecurityRequest, response SecurityResponse, index int) error {
	if index >= len(p.filters) {
		return p.chain.DoFilter(ctx, request, response)
	}

	nextChain := &filterChainAdapter{
		vfc: &VirtualFilterChain{proxy: p, index: index + 1},
	}
	return p.filters[index].DoFilter(ctx, request, response, nextChain)
}

// VirtualFilterChain 虚拟过滤器链
type VirtualFilterChain struct {
	proxy *FilterChainProxy
	index int
}

// DoFilter 执行下一个过滤器
func (c *VirtualFilterChain) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	return c.proxy.doFilterInternal(ctx, request, response, c.index)
}

// SecurityContextHolderFilter 安全上下文持有者过滤器
type SecurityContextHolderFilter struct {
	securityContext SecurityContext
}

func NewSecurityContextHolderFilter(securityContext SecurityContext) *SecurityContextHolderFilter {
	return &SecurityContextHolderFilter{
		securityContext: securityContext,
	}
}

// DoFilter 实现 filter.Filter 接口
func (f *SecurityContextHolderFilter) DoFilter(ctx interface{}, request interface{}, response interface{}, chain filter.FilterChain) error {
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

func (f *SecurityContextHolderFilter) doFilter(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	prevAuth := GetAuthentication()
	ClearAuthentication()

	err := chain.DoFilter(ctx, request, response)

	currentAuth := GetAuthentication()
	if currentAuth != nil {
		request.SetAttribute("security.currentAuthentication", currentAuth)
	}

	if f.securityContext != nil {
		f.securityContext.ClearAuthentication()
		if prevAuth != nil {
			f.securityContext.SetAuthentication(prevAuth)
		}
	}

	return err
}

// Order 实现 filter.Filter 接口
func (f *SecurityContextHolderFilter) Order() int { return SecurityContextHolderFilterOrder }

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
	if GetAuthentication() == nil {
		auth := NewAnonymousAuthenticationToken(f.key, f.principal, f.authorities)
		SetAuthentication(auth)
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

func (t *AnonymousAuthenticationToken) Principal() any        { return t.principal }
func (t *AnonymousAuthenticationToken) Credentials() any      { return nil }
func (t *AnonymousAuthenticationToken) Authorities() []string { return t.authorities }
func (t *AnonymousAuthenticationToken) Authenticated() bool   { return t.authenticated }

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
		if err == ErrAccessDenied {
			auth := GetAuthentication()
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

	auth := GetAuthentication()

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

func (f *FilterSecurityInterceptor) SetSecurityMetadataSource(source SecurityMetadataSource) {
	f.securityMetadataSource = source
}

func (f *FilterSecurityInterceptor) SetAccessDecisionManager(manager AccessDecisionManager) {
	f.accessDecisionManager = manager
}

func (f *FilterSecurityInterceptor) SetAuthenticationManager(manager AuthenticationManager) {
	f.authenticationManager = manager
}

// ExpressionBasedFilterInvocationSecurityMetadataSource 基于表达式的过滤器调用安全元数据源
type ExpressionBasedFilterInvocationSecurityMetadataSource struct {
	requestMap map[string][]string
}

func NewExpressionBasedFilterInvocationSecurityMetadataSource() *ExpressionBasedFilterInvocationSecurityMetadataSource {
	return &ExpressionBasedFilterInvocationSecurityMetadataSource{
		requestMap: make(map[string][]string),
	}
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) AddMapping(pattern string, attributes []string) {
	if !strings.Contains(pattern, "::") {
		pattern = "**::" + pattern
	}
	s.requestMap[pattern] = attributes
}

func (s *ExpressionBasedFilterInvocationSecurityMetadataSource) GetAttributes(ctx context.Context, request SecurityRequest) ([]string, error) {
	uri := request.GetURI()
	method := request.GetMethod()

	type matchedRule struct {
		pattern     string
		attributes  []string
		specificity int
	}
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

func (e *LoginUrlAuthenticationEntryPoint) Commence(ctx context.Context, request SecurityRequest, response SecurityResponse, err error) error {
	response.SetStatusCode(http.StatusFound)
	response.SetHeader("Location", e.loginFormUrl)
	return nil
}

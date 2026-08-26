package security

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xudefa/enhance/log"
)

// httpSecurity HTTP安全配置实现
type httpSecurity struct {
	authenticationManager      AuthenticationManager
	userDetailsService         UserDetailsService
	passwordEncoder            PasswordEncoder
	accessDecisionManager      AccessDecisionManager
	securityMetadataSource     SecurityMetadataSource
	filters                    []SecurityFilter
	anonymousFilter            *AnonymousAuthenticationFilter
	exceptionTranslationFilter *ExceptionTranslationFilter
	filterSecurityInterceptor  *FilterSecurityInterceptor
	authContextFilter          *AuthContextFilter

	csrfEnabled         bool
	csrfTokenRepository CsrfTokenRepository

	logoutUrl            string
	logoutHandlers       []LogoutHandler
	logoutSuccessHandler LogoutSuccessHandler

	formLoginEnabled   bool
	loginProcessingUrl string
	defaultSuccessUrl  string
	failureUrl         string

	httpBasicEnabled bool
	realmName        string

	authorizeRules []authorizeRule
}

// authorizeRule 授权规则，Build 时应用到元数据源。
type authorizeRule struct {
	patterns []string
	attrs    []string
}

// NewHttpSecurity 创建 HTTP 安全配置构建器。
//
// 返回:
//   - HttpSecurity: 安全配置构建器实例，支持链式调用配置安全策略
func NewHttpSecurity() HttpSecurity {
	return &httpSecurity{
		filters:             make([]SecurityFilter, 0),
		csrfTokenRepository: NewCookieCsrfTokenRepository(),
	}
}

// AuthenticationManager 设置认证管理器并返回 HttpSecurity 构建器。
func (h *httpSecurity) AuthenticationManager(authManager AuthenticationManager) HttpSecurity {
	h.authenticationManager = authManager
	return h
}

// UserDetailsService 设置用户详情服务并返回 HttpSecurity 构建器。
func (h *httpSecurity) UserDetailsService(userDetailsService UserDetailsService) HttpSecurity {
	h.userDetailsService = userDetailsService
	return h
}

// PasswordEncoder 设置密码编码器并返回 HttpSecurity 构建器。
func (h *httpSecurity) PasswordEncoder(encoder PasswordEncoder) HttpSecurity {
	h.passwordEncoder = encoder
	return h
}

// AccessDecisionManager 设置访问决策管理器并返回 HttpSecurity 构建器。
func (h *httpSecurity) AccessDecisionManager(manager AccessDecisionManager) HttpSecurity {
	h.accessDecisionManager = manager
	return h
}

// SecurityMetadataSource 设置安全元数据源并返回 HttpSecurity 构建器。
func (h *httpSecurity) SecurityMetadataSource(source SecurityMetadataSource) HttpSecurity {
	h.securityMetadataSource = source
	return h
}

// AuthorizeRequests 设置授权请求配置器并返回 HttpSecurity 构建器。
func (h *httpSecurity) AuthorizeRequests(config func(authorizer AuthorizeRequests)) HttpSecurity {
	authorizer := &httpSecurityAuthorizer{httpSecurity: h}
	config(authorizer)
	return h
}

// AddFilter 添加安全过滤器到过滤器链。
func (h *httpSecurity) AddFilter(filter SecurityFilter) HttpSecurity {
	h.filters = append(h.filters, filter)
	return h
}

// AddFilterBefore 在指定过滤器之前添加新的安全过滤器。
//
// 参数:
//   - filter: 待添加的过滤器
//   - beforeFilter: 目标前置过滤器
func (h *httpSecurity) AddFilterBefore(filter SecurityFilter, beforeFilter SecurityFilter) HttpSecurity {
	newFilters := make([]SecurityFilter, 0, len(h.filters)+1)
	inserted := false
	for _, f := range h.filters {
		if f == beforeFilter && !inserted {
			newFilters = append(newFilters, filter)
			inserted = true
		}
		newFilters = append(newFilters, f)
	}
	if !inserted {
		newFilters = append(newFilters, filter)
	}
	h.filters = newFilters
	return h
}

// AddFilterAfter 在指定过滤器之后添加新的安全过滤器。
//
// 参数:
//   - filter: 待添加的过滤器
//   - afterFilter: 目标后置过滤器
func (h *httpSecurity) AddFilterAfter(filter SecurityFilter, afterFilter SecurityFilter) HttpSecurity {
	newFilters := make([]SecurityFilter, 0, len(h.filters)+1)
	inserted := false
	for _, f := range h.filters {
		newFilters = append(newFilters, f)
		if f == afterFilter && !inserted {
			newFilters = append(newFilters, filter)
			inserted = true
		}
	}
	if !inserted {
		newFilters = append(newFilters, filter)
	}
	h.filters = newFilters
	return h
}

// Anonymous 启用匿名认证过滤器。
func (h *httpSecurity) Anonymous() HttpSecurity {
	h.anonymousFilter = NewAnonymousAuthenticationFilter()
	return h
}

// ExceptionHandling 设置异常处理，包括访问拒绝处理器和认证入口点。
//
// 参数:
//   - handler: 访问拒绝处理器
//   - entryPoint: 认证入口点
func (h *httpSecurity) ExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) HttpSecurity {
	h.exceptionTranslationFilter = NewExceptionTranslationFilter(handler, entryPoint)
	return h
}

// Csrf 启用 CSRF 保护过滤器。
func (h *httpSecurity) Csrf() HttpSecurity {
	h.csrfEnabled = true
	return h
}

// Logout 配置登出功能，可指定登出 URL 和可选的登出成功处理器。
//
// 参数:
//   - logoutUrl: 登出请求的 URL
//   - successHandler: 可选的登出成功处理器
func (h *httpSecurity) Logout(logoutUrl string, successHandler ...LogoutSuccessHandler) HttpSecurity {
	h.logoutUrl = "/logout"
	if logoutUrl != "" {
		h.logoutUrl = logoutUrl
	}
	h.logoutHandlers = make([]LogoutHandler, 0)
	h.logoutSuccessHandler = nil
	if len(successHandler) > 0 {
		h.logoutSuccessHandler = successHandler[0]
	} else {
		h.logoutSuccessHandler = NewDefaultLogoutSuccessHandler("/login?logout")
	}
	return h
}

// FormLogin 配置表单登录功能，可指定登录处理 URL 和可选的默认成功 URL。
//
// 参数:
//   - loginProcessingUrl: 处理登录请求的 URL
//   - defaultSuccessUrl: 可选的登录成功跳转地址
func (h *httpSecurity) FormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) HttpSecurity {
	h.formLoginEnabled = true
	h.loginProcessingUrl = "/login"
	h.defaultSuccessUrl = "/"
	h.failureUrl = "/login?error"
	if loginProcessingUrl != "" {
		h.loginProcessingUrl = loginProcessingUrl
	}
	if len(defaultSuccessUrl) > 0 && defaultSuccessUrl[0] != "" {
		h.defaultSuccessUrl = defaultSuccessUrl[0]
	}
	return h
}

// HttpBasic 启用 HTTP Basic 认证。
func (h *httpSecurity) HttpBasic() HttpSecurity {
	h.httpBasicEnabled = true
	h.realmName = "Secured Area"
	return h
}

// Build 构建安全过滤器链
func (h *httpSecurity) Build() (SecurityFilterChain, error) {
	if h.authenticationManager == nil {
		return nil, fmt.Errorf("authentication manager is required")
	}

	if h.securityMetadataSource == nil {
		h.securityMetadataSource = NewExpressionBasedFilterInvocationSecurityMetadataSource()
	}

	if source, ok := h.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource); ok {
		for _, rule := range h.authorizeRules {
			for _, pattern := range rule.patterns {
				source.AddMapping(pattern, rule.attrs)
			}
		}
	}

	if h.accessDecisionManager == nil {
		webExpressionVoter := NewWebExpressionVoter()
		authenticatedVoter := NewAuthenticatedVoter()
		roleVoter := NewRoleVoter()
		h.accessDecisionManager = NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)
	}

	h.authContextFilter = NewAuthContextFilter()

	if h.anonymousFilter == nil {
		h.anonymousFilter = NewAnonymousAuthenticationFilter()
	}

	h.filterSecurityInterceptor = NewFilterSecurityInterceptor(
		h.securityMetadataSource,
		h.accessDecisionManager,
		h.authenticationManager,
	)

	if h.exceptionTranslationFilter == nil {
		accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
		unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
		h.exceptionTranslationFilter = NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
	}

	defaultFilters := []SecurityFilter{
		h.authContextFilter,
		h.anonymousFilter,
	}

	if h.csrfEnabled {
		csrfFilter, err := NewCsrfFilter(h.csrfTokenRepository)
		if err != nil {
			return nil, fmt.Errorf("failed to create CSRF filter: %w", err)
		}
		defaultFilters = append(defaultFilters, csrfFilter)
	}

	if h.logoutUrl != "" {
		logoutFilter, err := NewLogoutFilter(h.logoutUrl, h.logoutHandlers)
		if err != nil {
			return nil, fmt.Errorf("failed to create logout filter: %w", err)
		}
		if h.logoutSuccessHandler != nil {
			logoutFilter.SetSuccessHandler(h.logoutSuccessHandler)
		}
		defaultFilters = append(defaultFilters, logoutFilter)
	}

	if h.formLoginEnabled {
		formLoginFilter := NewUsernamePasswordAuthenticationFilterWithDefaults(
			h.loginProcessingUrl,
			h.defaultSuccessUrl,
			h.failureUrl,
			h.authenticationManager,
			log.Build(),
		)
		defaultFilters = append(defaultFilters, formLoginFilter)
	}

	if h.httpBasicEnabled {
		basicFilter := NewBasicAuthenticationFilterWithRealm(h.authenticationManager, h.realmName, log.Build())
		defaultFilters = append(defaultFilters, basicFilter)
	}

	defaultFilters = append(defaultFilters, h.exceptionTranslationFilter)
	defaultFilters = append(defaultFilters, h.filterSecurityInterceptor)

	allFilters := make([]SecurityFilter, 0, len(defaultFilters)+len(h.filters))
	allFilters = append(allFilters, defaultFilters...)
	allFilters = append(allFilters, h.filters...)
	sort.SliceStable(allFilters, func(i, j int) bool {
		return allFilters[i].Order() < allFilters[j].Order()
	})

	proxy := newFilterChainProxy(allFilters, &DefaultSecurityFilterChain{})
	return &securityFilterChainAdapter{proxy: proxy}, nil
}

// DefaultSecurityFilterChain 默认安全过滤器链
type DefaultSecurityFilterChain struct{}

// DoFilter 实现 filter.SecurityFilterChain 接口，默认空操作。
func (c *DefaultSecurityFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	return nil
}

// Matches 实现 filter.SecurityFilterChain 接口，默认匹配所有请求。
func (c *DefaultSecurityFilterChain) Matches(request interface{}) bool {
	return true
}

// GetFilters 返回该过滤器链中的所有安全过滤器（默认返回空）。
func (c *DefaultSecurityFilterChain) GetFilters() []SecurityFilter {
	return nil
}

// httpSecurityAuthorizer HTTP安全授权器
type httpSecurityAuthorizer struct {
	httpSecurity *httpSecurity
}

// AntMatchers 配置匹配指定 URL 模式的授权规则。
func (a *httpSecurityAuthorizer) AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry {
	return &expressionInterceptUrlRegistry{
		httpSecurity: a.httpSecurity,
		patterns:     patterns,
	}
}

// AnyRequest 配置匹配所有请求的授权规则。
func (a *httpSecurityAuthorizer) AnyRequest() ExpressionInterceptUrlRegistry {
	return &expressionInterceptUrlRegistry{
		httpSecurity: a.httpSecurity,
		patterns:     []string{"/**"},
	}
}

// expressionInterceptUrlRegistry URL拦截注册
type expressionInterceptUrlRegistry struct {
	httpSecurity *httpSecurity
	patterns     []string
}

// addRule 记录授权规则，Build 时统一应用到元数据源。
func (r *expressionInterceptUrlRegistry) addRule(attrs []string) HttpSecurity {
	if len(r.patterns) == 0 || len(attrs) == 0 {
		return r.httpSecurity
	}
	r.httpSecurity.authorizeRules = append(r.httpSecurity.authorizeRules, authorizeRule{
		patterns: r.patterns,
		attrs:    attrs,
	})
	return r.httpSecurity
}

// PermitAll 配置匹配的 URL 模式允许所有访问。
func (r *expressionInterceptUrlRegistry) PermitAll() HttpSecurity {
	return r.addRule([]string{"permitAll"})
}

// Authenticated 配置匹配的 URL 模式需要已认证的用户才能访问。
func (r *expressionInterceptUrlRegistry) Authenticated() HttpSecurity {
	return r.addRule([]string{"authenticated"})
}

// HasRole 配置匹配的 URL 模式需要指定的角色才能访问。
func (r *expressionInterceptUrlRegistry) HasRole(role string) HttpSecurity {
	return r.addRule([]string{fmt.Sprintf("hasRole('%s')", role)})
}

// HasAnyRole 配置匹配的 URL 模式需要任意一个指定角色即可访问。
func (r *expressionInterceptUrlRegistry) HasAnyRole(roles ...string) HttpSecurity {
	var sb strings.Builder
	for i, role := range roles {
		if i > 0 {
			sb.WriteString("','")
		}
		sb.WriteString(role)
	}
	return r.addRule([]string{fmt.Sprintf("hasAnyRole('%s')", sb.String())})
}

// HasAuthority 配置匹配的 URL 模式需要指定的权限才能访问。
func (r *expressionInterceptUrlRegistry) HasAuthority(authority string) HttpSecurity {
	return r.addRule([]string{fmt.Sprintf("hasAuthority('%s')", authority)})
}

// HasAnyAuthority 配置匹配的 URL 模式需要任意一个指定权限即可访问。
func (r *expressionInterceptUrlRegistry) HasAnyAuthority(authorities ...string) HttpSecurity {
	var sb strings.Builder
	for i, auth := range authorities {
		if i > 0 {
			sb.WriteString("','")
		}
		sb.WriteString(auth)
	}
	return r.addRule([]string{fmt.Sprintf("hasAnyAuthority('%s')", sb.String())})
}

// DenyAll 配置匹配的 URL 模式拒绝所有访问。
func (r *expressionInterceptUrlRegistry) DenyAll() HttpSecurity {
	return r.addRule([]string{"denyAll"})
}

// WebSecurity Web安全配置入口
type WebSecurity struct{}

// NewWebSecurity 创建 Web 安全配置入口实例。
func NewWebSecurity() *WebSecurity {
	return &WebSecurity{}
}

// HttpSecurity 创建一个新的 HttpSecurity 构建器实例。
func (w *WebSecurity) HttpSecurity() *httpSecurity {
	return &httpSecurity{
		filters: make([]SecurityFilter, 0),
	}
}

// Build 构建安全过滤器链（WebSecurity 的 Build 方法仅用作占位）。
func (w *WebSecurity) Build() (SecurityFilterChain, error) {
	return nil, fmt.Errorf("use HttpSecurity().Build() instead")
}

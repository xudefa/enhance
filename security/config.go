package security

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xudefa/enhance/log"
)

// SecurityConfig 安全配置结构体。
//
// 作为函数式选项模式的核心配置对象，替代 HttpSecurity 的 God Interface。
// 所有安全配置通过 WithXxx 选项函数设置，编译期类型安全。
type SecurityConfig struct {
	AuthenticationManager    AuthenticationManager
	UserDetailsService       UserDetailsService
	PasswordEncoder          PasswordEncoder
	AccessDecisionManager    AccessDecisionManager
	SecurityMetadataSource   SecurityMetadataSource
	Filters                  []filterConfigEntry
	Anonymous                bool
	ExceptionHandling        *exceptionHandlingConfig
	CsrfEnabled              bool
	CsrfTokenRepository      CsrfTokenRepository
	Logout                   *logoutConfig
	FormLogin                *formLoginConfig
	HttpBasic                bool
	HttpBasicRealm           string
	AuthorizeRules           []authorizeRule
	AccessDeniedHandler      AccessDeniedHandler
	AuthenticationEntryPoint AuthenticationEntryPoint
}

type filterConfigEntry struct {
	Filter SecurityFilter
	Before SecurityFilter
	After  SecurityFilter
}

type exceptionHandlingConfig struct {
	AccessDeniedHandler AccessDeniedHandler
	EntryPoint          AuthenticationEntryPoint
}

// Option 安全配置选项函数类型。
type Option func(*SecurityConfig)

// NewSecurityConfig 创建默认安全配置。
//
// 返回值:
//   - *SecurityConfig: 带有合理默认值的配置实例
//
// 示例:
//
//	cfg := security.NewSecurityConfig(
//	    security.WithAuthenticationManager(authManager),
//	    security.WithFormLogin("/login", "/dashboard"),
//	    security.WithCsrf(),
//	    security.WithAuthorizeRequests(func(r security.AuthorizeRequests) {
//	        r.AntMatchers("/api/**").HasRole("ROLE_API")
//	        r.AnyRequest().Authenticated()
//	    }),
//	)
func NewSecurityConfig(opts ...Option) *SecurityConfig {
	cfg := &SecurityConfig{
		Filters:             make([]filterConfigEntry, 0),
		CsrfTokenRepository: NewCookieCsrfTokenRepository(),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

// WithAuthenticationManager 设置认证管理器。
func WithAuthenticationManager(manager AuthenticationManager) Option {
	return func(cfg *SecurityConfig) {
		cfg.AuthenticationManager = manager
	}
}

// WithUserDetailsService 设置用户详情服务。
func WithUserDetailsService(service UserDetailsService) Option {
	return func(cfg *SecurityConfig) {
		cfg.UserDetailsService = service
	}
}

// WithPasswordEncoder 设置密码编码器。
func WithPasswordEncoder(encoder PasswordEncoder) Option {
	return func(cfg *SecurityConfig) {
		cfg.PasswordEncoder = encoder
	}
}

// WithAccessDecisionManager 设置访问决策管理器。
func WithAccessDecisionManager(manager AccessDecisionManager) Option {
	return func(cfg *SecurityConfig) {
		cfg.AccessDecisionManager = manager
	}
}

// WithSecurityMetadataSource 设置安全元数据源。
func WithSecurityMetadataSource(source SecurityMetadataSource) Option {
	return func(cfg *SecurityConfig) {
		cfg.SecurityMetadataSource = source
	}
}

// WithFilter 添加安全过滤器。
func WithFilter(filter SecurityFilter) Option {
	return func(cfg *SecurityConfig) {
		cfg.Filters = append(cfg.Filters, filterConfigEntry{Filter: filter})
	}
}

// WithFilterBefore 在指定过滤器之前添加过滤器。
func WithFilterBefore(filter, before SecurityFilter) Option {
	return func(cfg *SecurityConfig) {
		cfg.Filters = append(cfg.Filters, filterConfigEntry{
			Filter: filter,
			Before: before,
		})
	}
}

// WithFilterAfter 在指定过滤器之后添加过滤器。
func WithFilterAfter(filter, after SecurityFilter) Option {
	return func(cfg *SecurityConfig) {
		cfg.Filters = append(cfg.Filters, filterConfigEntry{
			Filter: filter,
			After:  after,
		})
	}
}

// WithAnonymous 启用匿名访问。
func WithAnonymous() Option {
	return func(cfg *SecurityConfig) {
		cfg.Anonymous = true
	}
}

// WithExceptionHandling 配置异常处理。
func WithExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) Option {
	return func(cfg *SecurityConfig) {
		cfg.ExceptionHandling = &exceptionHandlingConfig{
			AccessDeniedHandler: handler,
			EntryPoint:          entryPoint,
		}
	}
}

// WithCsrf 启用 CSRF 防护。
func WithCsrf() Option {
	return func(cfg *SecurityConfig) {
		cfg.CsrfEnabled = true
	}
}

// WithCsrfTokenRepository 设置 CSRF 令牌仓库。
func WithCsrfTokenRepository(repo CsrfTokenRepository) Option {
	return func(cfg *SecurityConfig) {
		cfg.CsrfTokenRepository = repo
	}
}

// WithLogout 配置登出。
func WithLogout(logoutUrl string, successHandler ...LogoutSuccessHandler) Option {
	return func(cfg *SecurityConfig) {
		cfg.Logout = &logoutConfig{
			url: logoutUrl,
		}
		if len(successHandler) > 0 {
			cfg.Logout.successHandler = successHandler[0]
		}
	}
}

// WithFormLogin 配置表单登录。
func WithFormLogin(loginProcessingUrl string, defaultSuccessUrl ...string) Option {
	return func(cfg *SecurityConfig) {
		cfg.FormLogin = &formLoginConfig{
			processingUrl: loginProcessingUrl,
		}
		if len(defaultSuccessUrl) > 0 && defaultSuccessUrl[0] != "" {
			cfg.FormLogin.defaultSuccessUrl = defaultSuccessUrl[0]
		}
	}
}

// WithHttpBasic 启用 HTTP Basic 认证。
func WithHttpBasic(realm ...string) Option {
	return func(cfg *SecurityConfig) {
		cfg.HttpBasic = true
		if len(realm) > 0 {
			cfg.HttpBasicRealm = realm[0]
		} else {
			cfg.HttpBasicRealm = "Secured Area"
		}
	}
}

// WithAuthorizeRequests 配置授权规则。
func WithAuthorizeRequests(config func(authorizer AuthorizeRequests)) Option {
	return func(cfg *SecurityConfig) {
		authorizer := &configAuthorizer{cfg: cfg}
		config(authorizer)
	}
}

// configAuthorizer 函数式选项模式下的授权器实现。
type configAuthorizer struct {
	cfg *SecurityConfig
}

func (a *configAuthorizer) AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry {
	return &configRegistry{cfg: a.cfg, patterns: patterns}
}

func (a *configAuthorizer) AnyRequest() ExpressionInterceptUrlRegistry {
	return &configRegistry{cfg: a.cfg, patterns: []string{"/**"}}
}

// configRegistry 函数式选项模式下的 URL 拦截注册实现。
type configRegistry struct {
	cfg      *SecurityConfig
	patterns []string
}

func (r *configRegistry) addRule(attrs []string) HttpSecurity {
	if len(r.patterns) == 0 || len(attrs) == 0 {
		return nil
	}
	r.cfg.AuthorizeRules = append(r.cfg.AuthorizeRules, authorizeRule{
		patterns: r.patterns,
		attrs:    attrs,
	})
	return nil
}

func (r *configRegistry) PermitAll() HttpSecurity {
	return r.addRule([]string{"permitAll"})
}

func (r *configRegistry) Authenticated() HttpSecurity {
	return r.addRule([]string{"authenticated"})
}

func (r *configRegistry) HasRole(role string) HttpSecurity {
	return r.addRule([]string{"hasRole('" + role + "')"})
}

func (r *configRegistry) HasAnyRole(roles ...string) HttpSecurity {
	return r.addRule([]string{"hasAnyRole('" + joinRoles(roles) + "')"})
}

func (r *configRegistry) HasAuthority(authority string) HttpSecurity {
	return r.addRule([]string{"hasAuthority('" + authority + "')"})
}

func (r *configRegistry) HasAnyAuthority(authorities ...string) HttpSecurity {
	return r.addRule([]string{"hasAnyAuthority('" + joinRoles(authorities) + "')"})
}

func (r *configRegistry) DenyAll() HttpSecurity {
	return r.addRule([]string{"denyAll"})
}

// Build 从 SecurityConfig 构建 SecurityFilterChain。
//
// 这是函数式选项模式的核心构建方法，将配置对象转换为可执行的过滤器链。
//
// 返回值:
//   - SecurityFilterChain: 构建完成的安全过滤器链
//   - error: 构建过程中的错误（如缺少认证管理器）
func (cfg *SecurityConfig) Build() (SecurityFilterChain, error) {
	if cfg.AuthenticationManager == nil {
		return nil, fmt.Errorf("authentication manager is required")
	}

	if cfg.SecurityMetadataSource == nil {
		cfg.SecurityMetadataSource = NewExpressionBasedFilterInvocationSecurityMetadataSource()
	}

	if source, ok := cfg.SecurityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource); ok {
		for _, rule := range cfg.AuthorizeRules {
			for _, pattern := range rule.patterns {
				source.AddMapping(pattern, rule.attrs)
			}
		}
	}

	if cfg.AccessDecisionManager == nil {
		webExpressionVoter := NewWebExpressionVoter()
		authenticatedVoter := NewAuthenticatedVoter()
		roleVoter := NewRoleVoter()
		cfg.AccessDecisionManager = NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)
	}

	authContextFilter := NewAuthContextFilter()

	anonymousFilter := cfg.anonymousFilterOrDefault()

	filterSecurityInterceptor := NewFilterSecurityInterceptor(
		cfg.SecurityMetadataSource,
		cfg.AccessDecisionManager,
		cfg.AuthenticationManager,
	)

	exceptionTranslationFilter := cfg.exceptionTranslationFilterOrDefault()

	defaultFilters := cfg.buildDefaultFilters(
		authContextFilter,
		anonymousFilter,
		exceptionTranslationFilter,
		filterSecurityInterceptor,
	)

	allFilters := cfg.mergeFilters(defaultFilters)
	sort.SliceStable(allFilters, func(i, j int) bool {
		return allFilters[i].Order() < allFilters[j].Order()
	})

	proxy := newFilterChainProxy(allFilters, &DefaultSecurityFilterChain{})
	return &securityFilterChainAdapter{proxy: proxy}, nil
}

func (cfg *SecurityConfig) anonymousFilterOrDefault() *AnonymousAuthenticationFilter {
	if cfg.Anonymous {
		return NewAnonymousAuthenticationFilter()
	}
	return NewAnonymousAuthenticationFilter()
}

func (cfg *SecurityConfig) exceptionTranslationFilterOrDefault() *ExceptionTranslationFilter {
	if cfg.ExceptionHandling != nil {
		return NewExceptionTranslationFilter(
			cfg.ExceptionHandling.AccessDeniedHandler,
			cfg.ExceptionHandling.EntryPoint,
		)
	}
	accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
	unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
	return NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
}

func (cfg *SecurityConfig) buildDefaultFilters(
	authContextFilter *AuthContextFilter,
	anonymousFilter *AnonymousAuthenticationFilter,
	exceptionTranslationFilter *ExceptionTranslationFilter,
	filterSecurityInterceptor *FilterSecurityInterceptor,
) []SecurityFilter {
	defaultFilters := []SecurityFilter{
		authContextFilter,
		anonymousFilter,
	}

	if cfg.CsrfEnabled {
		csrfFilter, err := NewCsrfFilter(cfg.CsrfTokenRepository)
		if err != nil {
			// 在构建时处理错误，这里使用默认实现不会失败
			// 实际错误会在 NewCsrfFilter 中返回
			return defaultFilters
		}
		defaultFilters = append(defaultFilters, csrfFilter)
	}

	if cfg.Logout != nil && cfg.Logout.url != "" {
		logoutFilter, err := NewLogoutFilter(cfg.Logout.url, nil)
		if err == nil {
			if cfg.Logout.successHandler != nil {
				logoutFilter.SetSuccessHandler(cfg.Logout.successHandler)
			}
			defaultFilters = append(defaultFilters, logoutFilter)
		}
	}

	if cfg.FormLogin != nil {
		processingUrl := cfg.FormLogin.processingUrl
		defaultSuccessUrl := cfg.FormLogin.defaultSuccessUrl
		if defaultSuccessUrl == "" {
			defaultSuccessUrl = "/"
		}
		failureUrl := "/login?error"

		formLoginFilter := NewUsernamePasswordAuthenticationFilterWithDefaults(
			processingUrl,
			defaultSuccessUrl,
			failureUrl,
			cfg.AuthenticationManager,
			log.Build(),
		)
		defaultFilters = append(defaultFilters, formLoginFilter)
	}

	if cfg.HttpBasic {
		basicFilter := NewBasicAuthenticationFilterWithRealm(
			cfg.AuthenticationManager,
			cfg.HttpBasicRealm,
			log.Build(),
		)
		defaultFilters = append(defaultFilters, basicFilter)
	}

	defaultFilters = append(defaultFilters, exceptionTranslationFilter)
	defaultFilters = append(defaultFilters, filterSecurityInterceptor)

	return defaultFilters
}

func (cfg *SecurityConfig) mergeFilters(defaultFilters []SecurityFilter) []SecurityFilter {
	allFilters := make([]SecurityFilter, 0, len(defaultFilters)+len(cfg.Filters))
	allFilters = append(allFilters, defaultFilters...)

	for _, entry := range cfg.Filters {
		if entry.Before != nil {
			allFilters = insertFilterBefore(allFilters, entry.Filter, entry.Before)
		} else if entry.After != nil {
			allFilters = insertFilterAfter(allFilters, entry.Filter, entry.After)
		} else {
			allFilters = append(allFilters, entry.Filter)
		}
	}

	return allFilters
}

func insertFilterBefore(filters []SecurityFilter, filter, before SecurityFilter) []SecurityFilter {
	newFilters := make([]SecurityFilter, 0, len(filters)+1)
	inserted := false
	for _, f := range filters {
		if f == before && !inserted {
			newFilters = append(newFilters, filter)
			inserted = true
		}
		newFilters = append(newFilters, f)
	}
	if !inserted {
		newFilters = append(newFilters, filter)
	}
	return newFilters
}

func insertFilterAfter(filters []SecurityFilter, filter, after SecurityFilter) []SecurityFilter {
	newFilters := make([]SecurityFilter, 0, len(filters)+1)
	inserted := false
	for _, f := range filters {
		newFilters = append(newFilters, f)
		if f == after && !inserted {
			newFilters = append(newFilters, filter)
			inserted = true
		}
	}
	if !inserted {
		newFilters = append(newFilters, filter)
	}
	return newFilters
}

// joinRoles 将角色列表拼接为表达式格式。
func joinRoles(roles []string) string {
	var sb strings.Builder
	for i, role := range roles {
		if i > 0 {
			sb.WriteString("','")
		}
		sb.WriteString(role)
	}
	return sb.String()
}

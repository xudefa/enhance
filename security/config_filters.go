package security

import (
	"fmt"
	"sort"
	"strings"

	"github.com/xudefa/enhance/log"
)

// configAuthorizer 函数式选项模式下的授权器实现。
type configAuthorizer struct {
	cfg *SecurityConfig
}

// AntMatchers 为指定路径模式注册授权规则。
func (a *configAuthorizer) AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry {
	return &configRegistry{cfg: a.cfg, patterns: patterns}
}

// AnyRequest 为所有请求路径注册授权规则。
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

// PermitAll 放行所有请求。
func (r *configRegistry) PermitAll() HttpSecurity {
	return r.addRule([]string{"permitAll"})
}

// Authenticated 要求请求已认证。
func (r *configRegistry) Authenticated() HttpSecurity {
	return r.addRule([]string{"authenticated"})
}

// HasRole 要求用户拥有指定角色。
func (r *configRegistry) HasRole(role string) HttpSecurity {
	return r.addRule([]string{"hasRole('" + role + "')"})
}

// HasAnyRole 要求用户拥有任一指定角色。
func (r *configRegistry) HasAnyRole(roles ...string) HttpSecurity {
	return r.addRule([]string{"hasAnyRole('" + joinRoles(roles) + "')"})
}

// HasAuthority 要求用户拥有指定权限。
func (r *configRegistry) HasAuthority(authority string) HttpSecurity {
	return r.addRule([]string{"hasAuthority('" + authority + "')"})
}

// HasAnyAuthority 要求用户拥有任一指定权限。
func (r *configRegistry) HasAnyAuthority(authorities ...string) HttpSecurity {
	return r.addRule([]string{"hasAnyAuthority('" + joinRoles(authorities) + "')"})
}

// DenyAll 拒绝所有请求。
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

	cfg.applyAuthorizeRules()

	if cfg.AccessDecisionManager == nil {
		cfg.AccessDecisionManager = newDefaultAccessDecisionManager()
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

// applyAuthorizeRules 将授权规则写入安全元数据源。
func (cfg *SecurityConfig) applyAuthorizeRules() {
	source, ok := cfg.SecurityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource)
	if !ok {
		return
	}
	for _, rule := range cfg.AuthorizeRules {
		for _, pattern := range rule.patterns {
			source.AddMapping(pattern, rule.attrs)
		}
	}
}

// newDefaultAccessDecisionManager 创建使用默认投票者的访问决策管理器。
func newDefaultAccessDecisionManager() AccessDecisionManager {
	webExpressionVoter := NewWebExpressionVoter()
	authenticatedVoter := NewAuthenticatedVoter()
	roleVoter := NewRoleVoter()
	return NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)
}

// newDefaultExceptionTranslationFilter 创建使用默认 403/401 处理的异常转换过滤器。
func newDefaultExceptionTranslationFilter() *ExceptionTranslationFilter {
	accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
	unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
	return NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
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
		defaultFilters = append(defaultFilters, cfg.formLoginFilterOrDefault())
	}

	if cfg.HttpBasic {
		defaultFilters = append(defaultFilters, cfg.basicFilterOrDefault())
	}

	defaultFilters = append(defaultFilters, exceptionTranslationFilter)
	defaultFilters = append(defaultFilters, filterSecurityInterceptor)

	return defaultFilters
}

// formLoginFilterOrDefault 根据配置构建表单登录过滤器。
func (cfg *SecurityConfig) formLoginFilterOrDefault() *UsernamePasswordAuthenticationFilter {
	processingUrl := cfg.FormLogin.processingUrl
	defaultSuccessUrl := cfg.FormLogin.defaultSuccessUrl
	if defaultSuccessUrl == "" {
		defaultSuccessUrl = "/"
	}
	failureUrl := "/login?error"

	return NewUsernamePasswordAuthenticationFilterWithDefaults(
		processingUrl,
		cfg.AuthenticationManager,
		log.Build(),
		WithDefaultSuccessURL(defaultSuccessUrl),
		WithFailureURL(failureUrl),
	)
}

// basicFilterOrDefault 根据配置构建 Basic 认证过滤器。
func (cfg *SecurityConfig) basicFilterOrDefault() *BasicAuthenticationFilterWithRealm {
	return NewBasicAuthenticationFilterWithRealm(
		cfg.AuthenticationManager,
		cfg.HttpBasicRealm,
		log.Build(),
	)
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

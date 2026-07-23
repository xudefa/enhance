package security

import (
	"fmt"
	"strings"

	"github.com/xudefa/enhance/log"
)

// httpSecurity HTTP安全配置实现
type httpSecurity struct {
	authenticationManager       AuthenticationManager
	userDetailsService          UserDetailsService
	passwordEncoder             PasswordEncoder
	accessDecisionManager       AccessDecisionManager
	securityMetadataSource      SecurityMetadataSource
	filters                     []SecurityFilter
	anonymousFilter             *AnonymousAuthenticationFilter
	exceptionTranslationFilter  *ExceptionTranslationFilter
	filterSecurityInterceptor   *FilterSecurityInterceptor
	securityContextHolderFilter *SecurityContextHolderFilter

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
}

func NewHttpSecurity() HttpSecurity {
	return &httpSecurity{
		filters:             make([]SecurityFilter, 0),
		csrfTokenRepository: NewCookieCsrfTokenRepository(),
	}
}

func (h *httpSecurity) AuthenticationManager(authManager AuthenticationManager) HttpSecurity {
	h.authenticationManager = authManager
	return h
}

func (h *httpSecurity) UserDetailsService(userDetailsService UserDetailsService) HttpSecurity {
	h.userDetailsService = userDetailsService
	return h
}

func (h *httpSecurity) PasswordEncoder(encoder PasswordEncoder) HttpSecurity {
	h.passwordEncoder = encoder
	return h
}

func (h *httpSecurity) AccessDecisionManager(manager AccessDecisionManager) HttpSecurity {
	h.accessDecisionManager = manager
	return h
}

func (h *httpSecurity) SecurityMetadataSource(source SecurityMetadataSource) HttpSecurity {
	h.securityMetadataSource = source
	return h
}

func (h *httpSecurity) AuthorizeRequests(authorizer AuthorizeRequests) HttpSecurity {
	if httpSecurityAuthorizer, ok := authorizer.(*httpSecurityAuthorizer); ok {
		httpSecurityAuthorizer.httpSecurity = h
	}
	return h
}

func (h *httpSecurity) AddFilter(filter SecurityFilter) HttpSecurity {
	h.filters = append(h.filters, filter)
	return h
}

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

func (h *httpSecurity) Anonymous() HttpSecurity {
	h.anonymousFilter = NewAnonymousAuthenticationFilter()
	return h
}

func (h *httpSecurity) ExceptionHandling(handler AccessDeniedHandler, entryPoint AuthenticationEntryPoint) HttpSecurity {
	h.exceptionTranslationFilter = NewExceptionTranslationFilter(handler, entryPoint)
	return h
}

func (h *httpSecurity) Csrf() HttpSecurity {
	h.csrfEnabled = true
	return h
}

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

	if h.accessDecisionManager == nil {
		webExpressionVoter := NewWebExpressionVoter()
		authenticatedVoter := NewAuthenticatedVoter()
		roleVoter := NewRoleVoter()
		h.accessDecisionManager = NewAffirmativeBased(webExpressionVoter, authenticatedVoter, roleVoter)
	}

	h.securityContextHolderFilter = NewSecurityContextHolderFilter(GetSecurityContext())

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
		h.securityContextHolderFilter,
		h.anonymousFilter,
	}

	if h.csrfEnabled {
		csrfFilter := NewCsrfFilter(h.csrfTokenRepository)
		defaultFilters = append(defaultFilters, csrfFilter)
	}

	if h.logoutUrl != "" {
		logoutFilter := NewLogoutFilter(h.logoutUrl, h.logoutHandlers)
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

	proxy := NewFilterChainProxy(allFilters, &DefaultSecurityFilterChain{})
	return &securityFilterChainAdapter{proxy: proxy}, nil
}

// DefaultSecurityFilterChain 默认安全过滤器链
type DefaultSecurityFilterChain struct{}

func (c *DefaultSecurityFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	return nil
}

func (c *DefaultSecurityFilterChain) Matches(request interface{}) bool {
	return true
}

func (c *DefaultSecurityFilterChain) GetFilters() []SecurityFilter {
	return nil
}

// httpSecurityAuthorizer HTTP安全授权器
type httpSecurityAuthorizer struct {
	httpSecurity *httpSecurity
}

func (a *httpSecurityAuthorizer) AntMatchers(patterns ...string) ExpressionInterceptUrlRegistry {
	return &expressionInterceptUrlRegistry{
		httpSecurity: a.httpSecurity,
		patterns:     patterns,
	}
}

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

func (r *expressionInterceptUrlRegistry) PermitAll() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"permitAll"},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) Authenticated() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"authenticated"},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) HasRole(role string) HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasRole('%s')", role)},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) HasAnyRole(roles ...string) HttpSecurity {
	for _, pattern := range r.patterns {
		var sb strings.Builder
		for i, role := range roles {
			if i > 0 {
				sb.WriteString("','")
			}
			sb.WriteString(role)
		}
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAnyRole('%s')", sb.String())},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) HasAuthority(authority string) HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAuthority('%s')", authority)},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) HasAnyAuthority(authorities ...string) HttpSecurity {
	for _, pattern := range r.patterns {
		var sb strings.Builder
		for i, auth := range authorities {
			if i > 0 {
				sb.WriteString("','")
			}
			sb.WriteString(auth)
		}
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{fmt.Sprintf("hasAnyAuthority('%s')", sb.String())},
		)
	}
	return r.httpSecurity
}

func (r *expressionInterceptUrlRegistry) DenyAll() HttpSecurity {
	for _, pattern := range r.patterns {
		r.httpSecurity.securityMetadataSource.(*ExpressionBasedFilterInvocationSecurityMetadataSource).AddMapping(
			pattern,
			[]string{"denyAll"},
		)
	}
	return r.httpSecurity
}

// WebSecurity Web安全配置入口
type WebSecurity struct{}

func NewWebSecurity() *WebSecurity {
	return &WebSecurity{}
}

func (w *WebSecurity) HttpSecurity() *httpSecurity {
	return &httpSecurity{
		filters: make([]SecurityFilter, 0),
	}
}

func (w *WebSecurity) Build() (SecurityFilterChain, error) {
	return nil, fmt.Errorf("use HttpSecurity().Build() instead")
}

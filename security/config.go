package security

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

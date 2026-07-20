package security

// SecurityBuilder 安全配置构建器
type SecurityBuilder struct {
	authManager        AuthenticationManager
	userDetailsService UserDetailsService
	passwordEncoder    PasswordEncoder
	accessDecisionMgr  AccessDecisionManager
	filters            []filterEntry
	anonymous          bool
	csrf               bool
	formLogin          *formLoginConfig
	httpBasic          bool
	logoutConfig       *logoutConfig
}

type filterEntry struct {
	filter SecurityFilter
	before SecurityFilter
	after  SecurityFilter
}

type formLoginConfig struct {
	processingUrl     string
	defaultSuccessUrl string
}

type logoutConfig struct {
	url            string
	successHandler LogoutSuccessHandler
}

// NewSecurityBuilder 创建安全配置构建器
func NewSecurityBuilder() *SecurityBuilder {
	return &SecurityBuilder{}
}

// AuthenticationManager 设置认证管理器
func (b *SecurityBuilder) AuthenticationManager(manager AuthenticationManager) *SecurityBuilder {
	b.authManager = manager
	return b
}

// UserDetailsService 设置用户详情服务
func (b *SecurityBuilder) UserDetailsService(service UserDetailsService) *SecurityBuilder {
	b.userDetailsService = service
	return b
}

// PasswordEncoder 设置密码编码器
func (b *SecurityBuilder) PasswordEncoder(encoder PasswordEncoder) *SecurityBuilder {
	b.passwordEncoder = encoder
	return b
}

// AccessDecisionManager 设置访问决策管理器
func (b *SecurityBuilder) AccessDecisionManager(manager AccessDecisionManager) *SecurityBuilder {
	b.accessDecisionMgr = manager
	return b
}

// AddFilter 添加过滤器
func (b *SecurityBuilder) AddFilter(filter SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter})
	return b
}

// AddFilterBefore 在指定过滤器前添加
func (b *SecurityBuilder) AddFilterBefore(filter SecurityFilter, before SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter, before: before})
	return b
}

// AddFilterAfter 在指定过滤器后添加
func (b *SecurityBuilder) AddFilterAfter(filter SecurityFilter, after SecurityFilter) *SecurityBuilder {
	b.filters = append(b.filters, filterEntry{filter: filter, after: after})
	return b
}

// EnableAnonymous 启用匿名访问
func (b *SecurityBuilder) EnableAnonymous() *SecurityBuilder {
	b.anonymous = true
	return b
}

// EnableCsrf 启用 CSRF 保护
func (b *SecurityBuilder) EnableCsrf() *SecurityBuilder {
	b.csrf = true
	return b
}

// EnableFormLogin 启用表单登录
func (b *SecurityBuilder) EnableFormLogin(processingUrl string, defaultSuccessUrl ...string) *SecurityBuilder {
	cfg := &formLoginConfig{processingUrl: processingUrl}
	if len(defaultSuccessUrl) > 0 {
		cfg.defaultSuccessUrl = defaultSuccessUrl[0]
	}
	b.formLogin = cfg
	return b
}

// EnableHttpBasic 启用 HTTP Basic 认证
func (b *SecurityBuilder) EnableHttpBasic() *SecurityBuilder {
	b.httpBasic = true
	return b
}

// EnableLogout 启用登出
func (b *SecurityBuilder) EnableLogout(url string, successHandler ...LogoutSuccessHandler) *SecurityBuilder {
	cfg := &logoutConfig{url: url}
	if len(successHandler) > 0 {
		cfg.successHandler = successHandler[0]
	}
	b.logoutConfig = cfg
	return b
}

// Build 构建安全配置
func (b *SecurityBuilder) Build() SecurityConfig {
	return &builtSecurityConfig{builder: b}
}

type builtSecurityConfig struct {
	builder *SecurityBuilder
}

func (c *builtSecurityConfig) Configure(http HttpSecurity) error {
	if c.builder.authManager != nil {
		http.AuthenticationManager(c.builder.authManager)
	}
	if c.builder.userDetailsService != nil {
		http.UserDetailsService(c.builder.userDetailsService)
	}
	if c.builder.passwordEncoder != nil {
		http.PasswordEncoder(c.builder.passwordEncoder)
	}
	if c.builder.accessDecisionMgr != nil {
		http.AccessDecisionManager(c.builder.accessDecisionMgr)
	}

	for _, entry := range c.builder.filters {
		if entry.before != nil {
			http.AddFilterBefore(entry.filter, entry.before)
			continue
		}
		if entry.after != nil {
			http.AddFilterAfter(entry.filter, entry.after)
			continue
		}
		http.AddFilter(entry.filter)
	}

	if c.builder.anonymous {
		http.Anonymous()
	}
	if c.builder.csrf {
		http.Csrf()
	}
	if c.builder.formLogin != nil {
		if c.builder.formLogin.defaultSuccessUrl == "" {
			http.FormLogin(c.builder.formLogin.processingUrl)
		} else {
			http.FormLogin(c.builder.formLogin.processingUrl, c.builder.formLogin.defaultSuccessUrl)
		}
	}
	if c.builder.httpBasic {
		http.HttpBasic()
	}
	if c.builder.logoutConfig != nil {
		if c.builder.logoutConfig.successHandler == nil {
			http.Logout(c.builder.logoutConfig.url)
		} else {
			http.Logout(c.builder.logoutConfig.url, c.builder.logoutConfig.successHandler)
		}
	}

	return nil
}

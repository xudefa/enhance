package security

import (
	"reflect"
	"strings"

	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/log"
)

func init() {
	boot.RegisterAutoConfigWith(&SecurityAutoConfiguration{},
		boot.WithConditions(
			condition.OnProperty(SecurityEnabled, ConditionTrue),
		),
		boot.WithOrder(int(boot.OrderPrioritySecurityCore)), // 安全核心层，在认证和授权之后执行
	)
}

// SecurityAutoConfiguration 安全模块自动配置
// 当 security.enabled=true 时自动装配以下组件：
// - UserDetailsService: 默认使用内存存储（可自定义）
// - PasswordEncoder: 默认使用NoOp编码器（生产环境应使用BCrypt）
// - AuthenticationManager: 包含DaoAuthenticationProvider和AnonymousAuthenticationProvider
// - SecurityFilterChain: 默认安全过滤器链（支持CORS、限流、JWT等）
//
// 自动配置流程：
// 1. 检查容器是否已有UserDetailsService，没有则创建默认的InMemoryUserDetailsService
// 2. 检查容器是否已有PasswordEncoder，没有则创建NoOpPasswordEncoder
// 3. 构建AuthenticationManager（包含DAO认证提供者和匿名认证提供者）
// 4. 构建SecurityFilterChain（包含CORS、限流、认证、授权等过滤器）
// 5. 将组件注册到容器中
type SecurityAutoConfiguration struct {
	logger log.Logger
}

// Configure 执行自动配置
func (c *SecurityAutoConfiguration) Configure(ctx boot.ApplicationContext) error {
	container := ctx.Container()
	env := ctx.Environment()

	// 从容器获取日志记录器，如果不存在则使用默认值
	if logger, err := core.GetByName[log.Logger](container, ""); err == nil {
		c.logger = logger
	} else {
		c.logger = log.Build()
	}
	c.logger.Info(ctx.Context(), "开始配置安全模块...")

	// 获取或注册 UserDetailsService
	userDetailsService := c.getUserDetailsService(ctx, container)

	// 获取或注册 PasswordEncoder
	passwordEncoder := c.getPasswordEncoder(ctx, container)

	// 注册 AuthenticationManager
	authManager := c.buildAuthenticationManager(ctx, userDetailsService, passwordEncoder)
	// 使用接口类型注册，确保可以通过接口类型获取
	if regErr := ctx.Container().RegisterInstance(authManager, reflect.TypeOf(authManager)); regErr != nil {
		c.logger.Error(ctx.Context(), "注册 AuthenticationManager 失败", log.KeyValue{Key: "error", Value: regErr.Error()})
		return regErr
	}
	c.logger.Info(ctx.Context(), "AuthenticationManager 已注册")

	filterChain := c.getOrBuildSecurityFilterChain(ctx, authManager, userDetailsService, env, container)

	// 使用接口类型注册，确保可以通过接口类型获取
	if regErr := ctx.Container().RegisterInstance(filterChain, reflect.TypeOf(filterChain)); regErr != nil {
		c.logger.Error(ctx.Context(), "注册 SecurityFilterChain 失败", log.KeyValue{Key: "error", Value: regErr.Error()})
		return regErr
	}
	// 同时注册接口类型，确保可以通过 security.SecurityFilterChain 接口查找
	if regErr := ctx.Container().RegisterInstance(filterChain, reflect.TypeFor[SecurityFilterChain]()); regErr != nil {
		c.logger.Warn(ctx.Context(), "注册 SecurityFilterChain 接口类型失败（非致命）", log.KeyValue{Key: "error", Value: regErr.Error()})
	}
	c.logger.Info(ctx.Context(), "SecurityFilterChain 已注册")

	c.logger.Info(ctx.Context(), "安全模块配置完成")
	return nil
}

func (c *SecurityAutoConfiguration) getUserDetailsService(ctx boot.ApplicationContext, container core.Container) UserDetailsService {
	c.logger.Debug(ctx.Context(), "尝试获取 UserDetailsService...")
	beans, err := container.Get(reflect.TypeOf((*UserDetailsService)(nil)).Elem())
	if err != nil || len(beans) == 0 {
		c.logger.Debug(ctx.Context(), "未找到 UserDetailsService，创建默认实例")
		service := NewInMemoryUserDetailsService()
		if regErr := ctx.Container().RegisterInstance(service, reflect.TypeOf(service)); regErr != nil {
			c.logger.Error(ctx.Context(), "注册 UserDetailsService 失败", log.KeyValue{Key: "error", Value: regErr.Error()})
		}
		return service
	}
	c.logger.Debug(ctx.Context(), "找到已注册的 UserDetailsService", log.KeyValue{Key: "count", Value: len(beans)})
	service, ok := beans[0].(UserDetailsService)
	if !ok {
		c.logger.Error(ctx.Context(), "beans[0] is not UserDetailsService")
		return nil
	}
	return service
}

func (c *SecurityAutoConfiguration) getPasswordEncoder(ctx boot.ApplicationContext, container core.Container) PasswordEncoder {
	beans, err := container.Get(reflect.TypeOf((*PasswordEncoder)(nil)).Elem())
	if err != nil || len(beans) == 0 {
		// 使用 NoOp 作为默认编码器
		// 生产环境应通过 starter/bcrypt 模块注入真实的 bcrypt 编码器
		c.logger.Debug(ctx.Context(), "未找到 PasswordEncoder，创建 NoOp 编码器")
		encoder := NewNoOpPasswordEncoder()
		if regErr := ctx.Container().RegisterInstance(encoder, reflect.TypeOf(encoder)); regErr != nil {
			c.logger.Error(ctx.Context(), "注册 PasswordEncoder 失败", log.KeyValue{Key: "error", Value: regErr.Error()})
		}
		return encoder
	}
	c.logger.Debug(ctx.Context(), "找到已注册的 PasswordEncoder")
	encoder, ok := beans[0].(PasswordEncoder)
	if !ok {
		c.logger.Error(ctx.Context(), "beans[0] is not PasswordEncoder")
		return NewNoOpPasswordEncoder()
	}
	return encoder
}

// buildAuthenticationManager 构建 AuthenticationManager
func (c *SecurityAutoConfiguration) buildAuthenticationManager(ctx boot.ApplicationContext, userDetailsService UserDetailsService, passwordEncoder PasswordEncoder) AuthenticationManager {
	container := ctx.Container()
	var logger log.Logger
	if l, err := core.GetByName[log.Logger](container, ""); err == nil {
		logger = l
	} else {
		logger = log.Build()
	}
	daoAuthProvider := NewDaoAuthenticationProvider(userDetailsService, passwordEncoder, logger)
	anonymousProvider := NewAnonymousAuthenticationProvider()
	return NewProviderManager(daoAuthProvider, anonymousProvider)
}

// getOrBuildSecurityFilterChain 获取或构建 SecurityFilterChain
func (c *SecurityAutoConfiguration) getOrBuildSecurityFilterChain(ctx boot.ApplicationContext, authManager AuthenticationManager, userDetailsService UserDetailsService, env *environment.Environment, container core.Container) SecurityFilterChain {
	beans, err := container.Get(reflect.TypeOf((*SecurityFilterChain)(nil)).Elem())
	if err != nil || len(beans) == 0 {
		return c.buildSecurityFilterChain(authManager, userDetailsService, env, container)
	}
	// 如果用户注入的过滤器链中不存在CORS、限流过滤器，则添加默认的
	chain, ok := beans[0].(SecurityFilterChain)
	if !ok {
		c.logger.Error(ctx.Context(), "beans[0] is not SecurityFilterChain")
		return c.buildSecurityFilterChain(authManager, userDetailsService, env, container)
	}
	return c.addDefaultFiltersIfMissing(chain, env, container)
}

// buildSecurityFilterChain 构建安全过滤器链
func (c *SecurityAutoConfiguration) buildSecurityFilterChain(authManager AuthenticationManager, userDetailsService UserDetailsService, env *environment.Environment, container core.Container) SecurityFilterChain {
	filters := []SecurityFilter{}

	// 添加CORS过滤器
	if corsFilter := c.getOrCreateCorsFilter(env, container); corsFilter != nil {
		filters = append(filters, corsFilter)
	}

	// 添加限流过滤器
	if rateLimitFilter := c.getOrCreateRateLimitFilter(env, container); rateLimitFilter != nil {
		filters = append(filters, rateLimitFilter)
	}

	// 添加认证上下文过滤器（必须在最前面，保存初始状态）
	contextHolderFilter := NewAuthContextFilter()
	filters = append(filters, contextHolderFilter)

	// 添加JWT认证过滤器（如果容器中存在）
	if jwtFilter := c.getJwtAuthenticationFilter(container); jwtFilter != nil {
		filters = append(filters, jwtFilter)
	}

	// 添加其他默认安全过滤器
	defaultFilters := c.buildDefaultSecurityFilters(authManager, env, container, true) // skipContextFilter=true
	filters = append(filters, defaultFilters...)

	proxy := newFilterChainProxy(filters, &DefaultSecurityFilterChain{})
	return &securityFilterChainAdapter{proxy: proxy}
}

// getJwtAuthenticationFilter 从容器中获取JWT认证过滤器
func (c *SecurityAutoConfiguration) getJwtAuthenticationFilter(container core.Container) SecurityFilter {
	// 使用类型断言查找JWT过滤器
	beans, err := container.Get(reflect.TypeOf((*SecurityFilter)(nil)).Elem())
	if err != nil || len(beans) == 0 {
		return nil
	}
	for _, bean := range beans {
		if jwtFilter, ok := bean.(SecurityFilter); ok {
			// 检查是否是JWT过滤器（通过类型断言）
			if _, isJwt := jwtFilter.(interface{ IsJwtFilter() bool }); isJwt {
				return jwtFilter
			}
		}
	}
	return nil
}

// buildDefaultSecurityFilters 构建默认安全过滤器（不含CORS和限流）
func (c *SecurityAutoConfiguration) buildDefaultSecurityFilters(authManager AuthenticationManager, env *environment.Environment, container core.Container, skipContextFilter ...bool) []SecurityFilter {
	filters := []SecurityFilter{}

	skip := false
	if len(skipContextFilter) > 0 {
		skip = skipContextFilter[0]
	}

	if !skip {
		contextHolderFilter := NewAuthContextFilter()
		filters = append(filters, contextHolderFilter)
	}

	// 添加 Basic 认证过滤器
	basicAuthFilter := NewBasicAuthenticationFilter(authManager)
	filters = append(filters, basicAuthFilter)

	anonymousFilter := NewAnonymousAuthenticationFilter()
	filters = append(filters, anonymousFilter)

	// 收集所有投票者（包括容器中的自定义投票者，如 CasbinVoter）
	voters := []AccessDecisionVoter{
		NewWebExpressionVoter(),
		NewAuthenticatedVoter(),
		NewRoleVoter(),
	}

	// 从容器中查找额外的投票者
	if container != nil {
		if voterBeans, err := container.Get(reflect.TypeOf((*AccessDecisionVoter)(nil)).Elem()); err == nil {
			for _, voterBean := range voterBeans {
				if voter, ok := voterBean.(AccessDecisionVoter); ok {
					// 检查是否已经是默认投票者（通过类型判断）
					isDefault := false
					switch voter.(type) {
					case *WebExpressionVoter, *AuthenticatedVoter, *RoleVoter:
						isDefault = true
					}
					if !isDefault {
						voters = append(voters, voter)
					}
				}
			}
		}
	}

	accessDecisionManager := NewAffirmativeBased(voters...)

	metadataSource := NewExpressionBasedFilterInvocationSecurityMetadataSource()
	if rules := env.GetString("security.rules", ""); rules != "" {
		for _, rule := range parseStringSlice(rules) {
			parts := strings.SplitN(rule, "->", 2)
			if len(parts) == 2 {
				pattern := strings.TrimSpace(parts[0])
				expression := strings.TrimSpace(parts[1])
				metadataSource.AddMapping(pattern, []string{expression})
			}
		}
	}

	accessDeniedHandler := NewHttp403ForbiddenAccessDeniedHandler()
	unauthorizedEntryPoint := NewHttp401UnauthorizedEntryPoint()
	exceptionTranslationFilter := NewExceptionTranslationFilter(accessDeniedHandler, unauthorizedEntryPoint)
	filters = append(filters, exceptionTranslationFilter)

	filterSecurityInterceptor := NewFilterSecurityInterceptor(
		metadataSource,
		accessDecisionManager,
		authManager,
	)
	filters = append(filters, filterSecurityInterceptor)

	return filters
}

// getOrCreateCorsFilter 获取或创建CORS过滤器
func (c *SecurityAutoConfiguration) getOrCreateCorsFilter(env *environment.Environment, container core.Container) *CorsFilter {
	if !env.GetBool("security.cors.enabled", false) {
		return nil
	}
	// 优先从容器获取
	beans, _ := container.Get(reflect.TypeOf((*CorsFilter)(nil)).Elem())
	if len(beans) > 0 {
		if corsFilter, ok := beans[0].(*CorsFilter); ok {
			return corsFilter
		}
	}
	// 创建默认的CORS过滤器
	return c.createDefaultCorsFilter(env)
}

// createDefaultCorsFilter 创建默认的CORS过滤器
func (c *SecurityAutoConfiguration) createDefaultCorsFilter(env *environment.Environment) *CorsFilter {
	corsConfig := CorsConfig{
		AllowedOrigins:   parseStringSlice(env.GetString("security.cors.allowed-origins", "*")),
		AllowedMethods:   parseStringSlice(env.GetString("security.cors.allowed-methods", "GET,POST,PUT,DELETE,OPTIONS")),
		AllowedHeaders:   parseStringSlice(env.GetString("security.cors.allowed-headers", "Content-Type,Authorization,X-Requested-With")),
		ExposedHeaders:   parseStringSlice(env.GetString("security.cors.exposed-headers", "")),
		AllowCredentials: env.GetBool("security.cors.allow-credentials", false),
		MaxAge:           env.GetInt("security.cors.max-age", 3600),
	}
	return NewCorsFilter(corsConfig)
}

// getOrCreateRateLimitFilter 获取或创建限流过滤器
func (c *SecurityAutoConfiguration) getOrCreateRateLimitFilter(env *environment.Environment, container core.Container) *RateLimitFilter {
	if !env.GetBool("security.rate-limit.enabled", false) {
		return nil
	}
	// 优先从容器获取
	beans, _ := container.Get(reflect.TypeOf((*RateLimitFilter)(nil)).Elem())
	if len(beans) > 0 {
		if rateLimitFilter, ok := beans[0].(*RateLimitFilter); ok {
			return rateLimitFilter
		}
	}
	// 创建默认的限流过滤器
	return c.createDefaultRateLimitFilter(env, container)
}

// createDefaultRateLimitFilter 创建默认的限流过滤器
func (c *SecurityAutoConfiguration) createDefaultRateLimitFilter(env *environment.Environment, container core.Container) *RateLimitFilter {
	var logger log.Logger
	if l, err := core.GetByName[log.Logger](container, ""); err == nil {
		logger = l
	} else {
		logger = log.Build()
	}
	rateLimitConfig := RateLimitConfig{
		Enabled:           true,
		Rate:              env.GetInt("security.rate-limit.rate", 100),
		Burst:             env.GetInt("security.rate-limit.burst", 200),
		ExcludePaths:      parseStringSlice(env.GetString("security.rate-limit.exclude-paths", "/health,/actuator/health")),
		TrustProxyHeaders: env.GetBool("security.rate-limit.trust-proxy-headers", false),
		TrustedProxies:    parseStringSlice(env.GetString("security.rate-limit.trusted-proxies", "")),
		Log:               logger,
	}
	return NewRateLimitFilter(rateLimitConfig)
}

// addDefaultFiltersIfMissing 如果用户注入的过滤器链中不存在CORS、限流过滤器，则添加默认的
func (c *SecurityAutoConfiguration) addDefaultFiltersIfMissing(filterChain SecurityFilterChain, env *environment.Environment, container core.Container) SecurityFilterChain {
	// 尝试将用户的过滤器链转换为 securityFilterChainAdapter
	adapter, ok := filterChain.(*securityFilterChainAdapter)
	if !ok {
		return filterChain
	}

	// 检查是否已存在CORS过滤器
	hasCors := c.hasFilter(adapter.proxy.filters, (*CorsFilter)(nil))

	// 检查是否已存在限流过滤器
	hasRateLimit := c.hasFilter(adapter.proxy.filters, (*RateLimitFilter)(nil))

	// 如果不存在CORS过滤器且配置启用了CORS，则添加默认的
	if !hasCors && env.GetBool("security.cors.enabled", false) {
		corsFilter := c.getOrCreateCorsFilter(env, container)
		if corsFilter != nil {
			// 在最前面插入CORS过滤器
			adapter.proxy.filters = append([]SecurityFilter{corsFilter}, adapter.proxy.filters...)
		}
	}

	// 如果不存在限流过滤器且配置启用了限流，则添加默认的
	if !hasRateLimit && env.GetBool("security.rate-limit.enabled", false) {
		rateLimitFilter := c.getOrCreateRateLimitFilter(env, container)
		if rateLimitFilter != nil {
			// 在CORS过滤器之后（如果存在）或最前面插入限流过滤器
			insertIndex := 0
			if hasCors || (!hasCors && env.GetBool("security.cors.enabled", false)) {
				insertIndex = 1
			}
			adapter.proxy.filters = append(adapter.proxy.filters[:insertIndex], append([]SecurityFilter{rateLimitFilter}, adapter.proxy.filters[insertIndex:]...)...)
		}
	}

	return adapter
}

// hasFilter 检查过滤器列表中是否包含指定类型的过滤器
func (c *SecurityAutoConfiguration) hasFilter(filters []SecurityFilter, target any) bool {
	for _, f := range filters {
		switch target.(type) {
		case *CorsFilter:
			if _, ok := f.(*CorsFilter); ok {
				return true
			}
		case *RateLimitFilter:
			if _, ok := f.(*RateLimitFilter); ok {
				return true
			}
		}
	}
	return false
}

func parseStringSlice(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

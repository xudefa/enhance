package authorization

// expressionBasedUrlRegistry 表达式拦截 URL 注册表实现。
//
// 管理 URL 授权规则的注册和构建。
// 支持链式调用配置多个规则。
type expressionBasedUrlRegistry struct {
	rules   []UrlAuthorizationRule
	current *currentRule
}

// currentRule 当前正在构建的规则。
type currentRule struct {
	patterns []string
	attrs    []string
}

// NewExpressionBasedUrlRegistry 创建表达式拦截 URL 注册表。
func NewExpressionBasedUrlRegistry() ExpressionInterceptUrlRegistry {
	return &registryWrapper{
		registry: &expressionBasedUrlRegistry{
			rules: make([]UrlAuthorizationRule, 0),
		},
	}
}

// HasAnyAuthority 要求指定权限（传入单个或多个）。
func (r *expressionBasedUrlRegistry) HasAnyAuthority(authorities ...string) UrlAuthorizationRuleBuilder {
	if len(authorities) == 1 {
		r.addAttr("hasAuthority('" + authorities[0] + "')")
		return r
	}
	r.addAttr("hasAnyAuthority('" + joinStrings(authorities, "','") + "')")
	return r
}

// HasRole 要求特定角色。
func (r *expressionBasedUrlRegistry) HasRole(role string) UrlAuthorizationRuleBuilder {
	r.addAttr("hasRole('" + role + "')")
	return r
}

// HasAnyRole 要求任意角色。
func (r *expressionBasedUrlRegistry) HasAnyRole(roles ...string) UrlAuthorizationRuleBuilder {
	if len(roles) == 1 {
		r.addAttr("hasRole('" + roles[0] + "')")
		return r
	}
	r.addAttr("hasAnyRole('" + joinStrings(roles, "','") + "')")
	return r
}

// PermitAll 允许所有。
func (r *expressionBasedUrlRegistry) PermitAll() UrlAuthorizationRuleBuilder {
	r.addAttr("permitAll")
	return r
}

// DenyAll 拒绝所有。
func (r *expressionBasedUrlRegistry) DenyAll() UrlAuthorizationRuleBuilder {
	r.addAttr("denyAll")
	return r
}

// Authenticated 要求认证。
func (r *expressionBasedUrlRegistry) Authenticated() UrlAuthorizationRuleBuilder {
	r.addAttr("authenticated")
	return r
}

// And 提交当前规则并开始下一个。
func (r *expressionBasedUrlRegistry) And() UrlAuthorizationRuleBuilder {
	r.commit()
	return r
}

// Get 获取所有规则。
func (r *expressionBasedUrlRegistry) Get() []UrlAuthorizationRule {
	r.commit()
	result := make([]UrlAuthorizationRule, len(r.rules))
	copy(result, r.rules)
	return result
}

// requestMatchers 注册 URL 匹配规则并返回注册表。
func (r *expressionBasedUrlRegistry) requestMatchers(patterns ...string) UrlAuthorizationRuleBuilder {
	r.commit()
	r.current = &currentRule{
		patterns: patterns,
		attrs:    make([]string, 0),
	}
	return r
}

// anyRequest 配置所有请求。
func (r *expressionBasedUrlRegistry) anyRequest() UrlAuthorizationRuleBuilder {
	return r.requestMatchers("**")
}

// addAttr 向当前规则添加属性。
func (r *expressionBasedUrlRegistry) addAttr(attr string) {
	if r.current == nil {
		return
	}
	r.current.attrs = append(r.current.attrs, attr)
}

// commit 提交当前规则到规则列表。
func (r *expressionBasedUrlRegistry) commit() {
	if r.current == nil {
		return
	}
	if len(r.current.patterns) > 0 && len(r.current.attrs) > 0 {
		r.rules = append(r.rules, UrlAuthorizationRule{
			Patterns:   r.current.patterns,
			Attributes: r.current.attrs,
		})
	}
	r.current = nil
}

// authorizeRequests 授权请求配置实现。
type authorizeRequests struct {
	registry *registryWrapper
}

// NewAuthorizeRequests 创建授权请求配置。
func NewAuthorizeRequests() AuthorizeRequests {
	return &authorizeRequests{
		registry: &registryWrapper{
			registry: &expressionBasedUrlRegistry{
				rules: make([]UrlAuthorizationRule, 0),
			},
		},
	}
}

// RequestMatchers 注册 URL 匹配规则。
func (a *authorizeRequests) RequestMatchers(patterns ...string) UrlAuthorizationRuleBuilder {
	return a.registry.requestMatchers(patterns...)
}

// AnyRequest 配置所有请求。
func (a *authorizeRequests) AnyRequest() UrlAuthorizationRuleBuilder {
	return a.registry.anyRequest()
}

// registryWrapper 包装 expressionBasedUrlRegistry 以实现 ExpressionInterceptUrlRegistry 接口。
//
// 将 builder 和 registry 方法组合到单一类型中。
type registryWrapper struct {
	registry *expressionBasedUrlRegistry
}

// HasAnyAuthority 要求指定权限（传入单个或多个）。
func (w *registryWrapper) HasAnyAuthority(authorities ...string) UrlAuthorizationRuleBuilder {
	w.registry.HasAnyAuthority(authorities...)
	return w
}

// HasRole 要求特定角色。
func (w *registryWrapper) HasRole(role string) UrlAuthorizationRuleBuilder {
	w.registry.HasRole(role)
	return w
}

// HasAnyRole 要求任意角色。
func (w *registryWrapper) HasAnyRole(roles ...string) UrlAuthorizationRuleBuilder {
	w.registry.HasAnyRole(roles...)
	return w
}

// PermitAll 允许所有。
func (w *registryWrapper) PermitAll() UrlAuthorizationRuleBuilder {
	w.registry.PermitAll()
	return w
}

// DenyAll 拒绝所有。
func (w *registryWrapper) DenyAll() UrlAuthorizationRuleBuilder {
	w.registry.DenyAll()
	return w
}

// Authenticated 要求认证。
func (w *registryWrapper) Authenticated() UrlAuthorizationRuleBuilder {
	w.registry.Authenticated()
	return w
}

// And 添加下一个规则。
func (w *registryWrapper) And() UrlAuthorizationRuleBuilder {
	w.registry.And()
	return w
}

// Get 获取所有规则。
func (w *registryWrapper) Get() []UrlAuthorizationRule {
	return w.registry.Get()
}

// requestMatchers 注册 URL 匹配规则并返回注册表。
func (w *registryWrapper) requestMatchers(patterns ...string) UrlAuthorizationRuleBuilder {
	w.registry.requestMatchers(patterns...)
	return w
}

// anyRequest 配置所有请求。
func (w *registryWrapper) anyRequest() UrlAuthorizationRuleBuilder {
	return w.requestMatchers("**")
}

// joinStrings 使用分隔符连接字符串切片。
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for _, s := range strs[1:] {
		result += sep + s
	}
	return result
}

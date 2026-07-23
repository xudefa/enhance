package filter

// securityFilterImpl 安全过滤器实现。
//
// 使用 SecurityFilterChainManager 管理多个安全过滤器链，
// 根据请求匹配合适的过滤器链并执行。
type securityFilterImpl struct {
	manager SecurityFilterChainManager
	order   int
}

// NewSecurityFilter 创建安全过滤器。
func NewSecurityFilter(manager SecurityFilterChainManager) Filter {
	return &securityFilterImpl{
		manager: manager,
		order:   0,
	}
}

// NewSecurityFilterWithOrder 创建带顺序的安全过滤器。
func NewSecurityFilterWithOrder(manager SecurityFilterChainManager, order int) Filter {
	return &securityFilterImpl{
		manager: manager,
		order:   order,
	}
}

// DoFilter 执行安全过滤。
//
// 从 SecurityFilterChainManager 获取匹配的过滤器链，
// 如果存在匹配的链则执行，否则继续执行下一个过滤器。
func (f *securityFilterImpl) DoFilter(ctx interface{}, request interface{}, response interface{}, chain FilterChain) error {
	securityChain := f.manager.GetSecurityFilterChain(request)
	if securityChain != nil {
		return securityChain.DoFilter(ctx, request, response)
	}
	return chain.DoFilter(ctx, request, response)
}

// Order 过滤器顺序。
func (f *securityFilterImpl) Order() int {
	return f.order
}

// SetOrder 设置过滤器顺序。
func (f *securityFilterImpl) SetOrder(order int) {
	f.order = order
}

// Manager 获取安全管理器。
func (f *securityFilterImpl) Manager() SecurityFilterChainManager {
	return f.manager
}

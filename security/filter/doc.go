// Package filter 提供安全过滤器链抽象，用于 enhance 框架。
//
// 该模块将过滤器功能从 security 包中分离，提供通用的过滤器链机制。
// 参考 Spring Security Filter 模块的设计理念。
//
// # 架构设计
//
//   - Filter: 过滤器接口，定义过滤逻辑
//   - FilterChain: 过滤器链接口，管理过滤器的执行顺序
//   - SecurityFilterChain: 安全过滤器链接口，匹配请求并执行安全过滤
//   - SecurityFilterChainManager: 安全过滤器链管理器接口，管理多个安全过滤器链
//
// # 核心功能
//
//   - 过滤器链: 支持过滤器的顺序执行和链式调用
//   - 安全过滤器链: 支持请求匹配和安全过滤逻辑
//   - 过滤器管理: 支持多个安全过滤器链的管理
//
// # 使用方式
//
// 创建过滤器链：
//
//	chain := filter.NewDefaultFilterChain()
//	chain.AddFilter(authFilter)
//	chain.AddFilter(authorizationFilter)
//	err := chain.DoFilter(ctx, request, response)
//
// 管理安全过滤器链：
//
//	manager := filter.NewSecurityFilterChainManager()
//	manager.AddSecurityFilterChain(filterChain)
//	chain := manager.GetSecurityFilterChain(request)
package filter

// Filter 过滤器接口。
//
// 定义过滤逻辑，每个过滤器可以决定是否继续执行下一个过滤器。
// 过滤器按顺序执行，形成过滤器链。
type Filter interface {
	// DoFilter 执行过滤。
	DoFilter(ctx interface{}, request interface{}, response interface{}, chain FilterChain) error
	// Order 过滤器顺序，值越小优先级越高。
	Order() int
}

// FilterChain 过滤器链接口。
//
// 按顺序执行多个过滤器。
// 每个过滤器可以决定是否继续执行下一个过滤器。
type FilterChain interface {
	// DoFilter 执行过滤器链。
	DoFilter(ctx interface{}, request interface{}, response interface{}) error
	// AddFilter 添加过滤器到链末尾。
	AddFilter(filter Filter)
	// GetFilters 获取所有过滤器。
	GetFilters() []Filter
}

// SecurityFilterChain 安全过滤器链接口。
//
// 匹配请求并执行安全过滤逻辑。
// 每个安全过滤器链可以匹配不同的请求类型。
type SecurityFilterChain interface {
	// Matches 判断是否匹配该请求。
	Matches(request interface{}) bool
	// DoFilter 执行安全过滤器链。
	DoFilter(ctx interface{}, request interface{}, response interface{}) error
	// GetFilters 获取所有过滤器。
	GetFilters() []Filter
}

// SecurityFilterChainManager 安全过滤器链管理器接口。
//
// 管理多个安全过滤器链，根据请求匹配合适的过滤器链。
type SecurityFilterChainManager interface {
	// AddSecurityFilterChain 添加安全过滤器链。
	AddSecurityFilterChain(chain SecurityFilterChain)
	// GetSecurityFilterChain 获取匹配请求的安全过滤器链。
	GetSecurityFilterChain(request interface{}) SecurityFilterChain
	// GetSecurityFilterChains 获取所有安全过滤器链。
	GetSecurityFilterChains() []SecurityFilterChain
}

// MatcherFunc 请求匹配函数。
//
// 用于判断请求是否匹配特定条件，可作为 SecurityFilterChain 的匹配策略。
type MatcherFunc func(request interface{}) bool

// Matches 实现匹配逻辑。
func (f MatcherFunc) Matches(request interface{}) bool {
	return f(request)
}

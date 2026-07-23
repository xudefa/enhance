package filter

import "sync"

// securityFilterChainManagerImpl 安全过滤器链管理器实现。
//
// 管理多个安全过滤器链，根据请求匹配合适的过滤器链。
type securityFilterChainManagerImpl struct {
	chains []SecurityFilterChain
	mu     sync.RWMutex
}

// NewSecurityFilterChainManager 创建安全过滤器链管理器。
func NewSecurityFilterChainManager() SecurityFilterChainManager {
	return &securityFilterChainManagerImpl{
		chains: make([]SecurityFilterChain, 0),
	}
}

// AddSecurityFilterChain 添加安全过滤器链。
func (m *securityFilterChainManagerImpl) AddSecurityFilterChain(chain SecurityFilterChain) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.chains = append(m.chains, chain)
}

// GetSecurityFilterChain 获取匹配请求的安全过滤器链。
//
// 按添加顺序遍历，返回第一个匹配的过滤器链。
func (m *securityFilterChainManagerImpl) GetSecurityFilterChain(request interface{}) SecurityFilterChain {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, chain := range m.chains {
		if chain.Matches(request) {
			return chain
		}
	}
	return nil
}

// GetSecurityFilterChains 获取所有安全过滤器链。
func (m *securityFilterChainManagerImpl) GetSecurityFilterChains() []SecurityFilterChain {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]SecurityFilterChain, len(m.chains))
	copy(result, m.chains)
	return result
}

// simpleSecurityFilterChain 简单安全过滤器链实现。
//
// 基于 MatcherFunc 匹配请求，并执行过滤器链。
type simpleSecurityFilterChain struct {
	matcher MatcherFunc
	filters []Filter
}

// NewSimpleSecurityFilterChain 创建简单安全过滤器链。
func NewSimpleSecurityFilterChain(matcher MatcherFunc, filters ...Filter) SecurityFilterChain {
	return &simpleSecurityFilterChain{
		matcher: matcher,
		filters: filters,
	}
}

// Matches 判断是否匹配该请求。
func (c *simpleSecurityFilterChain) Matches(request interface{}) bool {
	return c.matcher.Matches(request)
}

// DoFilter 执行安全过滤器链。
func (c *simpleSecurityFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	chain := NewDefaultFilterChainWithFilters(c.filters...)
	return chain.DoFilter(ctx, request, response)
}

// GetFilters 获取所有过滤器。
func (c *simpleSecurityFilterChain) GetFilters() []Filter {
	return c.filters
}

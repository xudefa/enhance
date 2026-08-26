package security

import (
	"context"
	"fmt"

	"github.com/xudefa/enhance/security/filter"
)

// filterChainProxy 过滤器链代理（内部实现）。
//
// 负责将多个 SecurityFilter 按顺序执行，并在最后执行 SecurityFilterChain。
// 不对外暴露，仅通过 securityFilterChainAdapter 适配为外部接口。
type filterChainProxy struct {
	filters []SecurityFilter
	chain   SecurityFilterChain
}

// newFilterChainProxy 创建过滤器链代理实例。
func newFilterChainProxy(filters []SecurityFilter, chain SecurityFilterChain) *filterChainProxy {
	return &filterChainProxy{
		filters: filters,
		chain:   chain,
	}
}

// doFilterWithChain 以类型安全方式执行过滤器链
func (p *filterChainProxy) doFilterWithChain(ctx context.Context, request SecurityRequest, response SecurityResponse, chain filter.FilterChain) error {
	if p == nil || chain == nil {
		return nil
	}
	return chain.DoFilter(ctx, request, response)
}

// doFilterInternal 执行过滤器链中的指定索引过滤器
func (p *filterChainProxy) doFilterInternal(ctx context.Context, request SecurityRequest, response SecurityResponse, index int) error {
	if index >= len(p.filters) {
		return p.chain.DoFilter(ctx, request, response)
	}

	nextChain := &filterChainAdapter{
		vfc: &virtualFilterChain{proxy: p, index: index + 1},
	}
	return p.filters[index].DoFilter(ctx, request, response, nextChain)
}

// virtualFilterChain 虚拟过滤器链（内部实现）。
//
// 用于在过滤器链中递归调用下一个过滤器。
type virtualFilterChain struct {
	proxy *filterChainProxy
	index int
}

// DoFilter 执行下一个过滤器
func (c *virtualFilterChain) DoFilter(ctx context.Context, request SecurityRequest, response SecurityResponse) error {
	return c.proxy.doFilterInternal(ctx, request, response, c.index)
}

// securityFilterChainAdapter 将内部 typed filterChainProxy 适配为 filter.SecurityFilterChain 接口。
type securityFilterChainAdapter struct {
	proxy *filterChainProxy
}

func (a *securityFilterChainAdapter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return a.proxy.doFilterWithChain(ctxVal, req, resp, &filterChainAdapter{vfc: &virtualFilterChain{proxy: a.proxy, index: 0}})
}

func (a *securityFilterChainAdapter) Matches(request interface{}) bool {
	_, ok := request.(SecurityRequest)
	return ok
}

func (a *securityFilterChainAdapter) GetFilters() []filter.Filter {
	result := make([]filter.Filter, len(a.proxy.filters))
	copy(result, a.proxy.filters)
	return result
}

// filterChainAdapter 将 virtualFilterChain 适配为 filter.FilterChain 接口。
type filterChainAdapter struct {
	vfc *virtualFilterChain
}

func (a *filterChainAdapter) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	ctxVal, ok := ctx.(context.Context)
	if !ok {
		return fmt.Errorf("expected context.Context, got %T", ctx)
	}
	req, ok := request.(SecurityRequest)
	if !ok {
		return fmt.Errorf("expected SecurityRequest, got %T", request)
	}
	resp, ok := response.(SecurityResponse)
	if !ok {
		return fmt.Errorf("expected SecurityResponse, got %T", response)
	}
	return a.vfc.DoFilter(ctxVal, req, resp)
}

func (a *filterChainAdapter) AddFilter(f filter.Filter) {}

func (a *filterChainAdapter) GetFilters() []filter.Filter {
	return nil
}

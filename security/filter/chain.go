package filter

import (
	"fmt"
	"sort"
)

// defaultFilterChain 默认过滤器链实现。
//
// 按顺序执行多个过滤器，支持动态添加过滤器。
type defaultFilterChain struct {
	filters []Filter
}

// NewDefaultFilterChain 创建默认过滤器链。
func NewDefaultFilterChain() FilterChain {
	return &defaultFilterChain{}
}

// NewDefaultFilterChainWithFilters 创建带过滤器的默认过滤器链。
func NewDefaultFilterChainWithFilters(filters ...Filter) FilterChain {
	c := &defaultFilterChain{
		filters: make([]Filter, 0, len(filters)),
	}
	c.filters = append(c.filters, filters...)
	c.SortFilters()
	return c
}

// DoFilter 执行过滤器链。
//
// 按顺序执行所有过滤器，每个过滤器可以调用 chain.DoFilter 继续执行。
func (c *defaultFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	return c.doFilterInternal(ctx, request, response, 0)
}

// doFilterInternal 内部过滤器执行方法。
func (c *defaultFilterChain) doFilterInternal(ctx interface{}, request interface{}, response interface{}, index int) error {
	if index >= len(c.filters) {
		return nil
	}

	filter := c.filters[index]
	return filter.DoFilter(ctx, request, response, &virtualFilterChain{
		chain: c,
		index: index + 1,
	})
}

// AddFilter 添加过滤器到链末尾。
func (c *defaultFilterChain) AddFilter(filter Filter) {
	c.filters = append(c.filters, filter)
}

// GetFilters 获取所有过滤器。
func (c *defaultFilterChain) GetFilters() []Filter {
	result := make([]Filter, len(c.filters))
	copy(result, c.filters)
	return result
}

// SortFilters 按 Order 排序过滤器。
func (c *defaultFilterChain) SortFilters() {
	sort.SliceStable(c.filters, func(i, j int) bool {
		return c.filters[i].Order() < c.filters[j].Order()
	})
}

// virtualFilterChain 虚拟过滤器链。
//
// 在过滤器执行过程中跟踪当前索引，实现链式调用。
type virtualFilterChain struct {
	chain *defaultFilterChain
	index int
}

// DoFilter 执行下一个过滤器。
func (c *virtualFilterChain) DoFilter(ctx interface{}, request interface{}, response interface{}) error {
	return c.chain.doFilterInternal(ctx, request, response, c.index)
}

// AddFilter 添加过滤器（虚拟链不支持）。
func (c *virtualFilterChain) AddFilter(filter Filter) {}

// GetFilters 获取过滤器（虚拟链返回空列表）。
func (c *virtualFilterChain) GetFilters() []Filter {
	return nil
}

// String 返回过滤器链的字符串表示。
func (c *defaultFilterChain) String() string {
	return fmt.Sprintf("DefaultFilterChain{filters=%d}", len(c.filters))
}

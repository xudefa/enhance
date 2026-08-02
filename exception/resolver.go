package exception

import (
	"context"
	"sort"
	"sync"
)

// ResolverChain 解析器链
//
// ResolverChain 管理多个 ExceptionResolver，并按优先级顺序执行它们。
// 当处理异常时，会遍历解析器链，找到第一个支持该异常的解析器进行处理。
type ResolverChain struct {
	mu        sync.RWMutex
	resolvers []ExceptionResolver
}

// NewResolverChain 创建新的解析器链
//
// 创建一个空的解析器链，可以通过 AddResolver 方法添加解析器。
func NewResolverChain() *ResolverChain {
	return &ResolverChain{
		resolvers: make([]ExceptionResolver, 0),
	}
}

// AddResolver 添加解析器（自动按 Order 排序）
//
// 将解析器添加到链中，并自动按照 Order 值进行排序。
// Order 值越小，优先级越高，会先被调用。
func (c *ResolverChain) AddResolver(resolver ExceptionResolver) {
	c.mu.Lock()
	c.resolvers = append(c.resolvers, resolver)
	c.mu.Unlock()

	// 排序在锁外执行，避免在持有锁时调用用户 Order()（可能重入链）导致死锁。
	// 使用一致快照排序后回写；若排序期间有并发 AddResolver，重试直到成功写入，
	// 避免用陈旧快照覆盖并发添加的解析器（lost update）。
	for {
		c.mu.RLock()
		snapshot := make([]ExceptionResolver, len(c.resolvers))
		copy(snapshot, c.resolvers)
		c.mu.RUnlock()

		sort.SliceStable(snapshot, func(i, j int) bool {
			return snapshot[i].Order() < snapshot[j].Order()
		})

		c.mu.Lock()
		if len(c.resolvers) == len(snapshot) {
			// 排序期间没有新增解析器，安全写入排序结果
			c.resolvers = snapshot
			c.mu.Unlock()
			return
		}
		// 期间有并发新增，重新快照排序
		c.mu.Unlock()
	}
}

// GetResolvers 获取所有解析器
//
// 返回当前解析器链中的所有解析器，按优先级排序。
// 返回的是内部切片的副本，避免调用方修改内部状态。
func (c *ResolverChain) GetResolvers() []ExceptionResolver {
	c.mu.RLock()
	defer c.mu.RUnlock()
	resolvers := make([]ExceptionResolver, len(c.resolvers))
	copy(resolvers, c.resolvers)
	return resolvers
}

// Resolve 查找第一个支持的解析器并解析异常
//
// 遍历解析器链，找到第一个 Supports 返回 true 的解析器，
// 并调用其 Resolve 方法处理异常。如果没有找到支持的解析器，返回 nil。
// 先在锁内获取快照再释放锁，避免在持有锁时调用用户代码导致死锁。
func (c *ResolverChain) Resolve(ctx context.Context, err error) *ErrorResponse {
	c.mu.RLock()
	resolvers := make([]ExceptionResolver, len(c.resolvers))
	copy(resolvers, c.resolvers)
	c.mu.RUnlock()

	for _, resolver := range resolvers {
		if resolver.Supports(err) {
			return resolver.Resolve(ctx, err)
		}
	}
	return nil
}

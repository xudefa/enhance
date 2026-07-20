package server

import (
	"sync"
)

// RequestScope 请求级别作用域
//
// 为每个 HTTP 请求创建独立的作用域，用于管理请求生命周期内的 Bean 实例。
// 请求结束时自动清理所有 Bean 实例。
type RequestScope struct {
	mu    sync.RWMutex
	cache map[string]any
}

// NewRequestScope 创建新的请求级别作用域
func NewRequestScope() *RequestScope {
	return &RequestScope{
		cache: make(map[string]any),
	}
}

// Get 获取或创建指定名称的 Bean 实例
//
// 如果缓存中已存在该 Bean，直接返回；否则调用 factory 创建并缓存。
func (s *RequestScope) Get(name string, factory func() any) any {
	s.mu.RLock()
	if val, ok := s.cache[name]; ok {
		s.mu.RUnlock()
		return val
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// 双重检查
	if val, ok := s.cache[name]; ok {
		return val
	}

	val := factory()
	s.cache[name] = val
	return val
}

// Set 设置 Bean 实例
func (s *RequestScope) Set(name string, bean any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[name] = bean
}

// Clear 清理作用域中的所有 Bean 实例
func (s *RequestScope) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache = make(map[string]any)
}

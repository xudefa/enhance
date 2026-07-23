package lifecycle

import (
	"context"
	"fmt"
	"sync"
)

// ==================== 简化生命周期钩子（Go 风格） ====================

// OnInit 在容器初始化完成后调用
func (h *HookFunc) OnInit(ctx context.Context) error {
	if h.onInit != nil {
		return h.onInit(ctx)
	}
	return nil
}

// OnStart 在应用启动时调用
func (h *HookFunc) OnStart(ctx context.Context) error {
	if h.onStart != nil {
		return h.onStart(ctx)
	}
	return nil
}

// OnStop 在应用停止时调用
func (h *HookFunc) OnStop(ctx context.Context) error {
	if h.onStop != nil {
		return h.onStop(ctx)
	}
	return nil
}

// NewHookFunc 创建函数式钩子
func NewHookFunc(onInit, onStart, onStop func(context.Context) error) Hook {
	return &HookFunc{
		onInit:  onInit,
		onStart: onStart,
		onStop:  onStop,
	}
}

// OnInitFunc 创建仅实现 OnInit 的钩子
func OnInitFunc(fn func(context.Context) error) Hook {
	return &HookFunc{onInit: fn}
}

// OnStartFunc 创建仅实现 OnStart 的钩子
func OnStartFunc(fn func(context.Context) error) Hook {
	return &HookFunc{onStart: fn}
}

// OnStopFunc 创建仅实现 OnStop 的钩子
func OnStopFunc(fn func(context.Context) error) Hook {
	return &HookFunc{onStop: fn}
}

// ==================== Hook 注册表 ====================

// HookRegistry 钩子注册表
//
// 管理多个 Hook，按注册顺序执行 OnInit/OnStart，逆序执行 OnStop。
// 使用 RWMutex 优化读多写少场景，读操作无写锁竞争。
type HookRegistry struct {
	mu    sync.RWMutex
	hooks []Hook
}

// NewHookRegistry 创建钩子注册表
func NewHookRegistry() *HookRegistry {
	return &HookRegistry{
		hooks: make([]Hook, 0),
	}
}

// Register 注册一个钩子
func (r *HookRegistry) Register(hook Hook) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.hooks = append(r.hooks, hook)
}

// RegisterFunc 通过函数注册钩子
func (r *HookRegistry) RegisterFunc(onInit, onStart, onStop func(context.Context) error) {
	r.Register(&HookFunc{
		onInit:  onInit,
		onStart: onStart,
		onStop:  onStop,
	})
}

// InitAll 按注册顺序执行所有 OnInit
func (r *HookRegistry) InitAll(ctx context.Context) error {
	r.mu.RLock()
	// 复制切片内容，避免并发 Register 导致的数据竞争
	hooks := make([]Hook, len(r.hooks))
	copy(hooks, r.hooks)
	r.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook.OnInit(ctx); err != nil {
			return fmt.Errorf("hook %d OnInit failed: %w", i, err)
		}
	}
	return nil
}

// StartAll 按注册顺序执行所有 OnStart
func (r *HookRegistry) StartAll(ctx context.Context) error {
	r.mu.RLock()
	// 复制切片内容，避免并发 Register 导致的数据竞争
	hooks := make([]Hook, len(r.hooks))
	copy(hooks, r.hooks)
	r.mu.RUnlock()

	for i, hook := range hooks {
		if err := hook.OnStart(ctx); err != nil {
			return fmt.Errorf("hook %d OnStart failed: %w", i, err)
		}
	}
	return nil
}

// StopAll 按注册逆序执行所有 OnStop
func (r *HookRegistry) StopAll(ctx context.Context) error {
	r.mu.RLock()
	// 复制切片内容，避免并发 Register 导致的数据竞争
	hooks := make([]Hook, len(r.hooks))
	copy(hooks, r.hooks)
	r.mu.RUnlock()

	for i := len(hooks) - 1; i >= 0; i-- {
		if err := hooks[i].OnStop(ctx); err != nil {
			return fmt.Errorf("hook %d OnStop failed: %w", i, err)
		}
	}
	return nil
}

// Count 返回已注册的钩子数量
func (r *HookRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.hooks)
}

// GetAll 返回所有已注册的钩子
func (r *HookRegistry) GetAll() []Hook {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]Hook, len(r.hooks))
	copy(result, r.hooks)
	return result
}

// ==================== 全局钩子注册表 ====================

var globalHooks = NewHookRegistry()

// GlobalHookRegistry 获取全局钩子注册表
func GlobalHookRegistry() *HookRegistry {
	return globalHooks
}

// RegisterHook 注册到全局钩子注册表
func RegisterHook(hook Hook) {
	globalHooks.Register(hook)
}

// RegisterHookFunc 通过函数注册到全局钩子注册表
func RegisterHookFunc(onInit, onStart, onStop func(context.Context) error) {
	globalHooks.RegisterFunc(onInit, onStart, onStop)
}

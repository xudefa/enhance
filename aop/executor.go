// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// classifiedAdvices 按类型分类后的通知列表。
type classifiedAdvices struct {
	before         []Advice
	after          []Advice
	afterReturning []Advice
	afterThrowing  []Advice
	around         []Advice
}

// chainExecutorConfig 执行器配置。
type chainExecutorConfig struct {
	recoverPanic bool
	interceptors []Interceptor
}

// ChainExecutorOption 执行器选项函数。
type ChainExecutorOption func(*chainExecutorConfig)

// WithRecovery 启用 panic 恢复。
func WithRecovery() ChainExecutorOption {
	return func(c *chainExecutorConfig) {
		c.recoverPanic = true
	}
}

// WithInterceptor 添加自定义拦截器。
func WithInterceptor(i Interceptor) ChainExecutorOption {
	return func(c *chainExecutorConfig) {
		c.interceptors = append(c.interceptors, i)
	}
}

// defaultChainExecutor 默认通知链执行器。
type defaultChainExecutor struct {
	config chainExecutorConfig
}

// classifyAdvices 将切面列表中的通知按类型分类。
func classifyAdvices(aspects []*AspectMeta) *classifiedAdvices {
	counts := make(map[AdviceType]int, 5)
	for _, aspect := range aspects {
		if aspect != nil && aspect.Advice != nil {
			counts[aspect.Advice.Type()]++
		}
	}

	ca := &classifiedAdvices{
		before:         make([]Advice, 0, counts[AdviceTypeBefore]),
		after:          make([]Advice, 0, counts[AdviceTypeAfter]),
		afterReturning: make([]Advice, 0, counts[AdviceTypeAfterReturning]),
		afterThrowing:  make([]Advice, 0, counts[AdviceTypeAfterThrowing]),
		around:         make([]Advice, 0, counts[AdviceTypeAround]),
	}
	for _, aspect := range aspects {
		if aspect == nil || aspect.Advice == nil {
			continue
		}
		switch aspect.Advice.Type() {
		case AdviceTypeBefore:
			ca.before = append(ca.before, aspect.Advice)
		case AdviceTypeAfter:
			ca.after = append(ca.after, aspect.Advice)
		case AdviceTypeAfterReturning:
			ca.afterReturning = append(ca.afterReturning, aspect.Advice)
		case AdviceTypeAfterThrowing:
			ca.afterThrowing = append(ca.afterThrowing, aspect.Advice)
		case AdviceTypeAround:
			ca.around = append(ca.around, aspect.Advice)
		}
	}
	return ca
}

// NewChainExecutor 创建通知链执行器
func NewChainExecutor(opts ...ChainExecutorOption) ChainExecutor {
	config := chainExecutorConfig{
		recoverPanic: true,
	}
	for _, opt := range opts {
		opt(&config)
	}
	return &defaultChainExecutor{config: config}
}

// Execute 执行通知链
func (e *defaultChainExecutor) Execute(inv Invocation, aspects []*AspectMeta, targetFunc func(...any) any) any {
	if inv == nil || targetFunc == nil {
		return nil
	}

	ca := classifyAdvices(aspects)
	// 从 Invocation 的 JoinPoint 获取 context，如果不可用则使用 Background
	ctx := context.Background()
	if jp := inv.JoinPoint(); jp != nil {
		ctx = jp.Context()
	}

	coreExecute := func(invocation Invocation) any {
		joinPoint := invocation.JoinPoint()
		if joinPoint == nil {
			return targetFunc(invocation.Arguments()...)
		}

		// 1. 执行所有 Before 通知
		for _, advice := range ca.before {
			_, _ = advice.Execute(ctx, joinPoint)
		}

		var result any
		var panicked any

		// 2. 执行 Around 通知链或目标方法
		executeBody := func() {
			if len(ca.around) > 0 {
				chain := buildAdviceChain(ca.around, targetFunc)
				result = chain(invocation)
			} else {
				var proceedErr error
				result, proceedErr = joinPoint.Proceed()
				if proceedErr != nil {
					joinPoint.SetError(proceedErr)
					if ei, ok := inv.(*invocationImpl); ok {
						ei.SetError(proceedErr)
					}
				}
			}
			joinPoint.SetResult(result)
			// 从返回值中提取 error，确保 AfterThrowing 能正确触发。
			// 覆盖两种形态：
			//   - 多返回值：末尾元素为 error（如 func() (T, error)）
			//   - 唯一返回值即 error（如 func() error）
			var extractedErr error
			if results, ok := result.([]any); ok && len(results) > 0 {
				if err, isErr := results[len(results)-1].(error); isErr {
					extractedErr = err
				}
			} else if err, ok := result.(error); ok {
				extractedErr = err
			}
			if extractedErr != nil {
				joinPoint.SetError(extractedErr)
				if ei, ok := inv.(*invocationImpl); ok {
					ei.SetError(extractedErr)
				}
			}
		}

		if e.config.recoverPanic {
			func() {
				defer func() {
					if r := recover(); r != nil {
						panicked = r
					}
				}()
				executeBody()
			}()
		} else {
			executeBody()
		}

		// 3. 执行所有 After 通知
		for _, advice := range ca.after {
			_, _ = advice.Execute(ctx, joinPoint)
		}

		// 4. 根据 panic 状态执行 AfterThrowing 或 AfterReturning
		if panicked != nil {
			panicErr := fmt.Errorf("panic: %v", panicked)
			joinPoint.SetError(panicErr)
			if ei, ok := inv.(*invocationImpl); ok {
				ei.SetError(panicErr)
			}
			for _, advice := range ca.afterThrowing {
				func() {
					defer func() {
						if r := recover(); r != nil {
							_ = r
						}
					}()
					_, _ = advice.Execute(ctx, joinPoint)
				}()
			}
			panic(panicked)
		}

		// 5. 方法返回错误时执行 AfterThrowing，正常返回时执行 AfterReturning
		if joinPoint.GetError() != nil {
			for _, advice := range ca.afterThrowing {
				func() {
					defer func() {
						if r := recover(); r != nil {
							_ = r
						}
					}()
					_, _ = advice.Execute(ctx, joinPoint)
				}()
			}
		} else {
			for _, advice := range ca.afterReturning {
				_, _ = advice.Execute(ctx, joinPoint)
			}
		}

		return result
	}

	// 从内到外包装拦截器
	interceptorCount := len(e.config.interceptors)
	execute := coreExecute
	for i := len(e.config.interceptors) - 1; i >= 0; i-- {
		interceptor := e.config.interceptors[i]
		prev := execute
		captured := interceptor
		execute = func(inv Invocation) any {
			return captured(inv, prev)
		}
	}

	// 执行通知链并更新统计
	var finalResult any
	func() {
		defer func() {
			if r := recover(); r != nil {
				stack := debug.Stack()
				panicInfo := &PanicInfo{
					Value: r,
					Stack: stack,
				}
				updateStats(panicInfo, interceptorCount)
				panic(panicInfo)
			}
		}()
		finalResult = execute(inv)
	}()
	updateStats(nil, interceptorCount)
	return finalResult
}

// PanicInfo 包含 panic 信息和堆栈
type PanicInfo struct {
	Value any
	Stack []byte
}

// Error 实现 error 接口
func (p *PanicInfo) Error() string {
	return fmt.Sprintf("panic: %v\n%s", p.Value, p.Stack)
}

// updateStats 更新通知链统计信息
func updateStats(panicked any, interceptorCount int) {
	GlobalChainStats.TotalExecutions.Add(1)
	if panicked != nil {
		GlobalChainStats.TotalPanics.Add(1)
	}
	if interceptorCount > 0 {
		GlobalChainStats.TotalInterceptors.Add(int64(interceptorCount))
	}
}

// buildAdviceChain 构建环绕通知链
func buildAdviceChain(advices []Advice, targetFunc func(...any) any) func(Invocation) any {
	return func(inv Invocation) any {
		result, err := executeAdviceChain(0, advices, inv, targetFunc)
		// 链式执行产生的错误提交到 JoinPoint 和 Invocation，
		// 使 AfterThrowing 通知与调用方（CallContext/InvokeContext）能够感知。
		if err != nil {
			if jp := inv.JoinPoint(); jp != nil {
				jp.SetError(err)
			}
			if ei, ok := inv.(*invocationImpl); ok {
				ei.SetError(err)
			}
		}
		return result
	}
}

// chainJoinPoint 包装 JoinPoint，拦截 Proceed 调用以实现链式执行。
type chainJoinPoint struct {
	inner   JoinPoint
	inv     Invocation
	proceed func() (any, error)
}

func (c *chainJoinPoint) Target() any              { return c.inner.Target() }
func (c *chainJoinPoint) Method() string           { return c.inner.Method() }
func (c *chainJoinPoint) Args() []any              { return c.inner.Args() }
func (c *chainJoinPoint) Proceed() (any, error)    { return c.proceed() }
func (c *chainJoinPoint) Context() context.Context { return c.inner.Context() }
func (c *chainJoinPoint) GetResult() any           { return c.inner.GetResult() }
func (c *chainJoinPoint) GetError() error          { return c.inner.GetError() }
func (c *chainJoinPoint) SetResult(v any)          { c.inner.SetResult(v) }
func (c *chainJoinPoint) SetError(err error)       { c.inner.SetError(err) }

func (c *chainJoinPoint) ProceedWithArgs(args []any) (any, error) {
	// 更新底层 Invocation 的参数，使后续链和目标方法使用新参数
	if s, ok := c.inv.(interface{ SetArgs([]any) }); ok {
		s.SetArgs(args)
	}
	// 同步更新 JoinPoint 的可见参数，保证链外读取到新参数
	if jp, ok := c.inner.(*joinPointImpl); ok {
		jp.args = args
	}
	return c.proceed()
}

// executeAdviceChain 递归执行环绕通知链，返回结果与错误。
//
// 目标方法或内层 Around 通知返回的错误通过返回值逐层上抛，
// 使外层 Around 通知能通过 proceed() 的 error 返回值感知并处理。
func executeAdviceChain(idx int, advices []Advice, inv Invocation, targetFunc func(...any) any) (any, error) {
	if idx >= len(advices) {
		return extractChainError(targetFunc(inv.Arguments()...))
	}

	currentIdx := idx
	ctx := context.Background()
	if jp := inv.JoinPoint(); jp != nil {
		ctx = jp.Context()
	}

	proceed := func() (any, error) {
		return executeAdviceChain(currentIdx+1, advices, inv, targetFunc)
	}

	// 包装 JoinPoint，使 Around 通知调用 Proceed 时走链式调用而非原始目标
	innerJP := inv.JoinPoint()
	wrapper := &chainJoinPoint{
		inner:   innerJP,
		inv:     inv,
		proceed: proceed,
	}

	return advices[idx].Execute(ctx, wrapper)
}

// extractChainError 从目标函数结果中提取错误。
//
// 覆盖两种形态：
//   - 唯一返回值即 error（如 func() error）
//   - 多返回值末尾为 error（如 func() (T, error)）
//
// 返回 (结果, 错误)，错误可被 Around 通知通过 proceed() 感知。
func extractChainError(result any) (any, error) {
	if err, ok := result.(error); ok && err != nil {
		return nil, err
	}
	if results, ok := result.([]any); ok && len(results) > 0 {
		if err, isErr := results[len(results)-1].(error); isErr && err != nil {
			return results, err
		}
	}
	return result, nil
}

// defaultExecutor 全局默认通知链执行器
var defaultExecutor ChainExecutor = NewChainExecutor()
var defaultExecutorMu sync.RWMutex

// DefaultChainExecutor 获取默认通知链执行器
func DefaultChainExecutor() ChainExecutor {
	defaultExecutorMu.RLock()
	defer defaultExecutorMu.RUnlock()
	return defaultExecutor
}

// SetDefaultChainExecutor 设置默认通知链执行器
func SetDefaultChainExecutor(executor ChainExecutor) {
	if executor == nil {
		return
	}
	defaultExecutorMu.Lock()
	defer defaultExecutorMu.Unlock()
	defaultExecutor = executor
}

// getDefaultExecutor 内部获取默认执行器（无锁，仅包内使用）
func getDefaultExecutor() ChainExecutor {
	defaultExecutorMu.RLock()
	defer defaultExecutorMu.RUnlock()
	return defaultExecutor
}

// ChainStats 通知链统计信息
type ChainStats struct {
	TotalExecutions   atomic.Int64
	TotalPanics       atomic.Int64
	TotalInterceptors atomic.Int64
}

// GlobalChainStats 全局通知链统计
var GlobalChainStats ChainStats

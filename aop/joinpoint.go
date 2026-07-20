// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"context"
	"reflect"
	"sync"
)

// invocationPool 复用 invocation 对象。
//
// 性能优化：减少每次方法调用时的内存分配和 GC 压力。
var invocationPool = sync.Pool{
	New: func() any {
		return &invocation{
			args: make([]any, 0, 8),
		}
	},
}

// acquireInvocation 从池中获取 invocation 对象。
func acquireInvocation() *invocation {
	inv := invocationPool.Get().(*invocation)
	inv.args = inv.args[:0]
	return inv
}

// releaseInvocation 归还 invocation 对象到池中。
func releaseInvocation(inv *invocation) {
	inv.method = nil
	inv.args = inv.args[:0]
	inv.this = nil
	inv.target = nil
	inv.sig = nil
	inv.proceed = nil
	inv.ctx = nil
	invocationPool.Put(inv)
}

// methodSignature 方法签名内部实现。
type methodSignature struct {
	name          string       // 方法名
	declaringType reflect.Type // 方法声明的类型
}

func (m *methodSignature) Name() string {
	return m.name
}

func (m *methodSignature) DeclaringType() reflect.Type {
	return m.declaringType
}

// NewMethodSignature 创建方法签名。
//
// 参数:
//   - name: 方法名
//   - t: 方法声明的类型
//
// 返回值:
//   - MethodSignature: 方法签名实例
func NewMethodSignature(name string, t reflect.Type) MethodSignature {
	return &methodSignature{
		name:          name,
		declaringType: t,
	}
}

// invocation 调用信息内部实现。
//
// 存储方法调用的完整上下文，包括方法本身、参数、代理对象、目标对象和继续执行函数。
type invocation struct {
	method  any             // 被调用的方法
	args    []any           // 调用参数列表
	this    any             // 代理对象
	target  any             // 原始目标对象
	sig     MethodSignature // 方法签名
	proceed ProceedFunc     // 继续执行函数（用于通知链）
	ctx     context.Context // 调用上下文
}

func (i *invocation) Method() any {
	return i.method
}

func (i *invocation) Args() []any {
	return i.args
}

func (i *invocation) This() any {
	return i.this
}

func (i *invocation) Target() any {
	return i.target
}

func (i *invocation) Signature() MethodSignature {
	return i.sig
}

func (i *invocation) Proceed(args ...any) any {
	if i.proceed != nil {
		return i.proceed(args...)
	}
	return nil
}

func (i *invocation) SetProceed(p ProceedFunc) {
	i.proceed = p
}

func (i *invocation) Context() context.Context {
	if i.ctx != nil {
		return i.ctx
	}
	return context.Background()
}

func (i *invocation) SetContext(ctx context.Context) {
	i.ctx = ctx
}

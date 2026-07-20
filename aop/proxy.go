// Package aop 提供面向切面编程（AOP）支持。
// ProxyFactory 结构体定义在 doc.go 中，此处为实现。
package aop

import (
	"context"
	"fmt"
	"reflect"
	"sync"
)

// resultsPool 复用方法返回值切片。
//
// 性能优化：减少多返回值方法调用时的内存分配和 GC 压力。
var resultsPool = sync.Pool{
	New: func() any {
		s := make([]any, 0, 4)
		return &s
	},
}

// methodMeta 缓存的方法反射元数据。
//
// 避免每次方法调用都重复执行反射解析，提升性能。
type methodMeta struct {
	method       reflect.Method // 方法反射信息
	numIn        int            // 参数数量（含 receiver）
	numOut       int            // 返回值数量
	receiverType reflect.Type   // receiver 类型
	inTypes      []reflect.Type // 参数类型列表
	valuePool    *sync.Pool     // []reflect.Value 对象池
	receiverPool *sync.Pool     // receiver reflect.Value 对象池
}

// ReflectiveAopProxy 基于反射的 AOP 代理。
//
// 由于 Go 运行时无法动态替换结构体方法，ReflectiveAopProxy 通过反射
// 拦截方法调用并执行通知链。用户通过 Call/CallContext 方法调用目标方法，
// AOP 通知会自动执行。
//
// 使用方式:
//
//	proxy := weaver.Weave(target).(*aop.ReflectiveAopProxy)
//	result, err := proxy.Call("MethodName", arg1, arg2)
//	result, err := proxy.CallContext(ctx, "MethodName", arg1, arg2)
type ReflectiveAopProxy struct {
	target      any
	targetType  reflect.Type
	aspects     []*AspectMeta
	methodCache map[string][]*AspectMeta
	cacheMu     sync.RWMutex
	executor    ChainExecutor
	metaCache   sync.Map // map[string]*methodMeta，缓存反射元数据
}

// ProxyFactory 代理工厂（实现 doc.go 中定义的结构体）。
//
// 负责创建 AOP 代理对象。根据目标对象的类型（接口或结构体），创建相应的代理。
//
// 使用示例：
//
//	factory := aop.NewProxyFactory(&UserService{})
//	factory.SetAspects(aspects)
//	proxy := factory.GetProxy()
type ProxyFactory struct {
	cacheMu     sync.RWMutex             // 保护 methodCache 的读写锁
	aspectsMu   sync.RWMutex             // 保护 aspects 切片的读写锁
	target      any                      // 目标对象
	aspects     []*AspectMeta            // 切面元数据列表
	proxyType   reflect.Type             // 代理类型
	isInterface bool                     // 是否为接口类型
	methodCache map[string][]*AspectMeta // 缓存方法对应的切面列表
	executor    ChainExecutor            // 通知链执行器
}

// NewProxyFactory 创建代理工厂
//
// 参数:
//   - target: 目标对象，可以是结构体指针或接口实现
//
// 返回值:
//   - *ProxyFactory: 代理工厂实例
//
// 示例:
//
//	factory := aop.NewProxyFactory(&UserService{})
func NewProxyFactory(target any) *ProxyFactory {
	t := reflect.TypeOf(target)
	if t == nil {
		return &ProxyFactory{
			target:      target,
			proxyType:   nil,
			isInterface: false,
		}
	}
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return &ProxyFactory{
		target:      target,
		proxyType:   t,
		isInterface: t.Kind() == reflect.Interface,
	}
}

// SetAspects 设置切面
//
// 参数:
//   - aspects: 切面元数据列表
func (p *ProxyFactory) SetAspects(aspects []*AspectMeta) {
	p.aspectsMu.Lock()
	defer p.aspectsMu.Unlock()
	p.aspects = make([]*AspectMeta, len(aspects))
	copy(p.aspects, aspects)
	// 清除方法缓存，避免旧切面的缓存结果被新切面污染
	p.cacheMu.Lock()
	p.methodCache = nil
	p.cacheMu.Unlock()
}

// SetExecutor 设置通知链执行器
//
// 设置后，由此工厂创建的 ReflectiveAopProxy 将使用指定的执行器。
// 传入 nil 表示使用全局默认执行器。
func (p *ProxyFactory) SetExecutor(executor ChainExecutor) {
	p.executor = executor
}

// GetProxy 获取代理对象
//
// 根据目标对象的类型，创建并返回代理对象。
// 如果没有匹配的切面，则返回原对象。
//
// 返回值:
//   - any: 代理对象或原对象
func (p *ProxyFactory) GetProxy() any {
	if p.proxyType == nil {
		return p.target
	}

	targetVal := reflect.ValueOf(p.target)
	targetType := reflect.TypeOf(p.target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	if p.isInterface {
		return p.createInterfaceProxy()
	}

	if targetType.Kind() == reflect.Struct {
		return p.createStructProxy(targetVal, targetType)
	}

	return p.target
}

// createInterfaceProxy 创建接口代理
func (p *ProxyFactory) createInterfaceProxy() any {
	if p.target == nil {
		return nil
	}

	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	iface := reflect.TypeOf(p.target)

	// 检查是否有匹配的切面
	hasMatched := false
	for i := range iface.NumMethod() {
		method := iface.Method(i)
		if len(p.filterAspects(method)) > 0 {
			hasMatched = true
			break
		}
	}

	if !hasMatched {
		return p.target
	}

	return NewInterfaceProxyWrapper(p.target, aspectsSnapshot, iface)
}

// Target 返回原始目标对象。
func (p *ReflectiveAopProxy) Target() any {
	return p.target
}

// getExecutor 获取执行器，优先使用自定义执行器
func (p *ReflectiveAopProxy) getExecutor() ChainExecutor {
	if p.executor != nil {
		return p.executor
	}
	return getDefaultExecutor()
}

// Call 通过反射调用目标方法并执行通知链
func (p *ReflectiveAopProxy) Call(methodName string, args ...any) (any, error) {
	return p.CallContext(context.Background(), methodName, args...)
}

// CallContext 通过反射调用目标方法并执行通知链（带 context）
//
// 参数:
//   - ctx: 上下文，可通过 JoinPoint.Context() 在通知中获取
//   - methodName: 方法名
//   - args: 方法参数
//
// 返回值:
//   - any: 方法返回值（多返回值时返回 []any）
//   - error: 调用错误（仅表示方法查找失败，不包含目标方法的 panic）
func (p *ReflectiveAopProxy) CallContext(ctx context.Context, methodName string, args ...any) (any, error) {
	meta := p.getOrCacheMethodMeta(methodName)
	if meta == nil {
		return nil, fmt.Errorf("method %s not found on %s", methodName, p.targetType)
	}

	matchedAspects := p.getMatchedAspects(methodName, meta.method)

	targetFunc := func(callArgs ...any) any {
		return p.invokeWithMeta(meta, callArgs)
	}

	inv := acquireInvocation()
	inv.method = meta.method.Func.Interface()
	inv.args = append(inv.args, args...)
	inv.this = p.target
	inv.target = p.target
	inv.sig = NewMethodSignature(methodName, p.targetType)
	inv.ctx = ctx

	result := p.getExecutor().Execute(inv, matchedAspects, targetFunc)
	releaseInvocation(inv)
	return result, nil
}

// MustCall 调用目标方法，panic on error
func (p *ReflectiveAopProxy) MustCall(methodName string, args ...any) any {
	result, err := p.Call(methodName, args...)
	if err != nil {
		panic(err)
	}
	return result
}

// SetExecutor 设置通知链执行器
func (p *ReflectiveAopProxy) SetExecutor(executor ChainExecutor) {
	p.executor = executor
}

// getOrCacheMethodMeta 获取或缓存方法反射元数据
func (p *ReflectiveAopProxy) getOrCacheMethodMeta(methodName string) *methodMeta {
	if cached, ok := p.metaCache.Load(methodName); ok {
		return cached.(*methodMeta)
	}

	method, ok := p.targetType.MethodByName(methodName)
	if !ok {
		return nil
	}

	meta := &methodMeta{
		method:       method,
		numIn:        method.Type.NumIn(),
		numOut:       method.Type.NumOut(),
		receiverType: method.Type.In(0),
	}

	if meta.numIn > 1 {
		meta.inTypes = make([]reflect.Type, meta.numIn-1)
		for i := 1; i < meta.numIn; i++ {
			meta.inTypes[i-1] = method.Type.In(i)
		}
	}

	meta.valuePool = &sync.Pool{
		New: func() any {
			v := make([]reflect.Value, 0, meta.numIn)
			return &v
		},
	}

	meta.receiverPool = &sync.Pool{
		New: func() any {
			v := reflect.ValueOf(p.target)
			return &v
		},
	}

	p.metaCache.Store(methodName, meta)
	return meta
}

// invokeWithMeta 使用缓存的元数据执行方法调用
func (p *ReflectiveAopProxy) invokeWithMeta(meta *methodMeta, callArgs []any) any {
	in := *(meta.valuePool.Get().(*[]reflect.Value))
	in = in[:0]

	// 使用缓存的 receiver 避免重复反射
	receiver := *(meta.receiverPool.Get().(*reflect.Value))
	in = append(in, receiver)

	for _, a := range callArgs {
		in = append(in, reflect.ValueOf(a))
	}

	results := meta.method.Func.Call(in)

	meta.valuePool.Put(&in)
	meta.receiverPool.Put(&receiver)

	switch meta.numOut {
	case 0:
		return nil
	case 1:
		return results[0].Interface()
	default:
		// 使用对象池复用返回值切片
		retPtr := resultsPool.Get().(*[]any)
		*retPtr = (*retPtr)[:0]
		for _, r := range results {
			*retPtr = append(*retPtr, r.Interface())
		}
		retCopy := make([]any, len(*retPtr))
		copy(retCopy, *retPtr)
		resultsPool.Put(retPtr)
		return retCopy
	}
}

// getMatchedAspects 获取方法匹配的切面
func (p *ReflectiveAopProxy) getMatchedAspects(methodName string, method reflect.Method) []*AspectMeta {
	p.cacheMu.RLock()
	if p.methodCache != nil {
		if cached, ok := p.methodCache[methodName]; ok {
			p.cacheMu.RUnlock()
			return cached
		}
	}
	p.cacheMu.RUnlock()

	var matched []*AspectMeta
	for _, a := range p.aspects {
		if a != nil && a.PointCut != nil && a.PointCut.MatchMethod(method) {
			matched = append(matched, a)
		}
	}
	SortAspectsByOrder(matched)

	p.cacheMu.Lock()
	if p.methodCache == nil {
		p.methodCache = make(map[string][]*AspectMeta)
	}
	p.methodCache[methodName] = matched
	p.cacheMu.Unlock()

	return matched
}

// IsReflectiveProxy 检查对象是否为 ReflectiveAopProxy
func IsReflectiveProxy(obj any) bool {
	_, ok := obj.(*ReflectiveAopProxy)
	return ok
}

// AsReflectiveProxy 将对象转换为 ReflectiveAopProxy
func AsReflectiveProxy(obj any) (*ReflectiveAopProxy, bool) {
	proxy, ok := obj.(*ReflectiveAopProxy)
	return proxy, ok
}

// createStructProxy 创建结构体代理
func (p *ProxyFactory) createStructProxy(targetVal reflect.Value, targetType reflect.Type) any {
	proxyVal := reflect.New(targetType)

	if targetVal.Kind() == reflect.Pointer {
		proxyVal.Elem().Set(targetVal.Elem())
	} else {
		proxyVal.Elem().Set(targetVal)
	}

	ptrType := reflect.PointerTo(targetType)

	hasMatched := false
	for i := range ptrType.NumMethod() {
		method := ptrType.Method(i)
		if method.PkgPath != "" {
			continue
		}
		if len(p.filterAspects(method)) > 0 {
			hasMatched = true
			break
		}
	}

	if !hasMatched {
		return proxyVal.Interface()
	}

	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	return &ReflectiveAopProxy{
		target:      p.target,
		targetType:  ptrType,
		aspects:     aspectsSnapshot,
		methodCache: make(map[string][]*AspectMeta),
		executor:    p.executor,
	}
}

// filterAspects 过滤匹配的切面
//
// 根据方法匹配切面,并按Order排序.
// 使用缓存避免每次调用都重新匹配和排序.
func (p *ProxyFactory) filterAspects(method reflect.Method) []*AspectMeta {
	// 检查缓存
	p.cacheMu.RLock()
	if p.methodCache != nil {
		if cached, ok := p.methodCache[method.Name]; ok {
			p.cacheMu.RUnlock()
			return cached
		}
	}
	p.cacheMu.RUnlock()

	// 在锁内快照 aspects，避免 SetAspects 和 filterAspects 的 TOCTOU 竞态
	p.aspectsMu.RLock()
	aspectsSnapshot := make([]*AspectMeta, len(p.aspects))
	copy(aspectsSnapshot, p.aspects)
	p.aspectsMu.RUnlock()

	var matched []*AspectMeta
	for _, a := range aspectsSnapshot {
		if a != nil && a.PointCut != nil && a.PointCut.MatchMethod(method) {
			matched = append(matched, a)
		}
	}

	// 按Order排序
	SortAspectsByOrder(matched)

	// 存入缓存
	p.cacheMu.Lock()
	if p.methodCache == nil {
		p.methodCache = make(map[string][]*AspectMeta)
	}
	p.methodCache[method.Name] = matched
	p.cacheMu.Unlock()

	return matched
}

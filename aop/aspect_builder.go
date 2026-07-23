package aop

import "strings"

// AspectBuilder 切面构建器
//
// 提供流式API构建切面
type AspectBuilder struct {
	pointCut PointCut
	advice   Advice
	order    int
	instance any
}

// NewAspectBuilder 创建切面构建器
func NewAspectBuilder() *AspectBuilder {
	return &AspectBuilder{
		order: 0,
	}
}

// PointCut 设置切点
func (b *AspectBuilder) PointCut(pointCut PointCut) *AspectBuilder {
	b.pointCut = pointCut
	return b
}

// Advice 设置通知
func (b *AspectBuilder) Advice(advice Advice) *AspectBuilder {
	b.advice = advice
	return b
}

// Order 设置执行顺序
func (b *AspectBuilder) Order(order int) *AspectBuilder {
	b.order = order
	return b
}

// Instance 设置切面实例
func (b *AspectBuilder) Instance(instance any) *AspectBuilder {
	b.instance = instance
	return b
}

// Before 设置前置通知
func (b *AspectBuilder) Before(fn func(JoinPoint)) *AspectBuilder {
	b.advice = Before(fn)
	return b
}

// After 设置后置通知
func (b *AspectBuilder) After(fn func(JoinPoint)) *AspectBuilder {
	b.advice = After(fn)
	return b
}

// Around 设置环绕通知
func (b *AspectBuilder) Around(fn func(JoinPoint, func() any) any) *AspectBuilder {
	b.advice = Around(fn)
	return b
}

// MatchByName 设置按名称匹配的切点
func (b *AspectBuilder) MatchByName(name string) *AspectBuilder {
	b.pointCut = MatchByName(name)
	return b
}

// MatchByRegex 设置按正则匹配的切点
func (b *AspectBuilder) MatchByRegex(pattern string) *AspectBuilder {
	b.pointCut = MatchByRegex(pattern)
	return b
}

// MatchInterface 设置按接口匹配的切点
func (b *AspectBuilder) MatchInterface(iface any) *AspectBuilder {
	b.pointCut = MatchInterface(iface)
	return b
}

// MatchAll 设置匹配所有
func (b *AspectBuilder) MatchAll() *AspectBuilder {
	b.pointCut = MatchAll()
	return b
}

// Build 构建切面元数据
func (b *AspectBuilder) Build() *AspectMeta {
	return &AspectMeta{
		PointCut: b.pointCut,
		Advice:   b.advice,
		Order:    b.order,
		Instance: b.instance,
	}
}

// BuildAndRegister 构建并注册切面
func (b *AspectBuilder) BuildAndRegister() *AspectMeta {
	aspect := b.Build()
	RegisterAspectToGlobal(aspect)
	return aspect
}

// CreateAspect 创建切面的便捷函数
func CreateAspect(pointCut PointCut, advice Advice, order int) *AspectMeta {
	return &AspectMeta{
		PointCut: pointCut,
		Advice:   advice,
		Order:    order,
	}
}

// CreateBeforeAspect 创建前置切面的便捷函数
func CreateBeforeAspect(methodName string, fn func(JoinPoint), order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), Before(fn), order)
}

// CreateAfterAspect 创建后置切面的便捷函数
func CreateAfterAspect(methodName string, fn func(JoinPoint), order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), After(fn), order)
}

// CreateAroundAspect 创建环绕切面的便捷函数
func CreateAroundAspect(methodName string, fn func(JoinPoint, func() any) any, order int) *AspectMeta {
	return CreateAspect(MatchByName(methodName), Around(fn), order)
}

// ParseAspectTarget 解析切面目标
//
// 解析类似 "UserService.GetUser" 的目标字符串
func ParseAspectTarget(target string) (structName, methodName string, err error) {
	parts := strings.Split(target, ".")
	if len(parts) != 2 {
		return "", "", errInvalidTargetFormat(target)
	}
	return parts[0], parts[1], nil
}

// CreateAspectFromTarget 从目标字符串创建切面
func CreateAspectFromTarget(target string, advice Advice, order int) (*AspectMeta, error) {
	structName, methodName, err := ParseAspectTarget(target)
	if err != nil {
		return nil, err
	}

	pointCut := MatchByName(methodName)
	return &AspectMeta{
		PointCut: pointCut,
		Advice:   advice,
		Order:    order,
		Instance: structName,
	}, nil
}

// errInvalidTargetFormat 创建无效目标格式错误
func errInvalidTargetFormat(target string) error {
	return &InvalidTargetFormatError{target: target}
}

// InvalidTargetFormatError 无效目标格式错误
type InvalidTargetFormatError struct {
	target string
}

func (e *InvalidTargetFormatError) Error() string {
	return "invalid target format: " + e.target + ", expected Struct.Method"
}

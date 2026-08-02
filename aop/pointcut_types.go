// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"
	"regexp"
	"strings"
)

// pointCutImpl 切点实现。
type pointCutImpl struct {
	classMatcher  ClassMatcher
	methodMatcher MethodMatcher
	regexPattern  string
	regex         *regexp.Regexp
	interfaceType reflect.Type
	packagePath   string
	name          string
}

// Matches 实现 PointCut 接口。
func (p *pointCutImpl) Matches(target any, methodName string) bool {
	// 如果 target 为 nil，仅检查方法匹配（通过构造虚拟 Method）
	if target == nil {
		if p.classMatcher != nil || p.interfaceType != nil {
			return false
		}
		if p.methodMatcher == nil {
			return true
		}
		// 构造一个虚拟 reflect.Method 用于匹配
		dummyMethod := reflect.Method{
			Name: methodName,
			Type: reflect.TypeOf(func() {}),
		}
		return p.methodMatcher(dummyMethod)
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	// 检查类匹配
	if p.classMatcher != nil && !p.classMatcher(targetType) {
		return false
	}

	// 检查接口匹配（同时检查值类型和指针类型）
	if p.interfaceType != nil && !targetType.Implements(p.interfaceType) && !reflect.PointerTo(targetType).Implements(p.interfaceType) {
		return false
	}

	// 如果 methodMatcher 为 nil，表示匹配所有方法
	if p.methodMatcher == nil {
		return true
	}

	// 通过反射查找方法
	method, ok := targetType.MethodByName(methodName)
	if !ok {
		// 尝试在指针类型上查找
		ptrType := reflect.PointerTo(targetType)
		method, ok = ptrType.MethodByName(methodName)
		if !ok {
			return false
		}
	}

	return p.methodMatcher(method)
}

// MatchClass 实现 PointCut 接口的类级别匹配。
func (p *pointCutImpl) MatchClass(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if p.classMatcher != nil && !p.classMatcher(t) {
		return false
	}
	if p.interfaceType != nil && !t.Implements(p.interfaceType) {
		return false
	}
	return true
}

// Expression 实现 PointCut 接口。
func (p *pointCutImpl) Expression() string {
	if p.regexPattern != "" {
		return p.regexPattern
	}
	if p.name != "" {
		return p.name
	}
	if p.packagePath != "" {
		return "package:" + p.packagePath
	}
	if p.interfaceType != nil {
		return "ByInterface(" + p.interfaceType.String() + ")"
	}
	if p.classMatcher != nil && p.methodMatcher != nil {
		return "ByClassAndMethod"
	}
	if p.classMatcher != nil {
		return "ByClass"
	}
	if p.methodMatcher != nil {
		return "ByMethod"
	}
	return "*"
}

// ClassMatcher 类匹配器类型
//
// 函数类型，接收一个反射类型，返回是否匹配。
// 用于定义切点的类级别匹配规则。
type ClassMatcher func(reflect.Type) bool

// MethodMatcher 方法匹配器类型
//
// 函数类型，接收一个反射方法，返回是否匹配。
// 用于定义切点的方法级别匹配规则。
type MethodMatcher func(reflect.Method) bool

// PointCutFunc 函数式切点适配器
//
// Go 惯用法：使用函数类型替代接口实现，符合 Go 标准库中 http.HandlerFunc 的设计模式。
// 将函数适配为 PointCut 接口，仅匹配方法（类匹配始终返回 true）。
type PointCutFunc func(reflect.Method) bool

// Matches 实现 PointCut 接口。
func (f PointCutFunc) Matches(target any, methodName string) bool {
	if target == nil {
		dummyMethod := reflect.Method{
			Name: methodName,
			Type: reflect.TypeOf(func() {}),
		}
		return f(dummyMethod)
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	method, ok := targetType.MethodByName(methodName)
	if !ok {
		ptrType := reflect.PointerTo(targetType)
		method, ok = ptrType.MethodByName(methodName)
		if !ok {
			return false
		}
	}

	return f(method)
}

// MatchClass 实现 PointCut 接口的类级别匹配。
//
// PointCutFunc 仅匹配方法，类匹配始终返回 true。
func (f PointCutFunc) MatchClass(t reflect.Type) bool {
	return true
}

// Expression 实现 PointCut 接口。
func (f PointCutFunc) Expression() string {
	return "PointCutFunc"
}

// PointCutWithClass 带类匹配的函数式切点
//
// 同时支持类和方法匹配的函数式切点适配器。
// 当需要同时指定类和方法匹配规则时使用。
type PointCutWithClass struct {
	Class ClassMatcher
	Match MethodMatcher
}

// Matches 实现 PointCut 接口。
func (p PointCutWithClass) Matches(target any, methodName string) bool {
	if target == nil {
		return true
	}

	targetType := reflect.TypeOf(target)
	if targetType.Kind() == reflect.Pointer {
		targetType = targetType.Elem()
	}

	// 检查类匹配
	if p.Class != nil && !p.Class(targetType) {
		return false
	}

	// 如果方法匹配器为 nil，表示匹配所有方法
	if p.Match == nil {
		return true
	}

	// 查找方法
	method, ok := targetType.MethodByName(methodName)
	if !ok {
		ptrType := reflect.PointerTo(targetType)
		method, ok = ptrType.MethodByName(methodName)
		if !ok {
			return false
		}
	}

	return p.Match(method)
}

// MatchClass 实现 PointCut 接口的类级别匹配。
func (p PointCutWithClass) MatchClass(t reflect.Type) bool {
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if p.Class != nil && !p.Class(t) {
		return false
	}
	return true
}

// Expression 实现 PointCut 接口。
func (p PointCutWithClass) Expression() string {
	return "PointCutWithClass"
}

// compositePointCut 组合切点实现
type compositePointCut struct {
	pointcuts []PointCut
	and       bool // true=AND, false=OR
}

// Matches 实现 PointCut 接口。
func (c *compositePointCut) Matches(target any, methodName string) bool {
	if c.and {
		// AND 逻辑：所有切点都必须匹配
		for _, pc := range c.pointcuts {
			if !pc.Matches(target, methodName) {
				return false
			}
		}
		return true
	}

	// OR 逻辑：只要有一个切点匹配即可
	for _, pc := range c.pointcuts {
		if pc.Matches(target, methodName) {
			return true
		}
	}
	return false
}

// MatchClass 实现 PointCut 接口的类级别匹配。
func (c *compositePointCut) MatchClass(t reflect.Type) bool {
	if c.and {
		for _, pc := range c.pointcuts {
			if !pc.MatchClass(t) {
				return false
			}
		}
		return true
	}
	for _, pc := range c.pointcuts {
		if pc.MatchClass(t) {
			return true
		}
	}
	return false
}

// Expression 实现 PointCut 接口。
func (c *compositePointCut) Expression() string {
	if c.and {
		return "AND(" + strings.Join(pointcutStrings(c.pointcuts), ", ") + ")"
	}
	return "OR(" + strings.Join(pointcutStrings(c.pointcuts), ", ") + ")"
}

// pointcutStrings 获取切点列表的字符串表示
func pointcutStrings(pointcuts []PointCut) []string {
	result := make([]string, len(pointcuts))
	for i, pc := range pointcuts {
		result[i] = pc.Expression()
	}
	return result
}

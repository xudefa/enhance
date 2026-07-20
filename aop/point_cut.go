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
}

func (p *pointCutImpl) MatchClass(c reflect.Type) bool {
	if p.classMatcher == nil {
		return true
	}
	return p.classMatcher(c)
}

func (p *pointCutImpl) MatchMethod(m reflect.Method) bool {
	if p.methodMatcher == nil {
		return true
	}
	return p.methodMatcher(m)
}

func (p *pointCutImpl) String() string {
	if p.regexPattern != "" {
		return "ByRegex(" + p.regexPattern + ")"
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
	return "MatchAll"
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

// MatchClass 实现 PointCut 接口（始终返回 true）
func (f PointCutFunc) MatchClass(c reflect.Type) bool {
	return true
}

// MatchMethod 实现 PointCut 接口
func (f PointCutFunc) MatchMethod(m reflect.Method) bool {
	return f(m)
}

// String 实现 PointCut 接口
func (f PointCutFunc) String() string {
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

// MatchClass 实现 PointCut 接口
func (p PointCutWithClass) MatchClass(c reflect.Type) bool {
	if p.Class == nil {
		return true
	}
	return p.Class(c)
}

// MatchMethod 实现 PointCut 接口
func (p PointCutWithClass) MatchMethod(m reflect.Method) bool {
	if p.Match == nil {
		return true
	}
	return p.Match(m)
}

// String 实现 PointCut 接口
func (p PointCutWithClass) String() string {
	return "PointCutWithClass"
}

// MatchAll 匹配所有
//
// 返回匹配所有类和方法的切点。
//
// 返回值:
//   - PointCut: 匹配所有目标的切点
//
// 示例:
//
//	// 拦截所有方法
//	aop.MatchAll()
func MatchAll() PointCut {
	return &pointCutImpl{
		classMatcher:  nil,
		methodMatcher: nil,
	}
}

// MatchClass 匹配类
//
// 返回只匹配类的切点，不匹配具体方法。
//
// 参数:
//   - matcher: 类匹配函数
//
// 返回值:
//   - PointCut: 匹配给定类的切点
//
// 示例:
//
//	aop.MatchClass(func(t reflect.Type) bool {
//	    return t.Name() == "UserService"
//	})
func MatchClass(matcher ClassMatcher) PointCut {
	return &pointCutImpl{
		classMatcher:  matcher,
		methodMatcher: nil,
	}
}

// MatchMethod 匹配方法
//
// 返回只匹配方法的切点，不匹配具体类。
//
// 参数:
//   - matcher: 方法匹配函数
//
// 返回值:
//   - PointCut: 匹配给定方法的切点
//
// 示例:
//
//	aop.MatchMethod(func(m reflect.Method) bool {
//	    return m.Name == "DoSomething"
//	})
func MatchMethod(matcher MethodMatcher) PointCut {
	return &pointCutImpl{
		classMatcher:  nil,
		methodMatcher: matcher,
	}
}

// MatchClassMethod 匹配类和方法的组合切点
//
// 同时指定类和方法匹配条件。
//
// 参数:
//   - classMatcher: 类匹配函数
//   - methodMatcher: 方法匹配函数
//
// 返回值:
//   - PointCut: 组合切点
func MatchClassMethod(classMatcher ClassMatcher, methodMatcher MethodMatcher) PointCut {
	return &pointCutImpl{
		classMatcher:  classMatcher,
		methodMatcher: methodMatcher,
	}
}

// MatchByName 按方法名匹配
//
// 匹配指定名称的方法。
//
// 参数:
//   - name: 方法名
//
// 返回值:
//   - PointCut: 匹配指定方法名的切点
//
// 示例:
//
//	// 只拦截 DoSomething 方法
//	aop.MatchByName("DoSomething")
func MatchByName(name string) PointCut {
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return m.Name == name
		},
	}
}

// MatchByNamePrefix 按方法名前缀匹配
//
// 匹配指定前缀的方法。
//
// 参数:
//   - prefix: 方法名前缀
//
// 返回值:
//   - PointCut: 匹配指定前缀的切点
//
// 示例:
//
//	// 拦截所有以 Do 开头的方法
//	aop.MatchByNamePrefix("Do")
func MatchByNamePrefix(prefix string) PointCut {
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return strings.HasPrefix(m.Name, prefix)
		},
	}
}

// MatchByRegex 按正则表达式匹配
//
// 匹配符合正则表达式的方法名。
//
// 参数:
//   - pattern: 正则表达式
//
// 返回值:
//   - PointCut: 匹配正则表达式的切点
//
// 示例:
//
//	// 拦截所有以 do 或 Do 开头的方法
//	aop.MatchByRegex("(?i)^do.*")
func MatchByRegex(pattern string) PointCut {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return &pointCutImpl{
			methodMatcher: func(m reflect.Method) bool { return false },
			regexPattern:  pattern,
		}
	}
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return re.MatchString(m.Name)
		},
		regexPattern: pattern,
		regex:        re,
	}
}

// MatchByAnnotation 按注解类型匹配
//
// 匹配带有指定注解类型的方法。
//
// 参数:
//   - annotationType: 注解类型
//
// 返回值:
//   - PointCut: 匹配带注解方法的切点
//
// 注意:
//   - 此方法通过方法名前缀来匹配
func MatchByAnnotation(annotationType reflect.Type) PointCut {
	annotationName := annotationType.Name()
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return strings.HasPrefix(m.Name, annotationName+"_") ||
				strings.Contains(m.Name, "_"+annotationName+"_") ||
				strings.HasSuffix(m.Name, "_"+annotationName)
		},
	}
}

// MatchInterface 按接口类型匹配
//
// 匹配实现了指定接口的类型。
//
// 参数:
//   - y: 接口类型，传入接口变量即可
//
// 返回值:
//   - PointCut: 匹配实现接口的类的切点
//
// 示例:
//
//	// 拦截所有实现 ServiceInterface 接口的类
//	aop.MatchInterface((*ServiceInterface)(nil))
func MatchInterface(y any) PointCut {
	yType := reflect.TypeOf(y)
	if yType == nil {
		return &pointCutImpl{
			classMatcher:  func(t reflect.Type) bool { return false },
			methodMatcher: func(m reflect.Method) bool { return false },
		}
	}
	if yType.Kind() == reflect.Pointer {
		yType = yType.Elem()
	}

	if yType.Kind() != reflect.Interface {
		return &pointCutImpl{
			classMatcher: func(t reflect.Type) bool { return false },
		}
	}

	return &pointCutImpl{
		interfaceType: yType,
		classMatcher: func(t reflect.Type) bool {
			if t == nil {
				return false
			}
			for t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			return t.Implements(yType)
		},
	}
}

// MatchByMethodSignature 按方法签名匹配
//
// 匹配具有指定方法签名的方法（方法名 + 参数类型）。
// 这是最精确的匹配方式，类似 Spring 的 execution 表达式。
//
// 参数:
//   - methodName: 方法名
//   - paramTypes: 参数类型列表（可选，nil 表示只匹配方法名）
//
// 返回值:
//   - PointCut: 匹配指定方法签名的切点
//
// 示例:
//
//	// 精确匹配 GetUser(id int64) 方法
//	aop.MatchByMethodSignature("GetUser", reflect.TypeOf(int64(0)))
//
//	// 匹配所有名为 Save 的方法（不限参数）
//	aop.MatchByMethodSignature("Save", nil)
func MatchByMethodSignature(methodName string, paramTypes ...reflect.Type) PointCut {
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			// 首先匹配方法名
			if m.Name != methodName {
				return false
			}

			// 如果指定了参数类型，需要精确匹配
			if len(paramTypes) > 0 {
				// m.Type.NumIn() 包含 receiver，所以实际参数数量是 NumIn()-1
				actualParamCount := m.Type.NumIn() - 1
				if actualParamCount != len(paramTypes) {
					return false
				}

				for i, paramType := range paramTypes {
					// m.Type.In(0) 是 receiver，从 In(1) 开始是实际参数
					actualType := m.Type.In(i + 1)
					if actualType != paramType {
						return false
					}
				}
			}

			return true
		},
	}
}

// MatchByReturnType 按返回值类型匹配
//
// 匹配返回指定类型的方法。
//
// 参数:
//   - returnType: 返回值类型
//
// 返回值:
//   - PointCut: 匹配指定返回值类型的切点
//
// 示例:
//
//	// 匹配所有返回 error 的方法
//	aop.MatchByReturnType(reflect.TypeOf((*error)(nil)).Elem())
func MatchByReturnType(returnType reflect.Type) PointCut {
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			if m.Type.NumOut() == 0 {
				return false
			}

			// 检查第一个返回值类型
			firstReturnType := m.Type.Out(0)

			// 精确匹配
			if firstReturnType == returnType {
				return true
			}

			// 如果是接口类型，检查是否实现
			if returnType.Kind() == reflect.Interface {
				return firstReturnType.Implements(returnType)
			}

			return false
		},
	}
}

// MatchByPackage 按包路径匹配
//
// 匹配指定包路径下的所有类和方法。
//
// 参数:
//   - packagePath: 包路径前缀（如 "github.com/myapp/service"）
//
// 返回值:
//   - PointCut: 匹配指定包的切点
//
// 示例:
//
//	// 拦截 service 包下的所有方法
//	aop.MatchByPackage("github.com/myapp/service")
func MatchByPackage(packagePath string) PointCut {
	return &pointCutImpl{
		classMatcher: func(t reflect.Type) bool {
			if t == nil {
				return false
			}
			return strings.HasPrefix(t.PkgPath(), packagePath)
		},
	}
}

// MatchByClassName 按类名匹配
//
// 匹配指定类名的所有方法。
//
// 参数:
//   - className: 类名（支持通配符 *）
//
// 返回值:
//   - PointCut: 匹配指定类名的切点
//
// 示例:
//
//	// 匹配所有 Service 结尾的类
//	aop.MatchByClassName("*Service")
//
//	// 精确匹配 UserService 类
//	aop.MatchByClassName("UserService")
func MatchByClassName(className string) PointCut {
	// 如果包含通配符，使用正则匹配
	if strings.Contains(className, "*") {
		// 将通配符转换为正则表达式
		pattern := "^" + strings.ReplaceAll(strings.ReplaceAll(className, "*", ".*"), "?", ".") + "$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			return &pointCutImpl{
				classMatcher: func(t reflect.Type) bool { return false },
			}
		}

		return &pointCutImpl{
			classMatcher: func(t reflect.Type) bool {
				if t == nil {
					return false
				}
				// 如果是指针类型，获取元素类型
				if t.Kind() == reflect.Pointer {
					t = t.Elem()
				}
				return re.MatchString(t.Name())
			},
		}
	}

	// 精确匹配
	return &pointCutImpl{
		classMatcher: func(t reflect.Type) bool {
			if t == nil {
				return false
			}
			// 如果是指针类型，获取元素类型
			if t.Kind() == reflect.Pointer {
				t = t.Elem()
			}
			return t.Name() == className
		},
	}
}

// Compose 组合多个切点（AND 逻辑）
//
// 只有当所有切点都匹配时，才认为匹配。
//
// 参数:
//   - pointcuts: 切点列表
//
// 返回值:
//   - PointCut: 组合后的切点
//
// 示例:
//
//	// 匹配 Service 类中以 Get 开头的方法
//	aop.Compose(
//	    aop.MatchByClassName("*Service"),
//	    aop.MatchByNamePrefix("Get"),
//	)
func Compose(pointcuts ...PointCut) PointCut {
	return &compositePointCut{
		pointcuts: pointcuts,
		and:       true,
	}
}

// ComposeOr 组合多个切点（OR 逻辑）
//
// 只要有一个切点匹配，就认为匹配。
//
// 参数:
//   - pointcuts: 切点列表
//
// 返回值:
//   - PointCut: 组合后的切点
//
// 示例:
//
//	// 匹配 GetUser 或 UpdateUser 方法
//	aop.ComposeOr(
//	    aop.MatchByName("GetUser"),
//	    aop.MatchByName("UpdateUser"),
//	)
func ComposeOr(pointcuts ...PointCut) PointCut {
	return &compositePointCut{
		pointcuts: pointcuts,
		and:       false,
	}
}

// compositePointCut 组合切点实现
type compositePointCut struct {
	pointcuts []PointCut
	and       bool // true=AND, false=OR
}

func (c *compositePointCut) MatchClass(t reflect.Type) bool {
	if c.and {
		// AND 逻辑：所有切点都必须匹配类
		for _, pc := range c.pointcuts {
			if !pc.MatchClass(t) {
				return false
			}
		}
		return true
	}

	// OR 逻辑：只要有一个切点匹配类即可
	for _, pc := range c.pointcuts {
		if pc.MatchClass(t) {
			return true
		}
	}
	return false
}

func (c *compositePointCut) MatchMethod(m reflect.Method) bool {
	if c.and {
		// AND 逻辑：所有切点都必须匹配方法
		for _, pc := range c.pointcuts {
			if !pc.MatchMethod(m) {
				return false
			}
		}
		return true
	}

	// OR 逻辑：只要有一个切点匹配方法即可
	for _, pc := range c.pointcuts {
		if pc.MatchMethod(m) {
			return true
		}
	}
	return false
}

func (c *compositePointCut) String() string {
	if c.and {
		return "AND(" + strings.Join(pointcutStrings(c.pointcuts), ", ") + ")"
	}
	return "OR(" + strings.Join(pointcutStrings(c.pointcuts), ", ") + ")"
}

// pointcutStrings 获取切点列表的字符串表示
func pointcutStrings(pointcuts []PointCut) []string {
	result := make([]string, len(pointcuts))
	for i, pc := range pointcuts {
		result[i] = pc.String()
	}
	return result
}

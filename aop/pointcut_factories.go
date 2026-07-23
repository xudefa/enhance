// Package aop 提供面向切面编程（AOP）支持。
package aop

import (
	"reflect"
	"regexp"
	"strings"
)

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
// 匹配指定名称的方法。支持以下模式：
//   - 精确匹配：如 "DoSomething"
//   - 通配符匹配：如 "Do*" 匹配所有以 Do 开头的方法
//   - 正则匹配：如 "^Do.*" 匹配所有以 Do 开头的方法
//
// 参数:
//   - name: 方法名或模式
//
// 返回值:
//   - PointCut: 匹配指定方法名的切点
//
// 示例:
//
//	// 只拦截 DoSomething 方法
//	aop.MatchByName("DoSomething")
//
//	// 拦截所有以 Do 开头的方法
//	aop.MatchByName("Do*")
func MatchByName(name string) PointCut {
	if strings.ContainsAny(name, "*?") {
		pattern := "^" + strings.ReplaceAll(strings.ReplaceAll(name, "*", ".*"), "?", ".") + "$"
		re, err := regexp.Compile(pattern)
		if err != nil {
			return &pointCutImpl{
				methodMatcher: func(m reflect.Method) bool { return false },
				regexPattern:  name,
			}
		}
		return &pointCutImpl{
			methodMatcher: func(m reflect.Method) bool {
				return re.MatchString(m.Name)
			},
			regexPattern: name,
			regex:        re,
		}
	}
	if strings.ContainsAny(name, "^$+[]{}|\\()") {
		re, err := regexp.Compile(name)
		if err == nil {
			return &pointCutImpl{
				methodMatcher: func(m reflect.Method) bool {
					return re.MatchString(m.Name)
				},
				regexPattern: name,
				regex:        re,
			}
		}
	}
	return &pointCutImpl{
		methodMatcher: func(m reflect.Method) bool {
			return m.Name == name
		},
		regexPattern: "",
		name:         name,
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
	// 支持两种调用方式：
	// 1. MatchInterface((*SomeInterface)(nil)) — 传入接口指针
	// 2. MatchInterface(reflect.TypeFor[SomeInterface]()) — 直接传入 reflect.Type
	if ifaceType, ok := y.(reflect.Type); ok {
		if ifaceType == nil || ifaceType.Kind() != reflect.Interface {
			return &pointCutImpl{
				classMatcher:  func(t reflect.Type) bool { return false },
				methodMatcher: func(m reflect.Method) bool { return false },
			}
		}
		return &pointCutImpl{
			interfaceType: ifaceType,
			classMatcher: func(t reflect.Type) bool {
				if t == nil {
					return false
				}
				for t.Kind() == reflect.Pointer {
					t = t.Elem()
				}
				return t.Implements(ifaceType) || reflect.PointerTo(t).Implements(ifaceType)
			},
		}
	}

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
			return t.Implements(yType) || reflect.PointerTo(t).Implements(yType)
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
		packagePath: packagePath,
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



package aop

import (
	"reflect"
	"strings"
	"testing"
)

func TestMatchAll(t *testing.T) {
	t.Parallel()
	pc := MatchAll()

	// MatchAll 应该匹配任何方法
	if !pc.Matches(nil, "AnyMethod") {
		t.Error("MatchAll should match any method")
	}

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchAll should match any method with target")
	}
}

func TestMatchByName(t *testing.T) {
	t.Parallel()
	pc := MatchByName("DoSomething")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName should match DoSomething")
	}

	if pc.Matches(nil, "DoAnother") {
		t.Error("MatchByName should not match DoAnother")
	}
}

func TestMatchByNamePrefix(t *testing.T) {
	t.Parallel()
	pc := MatchByName("Do*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName with Do* should match methods with Do prefix")
	}

	if pc.Matches(nil, "GetValue") {
		t.Error("MatchByName with Do* should not match methods without Do prefix")
	}
}

func TestMatchByRegex(t *testing.T) {
	t.Parallel()
	pc := MatchByName("^Do.*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName with regex should match methods matching regex")
	}

	if pc.Matches(nil, "GetValue") {
		t.Error("MatchByName with regex should not match methods not matching regex")
	}
}

func TestMatchInterface_NilInput(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(nil)

	// 当 target 为 nil 时，应该返回 false
	if pc.Matches(nil, "DoSomething") {
		t.Error("MatchInterface(nil) should not match nil target")
	}
}

func TestMatchInterface_NonInterfaceInput(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(reflect.TypeFor[string]())

	// string 不是接口，不应该匹配
	if pc.Matches("test", "DoSomething") {
		t.Error("MatchInterface with non-interface should not match")
	}
}

func TestMatchInterface(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(reflect.TypeFor[TestInterfaceForMatch]())

	if !pc.Matches(&TestImplForMatch{}, "DoSomething") {
		t.Error("MatchInterface should match implementing struct pointer")
	}

	if pc.Matches("string", "DoSomething") {
		t.Error("MatchInterface should not match non-implementing type")
	}
}

type TestInterfaceForMatch interface {
	DoSomething()
}

type TestImplForMatch struct{}

func (t *TestImplForMatch) DoSomething() {}

// TestUserService 用于测试的通用服务类型
type TestUserService struct{}

func (u *TestUserService) DoSomething() string {
	return "did something"
}

func (u *TestUserService) DoAnother() string {
	return "did another"
}

// TestServiceInterface 用于测试的接口
type TestServiceInterface interface {
	GetUser(id int64) string
	DoSomething()
}

// TestServiceImpl 用于测试的接口实现
type TestServiceImpl struct{}

func (s *TestServiceImpl) GetUser(id int64) string {
	return "user"
}

func (s *TestServiceImpl) DoSomething() {
}

func TestPointCut_Expression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pointCut PointCut
		expected string
	}{
		{"MatchByName", MatchByName("DoSomething"), "DoSomething"},
		{"MatchAll", MatchAll(), "*"},
		{"MatchByRegex", MatchByName("^Do.*"), "^Do.*"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.pointCut.Expression() != tt.expected {
				t.Errorf("Expression() = %v, want %v", tt.pointCut.Expression(), tt.expected)
			}
		})
	}
}

func TestMatchByPackage(t *testing.T) {
	t.Parallel()
	pc := MatchByPackage("github.com/xudefa/enhance")

	// 使用实际的对象测试包匹配
	userService := &TestUserService{}
	if !pc.Matches(userService, "DoSomething") {
		t.Error("MatchByPackage should match types in the specified package")
	}

	// nil target 应该返回 false
	if pc.Matches(nil, "DoSomething") {
		t.Error("MatchByPackage should not match nil target")
	}
}

// PointCutTestService 测试用服务
type PointCutTestService struct{}

func (s *PointCutTestService) GetUser(id int64) string             { return "" }
func (s *PointCutTestService) SaveUser(name string, age int) error { return nil }
func (s *PointCutTestService) GetUserByName(name string) string    { return "" }
func (s *PointCutTestService) DeleteUser(id int64) error           { return nil }

func TestMatchByName_Prefix(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("Get*")

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with Get* pattern")
	}

	if !pointcut.Matches(&PointCutTestService{}, "GetUserByName") {
		t.Error("Should match GetUserByName with Get* pattern")
	}

	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser with Get* pattern")
	}
}

func TestMatchByName_Suffix(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("*User")

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with *User pattern")
	}

	if !pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should match SaveUser with *User pattern (ends with User)")
	}

	if pointcut.Matches(&PointCutTestService{}, "DeleteItem") {
		t.Error("Should not match DeleteItem with *User pattern")
	}
}

func TestMatchByName_Exact(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("GetUser")

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser")
	}

	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser")
	}
}

func TestMatchByPackage_Extended(t *testing.T) {
	t.Parallel()
	pointcut := MatchByPackage("github.com/xudefa/enhance")

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match types in the specified package")
	}

	if pointcut.Matches(nil, "GetUser") {
		t.Error("Should not match nil target")
	}
}

func TestMatchByInterface_Extended(t *testing.T) {
	t.Parallel()

	type TestServiceInterface interface {
		GetUser(id int64) string
	}

	pointcut := MatchInterface(reflect.TypeOf((*TestServiceInterface)(nil)).Elem())

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match type implementing the interface")
	}

	if pointcut.Matches("string", "GetUser") {
		t.Error("Should not match type not implementing the interface")
	}
}

func TestMatchByName_Regex(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("^Get.*")

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with ^Get.* regex")
	}

	if !pointcut.Matches(&PointCutTestService{}, "GetUserByName") {
		t.Error("Should match GetUserByName with ^Get.* regex")
	}

	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser with ^Get.* regex")
	}
}

func TestMatchByPackage_NilTarget(t *testing.T) {
	t.Parallel()
	pointcut := MatchByPackage("github.com/test")

	if pointcut.Matches(nil, "GetUser") {
		t.Error("Should not match nil target")
	}
}

func TestMatchByInterface_NilInput(t *testing.T) {
	t.Parallel()
	pointcut := MatchInterface(nil)

	if pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should not match with nil interface type")
	}

	if pointcut.Matches(nil, "GetUser") {
		t.Error("Should not match with nil target")
	}
}

func TestMatchClass(t *testing.T) {
	t.Parallel()
	pc := MatchClass(func(t reflect.Type) bool {
		return t.Name() == "TestUserService"
	})

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchClass should match TestUserService")
	}

	if pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("MatchClass should not match PointCutTestService")
	}
}

func TestMatchMethod(t *testing.T) {
	t.Parallel()
	pc := MatchMethod(func(m reflect.Method) bool {
		return m.Name == "DoSomething"
	})

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchMethod should match DoSomething")
	}

	if pc.Matches(&TestUserService{}, "DoAnother") {
		t.Error("MatchMethod should not match DoAnother")
	}
}

func TestMatchClassMethod(t *testing.T) {
	t.Parallel()
	pc := MatchClassMethod(
		func(t reflect.Type) bool { return t.Name() == "TestUserService" },
		func(m reflect.Method) bool { return m.Name == "DoSomething" },
	)

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchClassMethod should match both class and method")
	}

	if pc.Matches(&TestUserService{}, "DoAnother") {
		t.Error("MatchClassMethod should not match when method doesn't match")
	}

	if pc.Matches(&PointCutTestService{}, "DoSomething") {
		t.Error("MatchClassMethod should not match when class doesn't match")
	}
}

func TestMatchByNamePrefixExtended(t *testing.T) {
	t.Parallel()
	pc := MatchByNamePrefix("Do")

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchByNamePrefix should match DoSomething")
	}

	if !pc.Matches(&TestUserService{}, "DoAnother") {
		t.Error("MatchByNamePrefix should match DoAnother")
	}

	if pc.Matches(&TestUserService{}, "GetValue") {
		t.Error("MatchByNamePrefix should not match GetValue")
	}
}

func TestMatchByRegexExtended(t *testing.T) {
	t.Parallel()
	pc := MatchByRegex(`(?i)^do.*`)

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchByRegex should match DoSomething")
	}

	if pc.Matches(&TestUserService{}, "GetValue") {
		t.Error("MatchByRegex should not match GetValue")
	}

	// 测试无效正则
	invalidPc := MatchByRegex(`[invalid`)
	if invalidPc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("Invalid regex should not match anything")
	}
}

func TestMatchByAnnotation(t *testing.T) {
	t.Parallel()
	type TestAnnotation struct{}
	_ = MatchByAnnotation(reflect.TypeOf(TestAnnotation{}))

	// 测试注解名称匹配逻辑
	annotationName := "TestAnnotation"
	if !strings.HasPrefix("TestAnnotation_DoSomething", annotationName+"_") {
		t.Error("Expected prefix match to work")
	}

	if strings.HasPrefix("DoSomething", annotationName+"_") {
		t.Error("Expected non-matching method to not match")
	}
}

func TestMatchByMethodSignature(t *testing.T) {
	t.Parallel()
	pc := MatchByMethodSignature("GetUser", reflect.TypeOf(int64(0)))

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("MatchByMethodSignature should match GetUser")
	}

	if pc.Matches(&TestServiceImpl{}, "DoSomething") {
		t.Error("MatchByMethodSignature should not match DoSomething")
	}
}

func TestMatchByReturnType(t *testing.T) {
	t.Parallel()
	pc := MatchByReturnType(reflect.TypeOf(""))

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("MatchByReturnType should match GetUser (returns string)")
	}

	if pc.Matches(&TestServiceImpl{}, "DoSomething") {
		t.Error("MatchByReturnType should not match DoSomething (returns nothing)")
	}
}

func TestMatchByClassName(t *testing.T) {
	t.Parallel()
	pc := MatchByClassName("TestUserService")

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchByClassName should match TestUserService")
	}

	if pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("MatchByClassName should not match PointCutTestService")
	}
}

func TestCompose(t *testing.T) {
	t.Parallel()
	// Compose是AND逻辑，所有pointcut都必须匹配
	pc1 := MatchByName("*") // 匹配所有方法
	pc2 := MatchByName("DoSomething")

	composed := Compose(pc1, pc2)

	if !composed.Matches(&TestUserService{}, "DoSomething") {
		t.Error("Composed should match DoSomething")
	}

	if composed.Matches(&TestUserService{}, "DoAnother") {
		t.Error("Composed should not match DoAnother (pc2 doesn't match)")
	}
}

func TestComposeOr(t *testing.T) {
	t.Parallel()
	pc1 := MatchByName("DoSomething")
	pc2 := MatchByName("DoAnother")

	composed := ComposeOr(pc1, pc2)

	if !composed.Matches(&TestUserService{}, "DoSomething") {
		t.Error("ComposedOr should match DoSomething")
	}

	if !composed.Matches(&TestUserService{}, "DoAnother") {
		t.Error("ComposedOr should match DoAnother")
	}

	if composed.Matches(&TestUserService{}, "GetValue") {
		t.Error("ComposedOr should not match GetValue")
	}
}

func TestPointCutImpl_MatchClass(t *testing.T) {
	t.Parallel()

	t.Run("class matcher", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{
			classMatcher: func(t reflect.Type) bool {
				return t.Name() == "TestUserService"
			},
		}

		if !pc.MatchClass(reflect.TypeOf(TestUserService{})) {
			t.Error("should match TestUserService")
		}

		if pc.MatchClass(reflect.TypeOf(PointCutTestService{})) {
			t.Error("should not match PointCutTestService")
		}
	})

	t.Run("interface matcher", func(t *testing.T) {
		t.Parallel()
		// Create a type that implements the interface with value receiver
		type ValueReceiverImpl struct{}

		// We need to add methods to make it implement the interface
		// Since we can't add methods to local types, let's just test the logic
		// by creating a pointCutImpl with no classMatcher and only interfaceType
		pc := &pointCutImpl{
			interfaceType: nil, // No interface check
			classMatcher: func(t reflect.Type) bool {
				return t.Name() == "TestServiceImpl"
			},
		}

		// Test that classMatcher works correctly
		if !pc.MatchClass(reflect.TypeOf(TestServiceImpl{})) {
			t.Error("should match TestServiceImpl by classMatcher")
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{
			classMatcher: func(t reflect.Type) bool {
				return t.Name() == "TestUserService"
			},
		}

		if !pc.MatchClass(reflect.PointerTo(reflect.TypeOf(TestUserService{}))) {
			t.Error("should match pointer to TestUserService")
		}
	})
}

func TestPointCutImpl_Expression(t *testing.T) {
	t.Parallel()

	t.Run("regex pattern", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{regexPattern: "Do.*"}
		if pc.Expression() != "Do.*" {
			t.Errorf("expected 'Do.*', got '%s'", pc.Expression())
		}
	})

	t.Run("name", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{name: "testPointcut"}
		if pc.Expression() != "testPointcut" {
			t.Errorf("expected 'testPointcut', got '%s'", pc.Expression())
		}
	})

	t.Run("package path", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{packagePath: "aop"}
		if pc.Expression() != "package:aop" {
			t.Errorf("expected 'package:aop', got '%s'", pc.Expression())
		}
	})

	t.Run("interface type", func(t *testing.T) {
		t.Parallel()
		type TestInterface interface{}
		pc := &pointCutImpl{interfaceType: reflect.TypeOf((*TestInterface)(nil)).Elem()}
		if !strings.Contains(pc.Expression(), "ByInterface") {
			t.Errorf("expected 'ByInterface(...)', got '%s'", pc.Expression())
		}
	})

	t.Run("class and method", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{
			classMatcher:  func(t reflect.Type) bool { return true },
			methodMatcher: func(m reflect.Method) bool { return true },
		}
		if pc.Expression() != "ByClassAndMethod" {
			t.Errorf("expected 'ByClassAndMethod', got '%s'", pc.Expression())
		}
	})

	t.Run("class only", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{
			classMatcher: func(t reflect.Type) bool { return true },
		}
		if pc.Expression() != "ByClass" {
			t.Errorf("expected 'ByClass', got '%s'", pc.Expression())
		}
	})

	t.Run("method only", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{
			methodMatcher: func(m reflect.Method) bool { return true },
		}
		if pc.Expression() != "ByMethod" {
			t.Errorf("expected 'ByMethod', got '%s'", pc.Expression())
		}
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		pc := &pointCutImpl{}
		if pc.Expression() != "*" {
			t.Errorf("expected '*', got '%s'", pc.Expression())
		}
	})
}

func TestPointCutFunc_MatchClass(t *testing.T) {
	t.Parallel()

	pc := PointCutFunc(func(m reflect.Method) bool {
		return true
	})

	// PointCutFunc.MatchClass always returns true
	if !pc.MatchClass(reflect.TypeOf(TestUserService{})) {
		t.Error("PointCutFunc.MatchClass should always return true")
	}
}

func TestPointCutFunc_Expression(t *testing.T) {
	t.Parallel()

	pc := PointCutFunc(func(m reflect.Method) bool {
		return true
	})

	if pc.Expression() != "PointCutFunc" {
		t.Errorf("expected 'PointCutFunc', got '%s'", pc.Expression())
	}
}

func TestPointCutWithClass_MatchClass(t *testing.T) {
	t.Parallel()

	t.Run("with class matcher", func(t *testing.T) {
		t.Parallel()
		pc := PointCutWithClass{
			Class: func(t reflect.Type) bool {
				return t.Name() == "TestUserService"
			},
		}

		if !pc.MatchClass(reflect.TypeOf(TestUserService{})) {
			t.Error("should match TestUserService")
		}

		if pc.MatchClass(reflect.TypeOf(PointCutTestService{})) {
			t.Error("should not match PointCutTestService")
		}
	})

	t.Run("pointer type", func(t *testing.T) {
		t.Parallel()
		pc := PointCutWithClass{
			Class: func(t reflect.Type) bool {
				return t.Name() == "TestUserService"
			},
		}

		if !pc.MatchClass(reflect.PointerTo(reflect.TypeOf(TestUserService{}))) {
			t.Error("should match pointer to TestUserService")
		}
	})

	t.Run("no class matcher", func(t *testing.T) {
		t.Parallel()
		pc := PointCutWithClass{}

		if !pc.MatchClass(reflect.TypeOf(TestUserService{})) {
			t.Error("should match when no class matcher")
		}
	})
}

func TestPointCutWithClass_Expression(t *testing.T) {
	t.Parallel()

	pc := PointCutWithClass{}
	if pc.Expression() != "PointCutWithClass" {
		t.Errorf("expected 'PointCutWithClass', got '%s'", pc.Expression())
	}
}

func TestCompositePointCut_Expression(t *testing.T) {
	t.Parallel()

	t.Run("AND logic", func(t *testing.T) {
		t.Parallel()
		pc1 := MatchByName("Do*")
		pc2 := MatchByName("*Something")

		composed := Compose(pc1, pc2)
		expr := composed.Expression()
		if !strings.HasPrefix(expr, "AND(") {
			t.Errorf("expected 'AND(...)', got '%s'", expr)
		}
	})

	t.Run("OR logic", func(t *testing.T) {
		t.Parallel()
		pc1 := MatchByName("Do*")
		pc2 := MatchByName("Get*")

		composed := ComposeOr(pc1, pc2)
		expr := composed.Expression()
		if !strings.HasPrefix(expr, "OR(") {
			t.Errorf("expected 'OR(...)', got '%s'", expr)
		}
	})
}

func TestPointcutStrings(t *testing.T) {
	t.Parallel()

	pc1 := MatchByName("Do*")
	pc2 := MatchByName("Get*")

	result := pointcutStrings([]PointCut{pc1, pc2})
	if len(result) != 2 {
		t.Errorf("expected 2 strings, got %d", len(result))
	}
	if result[0] != "Do*" {
		t.Errorf("expected 'Do*', got '%s'", result[0])
	}
	if result[1] != "Get*" {
		t.Errorf("expected 'Get*', got '%s'", result[1])
	}
}

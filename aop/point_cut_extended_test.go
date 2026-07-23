package aop

import (
	"reflect"
	"testing"
)

// PointCutTestService 测试用服务
type PointCutTestService struct{}

func (s *PointCutTestService) GetUser(id int64) string             { return "" }
func (s *PointCutTestService) SaveUser(name string, age int) error { return nil }
func (s *PointCutTestService) GetUserByName(name string) string    { return "" }
func (s *PointCutTestService) DeleteUser(id int64) error           { return nil }

// TestMatchByName_Prefix 测试前缀匹配
func TestMatchByName_Prefix(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("Get*")

	// 测试匹配 GetUser
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with Get* pattern")
	}

	// 测试匹配 GetUserByName
	if !pointcut.Matches(&PointCutTestService{}, "GetUserByName") {
		t.Error("Should match GetUserByName with Get* pattern")
	}

	// 测试不匹配 SaveUser
	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser with Get* pattern")
	}
}

// TestMatchByName_Suffix 测试后缀匹配
func TestMatchByName_Suffix(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("*User")

	// 测试匹配 GetUser
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with *User pattern")
	}

	// 测试匹配 SaveUser（也以 User 结尾）
	if !pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should match SaveUser with *User pattern (ends with User)")
	}

	// 测试不匹配 DeleteUser
	if pointcut.Matches(&PointCutTestService{}, "DeleteItem") {
		t.Error("Should not match DeleteItem with *User pattern")
	}
}

// TestMatchByName_Exact 测试精确匹配
func TestMatchByName_Exact(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("GetUser")

	// 测试匹配 GetUser
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser")
	}

	// 测试不匹配 SaveUser
	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser")
	}
}

// TestMatchByPackage 测试包路径匹配
func TestMatchByPackage_Extended(t *testing.T) {
	t.Parallel()
	pointcut := MatchByPackage("github.com/xudefa/enhance")

	// 测试匹配当前包的对象
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match types in the specified package")
	}

	// nil target 应该返回 false
	if pointcut.Matches(nil, "GetUser") {
		t.Error("Should not match nil target")
	}
}

// TestMatchByInterface_Extended 测试接口匹配
func TestMatchByInterface_Extended(t *testing.T) {
	t.Parallel()

	type TestServiceInterface interface {
		GetUser(id int64) string
	}

	pointcut := MatchInterface(reflect.TypeOf((*TestServiceInterface)(nil)).Elem())

	// 测试匹配实现接口的对象
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match type implementing the interface")
	}

	// 测试不匹配未实现接口的对象
	if pointcut.Matches("string", "GetUser") {
		t.Error("Should not match type not implementing the interface")
	}
}

// TestPointCut_Expression 测试切点表达式
func TestPointCut_Expression_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		pointCut PointCut
		expected string
	}{
		{"MatchByName", MatchByName("GetUser"), "GetUser"},
		{"MatchAll", MatchAll(), "*"},
		{"MatchByPackage", MatchByPackage("github.com/test"), "package:github.com/test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.pointCut.Expression() != tt.expected {
				t.Errorf("Expression() = %v, want %v", tt.pointCut.Expression(), tt.expected)
			}
		})
	}
}

// TestMatchAll_Extended 测试 MatchAll
func TestMatchAll_Extended(t *testing.T) {
	t.Parallel()
	pointcut := MatchAll()

	// 应该匹配任何方法
	if !pointcut.Matches(nil, "AnyMethod") {
		t.Error("Should match any method name")
	}

	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match any method with target")
	}
}

// TestMatchByName_Regex 测试正则匹配
func TestMatchByName_Regex(t *testing.T) {
	t.Parallel()
	pointcut := MatchByName("^Get.*")

	// 测试匹配 GetUser
	if !pointcut.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("Should match GetUser with ^Get.* regex")
	}

	// 测试匹配 GetUserByName
	if !pointcut.Matches(&PointCutTestService{}, "GetUserByName") {
		t.Error("Should match GetUserByName with ^Get.* regex")
	}

	// 测试不匹配 SaveUser
	if pointcut.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("Should not match SaveUser with ^Get.* regex")
	}
}

// TestMatchByPackage_NilTarget 测试 nil target
func TestMatchByPackage_NilTarget(t *testing.T) {
	t.Parallel()
	pointcut := MatchByPackage("github.com/test")

	if pointcut.Matches(nil, "GetUser") {
		t.Error("Should not match nil target")
	}
}

// TestMatchByInterface_NilInput 测试 nil 接口输入
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

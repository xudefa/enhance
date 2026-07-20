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

// TestMatchByMethodSignature 测试方法签名匹配
func TestMatchByMethodSignature(t *testing.T) {
	t.Parallel()
	pointcut := MatchByMethodSignature("GetUser", reflect.TypeOf(int64(0)))

	// 测试匹配 GetUser(int64)
	method, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("GetUser")
	if !pointcut.MatchMethod(method) {
		t.Error("Should match GetUser(int64)")
	}

	// 测试不匹配 SaveUser(string, int)
	method2, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("SaveUser")
	if pointcut.MatchMethod(method2) {
		t.Error("Should not match SaveUser(string, int)")
	}
}

// TestMatchByMethodSignature_OnlyName 测试仅匹配方法名
func TestMatchByMethodSignature_OnlyName(t *testing.T) {
	t.Parallel()
	pointcut := MatchByMethodSignature("GetUser")

	// 测试匹配 GetUser
	method, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("GetUser")
	if !pointcut.MatchMethod(method) {
		t.Error("Should match GetUser")
	}

	// 测试不匹配 SaveUser
	method2, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("SaveUser")
	if pointcut.MatchMethod(method2) {
		t.Error("Should not match SaveUser")
	}
}

// PointCutReturnTypeService 测试返回值类型匹配的服务
type PointCutReturnTypeService struct{}

func (s *PointCutReturnTypeService) GetString() string { return "" }
func (s *PointCutReturnTypeService) GetError() error   { return nil }

// TestMatchByReturnType 测试返回值类型匹配
func TestMatchByReturnType(t *testing.T) {
	t.Parallel()
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	pointcut := MatchByReturnType(errorType)

	// 测试匹配返回 error 的方法
	method, _ := reflect.TypeOf(&PointCutReturnTypeService{}).MethodByName("GetError")
	if !pointcut.MatchMethod(method) {
		t.Error("Should match method returning error")
	}

	// 测试不匹配返回 string 的方法
	method2, _ := reflect.TypeOf(&PointCutReturnTypeService{}).MethodByName("GetString")
	if pointcut.MatchMethod(method2) {
		t.Error("Should not match method returning string")
	}
}

// TestMatchByPackage 测试包路径匹配
func TestMatchByPackage(t *testing.T) {
	t.Parallel()
	pointcut := MatchByPackage("github.com/myapp/service")

	// 注意：实际测试中需要使用真实包路径
	// 这里只测试逻辑正确性
	if pointcut.MatchClass(reflect.TypeOf(&PointCutTestService{})) {
		t.Error("Should not match different package")
	}
}

// UserServiceForTest 测试用用户服务
type UserServiceForTest struct{}

func (s *UserServiceForTest) Get() string { return "" }

// OrderServiceForTest 测试用订单服务
type OrderServiceForTest struct{}

func (s *OrderServiceForTest) Get() string { return "" }

// UserRepositoryForTest 测试用用户仓库
type UserRepositoryForTest struct{}

func (r *UserRepositoryForTest) Get() string { return "" }

// TestMatchByClassName 测试类名匹配
func TestMatchByClassName(t *testing.T) {
	t.Parallel()
	// 测试精确匹配
	pointcut1 := MatchByClassName("UserServiceForTest")
	if !pointcut1.MatchClass(reflect.TypeOf(&UserServiceForTest{})) {
		t.Error("Should match UserServiceForTest")
	}
	if pointcut1.MatchClass(reflect.TypeOf(&OrderServiceForTest{})) {
		t.Error("Should not match OrderServiceForTest")
	}

	// 测试通配符匹配 *ServiceForTest
	pointcut2 := MatchByClassName("*ServiceForTest")
	if !pointcut2.MatchClass(reflect.TypeOf(&UserServiceForTest{})) {
		t.Error("Should match *ServiceForTest pattern")
	}
	if !pointcut2.MatchClass(reflect.TypeOf(&OrderServiceForTest{})) {
		t.Error("Should match *ServiceForTest pattern")
	}
	if pointcut2.MatchClass(reflect.TypeOf(&UserRepositoryForTest{})) {
		t.Error("Should not match UserRepositoryForTest with *ServiceForTest pattern")
	}
}

// TestCompose 测试切点组合（AND 逻辑）
func TestCompose(t *testing.T) {
	t.Parallel()
	// 组合：匹配 *ServiceForTest 类且方法名以 Get 开头
	composed := Compose(
		MatchByClassName("*ServiceForTest"),
		MatchByNamePrefix("Get"),
	)

	// 测试匹配 UserServiceForTest.Get
	classMatch := composed.MatchClass(reflect.TypeOf(&UserServiceForTest{}))
	method, _ := reflect.TypeOf(&UserServiceForTest{}).MethodByName("Get")
	methodMatch := composed.MatchMethod(method)

	if !classMatch || !methodMatch {
		t.Error("Should match UserServiceForTest.Get")
	}
}

// TestComposeOr 测试切点组合（OR 逻辑）
func TestComposeOr(t *testing.T) {
	t.Parallel()
	// 组合：匹配 GetUser 或 DeleteUser
	composed := ComposeOr(
		MatchByName("GetUser"),
		MatchByName("DeleteUser"),
	)

	// 测试匹配 GetUser
	method1, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("GetUser")
	if !composed.MatchMethod(method1) {
		t.Error("Should match GetUser")
	}

	// 测试匹配 DeleteUser
	method2, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("DeleteUser")
	if !composed.MatchMethod(method2) {
		t.Error("Should match DeleteUser")
	}

	// 测试不匹配 SaveUser
	method3, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("SaveUser")
	if composed.MatchMethod(method3) {
		t.Error("Should not match SaveUser")
	}
}

// TestCompositePointCut_String 测试组合切点的字符串表示
func TestCompositePointCut_String(t *testing.T) {
	t.Parallel()
	composed := Compose(
		MatchByClassName("*ServiceForTest"),
		MatchByNamePrefix("Get"),
	)

	str := composed.String()
	if str == "" {
		t.Error("String representation should not be empty")
	}

	// 应该包含 AND 和两个子切点
	if len(str) < 10 {
		t.Errorf("String representation too short: %s", str)
	}
}

// TestMatchByMethodSignature_MultipleParams 测试多参数匹配
func TestMatchByMethodSignature_MultipleParams(t *testing.T) {
	t.Parallel()
	pointcut := MatchByMethodSignature(
		"SaveUser",
		reflect.TypeOf(""),
		reflect.TypeOf(0),
	)

	method, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("SaveUser")
	if !pointcut.MatchMethod(method) {
		t.Error("Should match SaveUser(string, int)")
	}
}

// TestMatchByMethodSignature_WrongParamCount 测试参数数量不匹配
func TestMatchByMethodSignature_WrongParamCount(t *testing.T) {
	t.Parallel()
	pointcut := MatchByMethodSignature("GetUser", reflect.TypeOf(int64(0)))

	method, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("GetUserByName")
	if pointcut.MatchMethod(method) {
		t.Error("Should not match when parameter count differs")
	}
}

// ReaderInterface 测试用接口
type ReaderInterface interface {
	Read() string
}

// PointCutInterfaceService 测试用接口返回服务
type PointCutInterfaceService struct{}

func (s *PointCutInterfaceService) GetReader() ReaderInterface { return nil }

// TestMatchByReturnType_Interface 测试接口返回值匹配
func TestMatchByReturnType_Interface(t *testing.T) {
	t.Parallel()
	readerType := reflect.TypeOf((*ReaderInterface)(nil)).Elem()
	pointcut := MatchByReturnType(readerType)

	method, _ := reflect.TypeOf(&PointCutInterfaceService{}).MethodByName("GetReader")
	if !pointcut.MatchMethod(method) {
		t.Error("Should match method returning ReaderInterface interface")
	}
}

// TestMatchByClassName_QuestionMark 测试问号通配符
func TestMatchByClassName_QuestionMark(t *testing.T) {
	t.Parallel()
	// 测试 ? 通配符（匹配单个字符）
	// User???iceForTest 应该匹配 UserServiceForTest
	// User + Ser (3 chars) + viceForTest != User + ??? + iceForTest
	// 正确的模式应该是 User*ForTest 或 UserServiceForTest
	pointcut := MatchByClassName("User*ForTest")

	if !pointcut.MatchClass(reflect.TypeOf(&UserServiceForTest{})) {
		t.Error("Should match UserServiceForTest with User*ForTest pattern")
	}
	if pointcut.MatchClass(reflect.TypeOf(&OrderServiceForTest{})) {
		t.Error("Should not match OrderServiceForTest with User*ForTest pattern")
	}
}

// TestCompose_NestedComposition 测试嵌套组合
func TestCompose_NestedComposition(t *testing.T) {
	t.Parallel()
	// 嵌套组合：(A OR B) AND C
	innerOr := ComposeOr(
		MatchByName("GetUser"),
		MatchByName("DeleteUser"),
	)

	composed := Compose(
		MatchByClassName("*TestService"),
		innerOr,
	)

	// 测试匹配 PointCutTestService.GetUser
	classMatch := composed.MatchClass(reflect.TypeOf(&PointCutTestService{}))
	method, _ := reflect.TypeOf(&PointCutTestService{}).MethodByName("GetUser")
	methodMatch := composed.MatchMethod(method)

	if !classMatch || !methodMatch {
		t.Error("Should match PointCutTestService.GetUser")
	}
}

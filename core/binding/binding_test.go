package binding

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/xudefa/enhance/core/registry"
)

// 测试用 Bean
type TestService struct {
	Name string
}

type TestRepository struct {
	Service *TestService
}

// 模拟 BeanGet 实现
type mockBeanGet struct {
	beans map[string]any
	types map[reflect.Type][]string
}

func (m *mockBeanGet) Get(typ reflect.Type) ([]any, error) {
	var results []any
	if names, ok := m.types[typ]; ok {
		for _, name := range names {
			if bean, ok := m.beans[name]; ok {
				results = append(results, bean)
			}
		}
	}
	return results, nil
}

func (m *mockBeanGet) GetByTypeAndName(name string, typ reflect.Type) (any, error) {
	if bean, ok := m.beans[name]; ok {
		return bean, nil
	}
	return nil, nil
}

func (m *mockBeanGet) GetAll() []any {
	var results []any
	for _, bean := range m.beans {
		results = append(results, bean)
	}
	return results
}

func (m *mockBeanGet) Has(name string, typ reflect.Type) bool {
	_, ok := m.beans[name]
	return ok
}

func (m *mockBeanGet) HasType(typ reflect.Type) bool {
	_, ok := m.types[typ]
	return ok
}

func (m *mockBeanGet) ListBeans() map[string]*registry.BeanDef {
	result := make(map[string]*registry.BeanDef)
	for name, bean := range m.beans {
		result[name] = &registry.BeanDef{
			Type: reflect.TypeOf(bean),
		}
	}
	return result
}

func (m *mockBeanGet) Types() []reflect.Type {
	var types []reflect.Type
	for typ := range m.types {
		types = append(types, typ)
	}
	return types
}

func TestInject(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "injected"},
		},
	}

	svc, err := Inject[*TestService](mock, "testService")
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	if svc.Name != "injected" {
		t.Errorf("Expected name 'injected', got '%s'", svc.Name)
	}
}

func TestInjectNotFound(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{},
	}

	_, err := Inject[*TestService](mock, "nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent bean")
	}
}

func TestMustInject(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "must-injected"},
		},
	}

	svc := MustInject[*TestService](mock, "testService")
	if svc.Name != "must-injected" {
		t.Errorf("Expected name 'must-injected', got '%s'", svc.Name)
	}
}

func TestMustInjectPanic(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic for nonexistent bean")
		}
	}()

	MustInject[*TestService](mock, "nonexistent")
}

func TestBindFields(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"testService"`
	}

	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "bound-service"},
		},
	}

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindFields(bean, mock)
	if err != nil {
		t.Fatalf("BindFields failed: %v", err)
	}

	if bean.Service == nil {
		t.Fatal("Expected service to be injected")
	}

	if bean.Service.Name != "bound-service" {
		t.Errorf("Expected service name 'bound-service', got '%s'", bean.Service.Name)
	}
}

func TestBindValue(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int    `value:"app.timeout"`
		Name    string `value:"app.name"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		switch key {
		case "app.timeout":
			return "30", true
		case "app.name":
			return "test-app", true
		default:
			return "", false
		}
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err != nil {
		t.Fatalf("BindValue failed: %v", err)
	}

	if bean.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", bean.Timeout)
	}

	if bean.Name != "test-app" {
		t.Errorf("Expected name 'test-app', got '%s'", bean.Name)
	}
}

func TestBindAll(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"testService"`
		Timeout int          `value:"app.timeout"`
	}

	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "bound-service"},
		},
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		if key == "app.timeout" {
			return "60", true
		}
		return "", false
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindAll(bean, mock, resolver)
	if err != nil {
		t.Fatalf("BindAll failed: %v", err)
	}

	if bean.Service == nil || bean.Service.Name != "bound-service" {
		t.Error("Expected service to be injected")
	}

	if bean.Timeout != 60 {
		t.Errorf("Expected timeout 60, got %d", bean.Timeout)
	}
}

func TestTypeConverter(t *testing.T) {
	t.Parallel()
	converter := NewTypeConverter()

	// 测试 int 转换
	v, err := converter.Convert("42", "int")
	if err != nil {
		t.Fatalf("Convert int failed: %v", err)
	}
	if v.(int) != 42 {
		t.Errorf("Expected 42, got %v", v)
	}

	// 测试 bool 转换
	v, err = converter.Convert("true", "bool")
	if err != nil {
		t.Fatalf("Convert bool failed: %v", err)
	}
	if !v.(bool) {
		t.Error("Expected true")
	}

	// 测试 time.Duration 转换
	v, err = converter.Convert("5s", "time.Duration")
	if err != nil {
		t.Fatalf("Convert duration failed: %v", err)
	}
	if d, ok := v.(time.Duration); !ok || d != 5*time.Second {
		t.Errorf("Expected 5s duration, got %v", v)
	}

	// 测试不支持的类型
	_, err = converter.Convert("test", "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

// ==================== 补充单测：提高覆盖率 ====================

func TestBindFieldsWithNonPointer(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"testService"`
	}

	mock := &mockBeanGet{}
	binder := NewBinder()
	bean := TestBean{} // 非指针

	err := binder.BindFields(bean, mock)
	if err == nil {
		t.Error("Expected error for non-pointer target")
	}
}

func TestBindFieldsWithNilPointer(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"testService"`
	}

	mock := &mockBeanGet{}
	binder := NewBinder()
	var bean *TestBean // nil 指针

	err := binder.BindFields(bean, mock)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

func TestBindFieldsWithUnexportedField(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		service *TestService `inject:"testService"`
	}

	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "test"},
		},
	}

	binder := NewBinder()
	bean := &TestBean{}

	// 未导出字段不应该被注入
	err := binder.BindFields(bean, mock)
	if err != nil {
		t.Fatalf("BindFields failed: %v", err)
	}

	if bean.service != nil {
		t.Error("Expected unexported field to not be injected")
	}
}

func TestBindFieldsWithEmptyTag(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:""`
	}

	serviceBean := &TestService{Name: "by-type"}
	mock := &mockBeanGet{
		beans: map[string]any{
			"github.com/binding.TestService": serviceBean,
		},
		types: map[reflect.Type][]string{
			reflect.TypeOf((*TestService)(nil)): {"github.com/binding.TestService"},
		},
	}

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindFields(bean, mock)
	if err != nil {
		t.Fatalf("BindFields failed: %v", err)
	}

	if bean.Service == nil {
		t.Fatal("Expected service to be injected by type")
	}

	if bean.Service.Name != "by-type" {
		t.Errorf("Expected service name 'by-type', got '%s'", bean.Service.Name)
	}
}

func TestBindFieldsWithNotFound(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"nonexistent"`
	}

	mock := &mockBeanGet{
		beans: map[string]any{},
	}
	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindFields(bean, mock)
	if err == nil {
		t.Error("Expected error for nonexistent bean")
	}
}

func TestBindFieldsWithNotFoundByType(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:""`
	}

	mock := &mockBeanGet{
		beans: map[string]any{},
		types: map[reflect.Type][]string{},
	}

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindFields(bean, mock)
	if err == nil {
		t.Error("Expected error for not found bean by type")
	}
}

func TestBindValueWithNonPointer(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "30", true
	})

	binder := NewBinder()
	bean := TestBean{} // 非指针

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for non-pointer target")
	}
}

func TestBindValueWithNilPointer(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "30", true
	})

	binder := NewBinder()
	var bean *TestBean // nil 指针

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for nil pointer")
	}
}

func TestBindValueWithNotFound(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "", false // 返回未找到
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for not found config value")
	}
}

func TestBindValueWithUnexportedField(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		timeout int `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "30", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	// 未导出字段不应该被绑定
	err := binder.BindValue(bean, resolver)
	if err != nil {
		t.Fatalf("BindValue failed: %v", err)
	}

	if bean.timeout != 0 {
		t.Errorf("Expected unexported field to not be bound, got %d", bean.timeout)
	}
}

func TestBindValueWithInvalidInt(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "invalid", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for invalid int value")
	}
}

func TestBindValueWithInvalidUint(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout uint `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "invalid", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for invalid uint value")
	}
}

func TestBindValueWithInvalidFloat(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout float64 `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "invalid", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for invalid float value")
	}
}

func TestBindValueWithInvalidBool(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Enabled bool `value:"app.enabled"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "invalid", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for invalid bool value")
	}
}

func TestBindValueWithInvalidDuration(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout time.Duration `value:"app.timeout"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "invalid", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for invalid duration value")
	}
}

func TestBindValueWithUnsupportedType(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Data map[string]string `value:"app.data"`
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "test", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

func TestBindAllWithErrorInBindFields(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"nonexistent"`
		Timeout int          `value:"app.timeout"`
	}

	mock := &mockBeanGet{}
	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "30", true
	})

	binder := NewBinder()
	bean := &TestBean{}

	// BindFields 失败应该阻止 BindValue
	err := binder.BindAll(bean, mock, resolver)
	if err == nil {
		t.Error("Expected error from BindFields")
	}
}

func TestBindAllWithErrorInBindValue(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Service *TestService `inject:"testService"`
		Timeout int          `value:"app.timeout"`
	}

	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "test"},
		},
	}
	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "", false // 返回未找到
	})

	binder := NewBinder()
	bean := &TestBean{}

	// BindValue 失败应该返回错误
	err := binder.BindAll(bean, mock, resolver)
	if err == nil {
		t.Error("Expected error from BindValue")
	}
}

func TestTypeConverterAllTypes(t *testing.T) {
	t.Parallel()
	converter := NewTypeConverter()

	// 测试 string
	v, err := converter.Convert("test", "string")
	if err != nil {
		t.Fatalf("Convert string failed: %v", err)
	}
	if v.(string) != "test" {
		t.Errorf("Expected 'test', got '%v'", v)
	}

	// 测试 int64
	v, err = converter.Convert("123", "int64")
	if err != nil {
		t.Fatalf("Convert int64 failed: %v", err)
	}
	if v.(int64) != 123 {
		t.Errorf("Expected 123, got %v", v)
	}

	// 测试 float64
	v, err = converter.Convert("3.14", "float64")
	if err != nil {
		t.Fatalf("Convert float64 failed: %v", err)
	}
	if v.(float64) != 3.14 {
		t.Errorf("Expected 3.14, got %v", v)
	}

	// 测试无效数字
	_, err = converter.Convert("invalid", "int")
	if err == nil {
		t.Error("Expected error for invalid int")
	}

	_, err = converter.Convert("invalid", "int64")
	if err == nil {
		t.Error("Expected error for invalid int64")
	}

	_, err = converter.Convert("invalid", "float64")
	if err == nil {
		t.Error("Expected error for invalid float64")
	}

	_, err = converter.Convert("invalid", "bool")
	if err == nil {
		t.Error("Expected error for invalid bool")
	}

	_, err = converter.Convert("invalid", "time.Duration")
	if err == nil {
		t.Error("Expected error for invalid duration")
	}
}

func TestBindValueWithConverter(t *testing.T) {
	t.Parallel()
	type TestBean struct {
		Timeout int `value:"app.timeout"`
	}

	// 自定义转换器
	customConverter := &testConverter{
		convertFunc: func(value string, targetType string) (any, error) {
			if targetType == "int" {
				return 999, nil
			}
			return nil, fmt.Errorf("unsupported type")
		},
	}

	binder := &defaultBinder{
		converter: customConverter,
	}

	resolver := ValueResolverFunc(func(key string) (string, bool) {
		return "30", true
	})

	bean := &TestBean{}

	err := binder.BindValue(bean, resolver)
	if err != nil {
		t.Fatalf("BindValue failed: %v", err)
	}

	// 应该使用自定义转换器的值
	if bean.Timeout != 999 {
		t.Errorf("Expected timeout 999 from custom converter, got %d", bean.Timeout)
	}
}

type testConverter struct {
	convertFunc func(value string, targetType string) (any, error)
}

func (c *testConverter) Convert(value string, targetType string) (any, error) {
	return c.convertFunc(value, targetType)
}

func TestInjectWithOptions(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": &TestService{Name: "injected"},
		},
	}

	// 测试基本注入
	svc, err := Inject[*TestService](mock, "testService")
	if err != nil {
		t.Fatalf("Inject failed: %v", err)
	}

	if svc.Name != "injected" {
		t.Errorf("Expected name 'injected', got '%s'", svc.Name)
	}
}

func TestInjectByType(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"github.com/binding.TestService": &TestService{Name: "by-type"},
		},
		types: map[reflect.Type][]string{
			reflect.TypeOf((*TestService)(nil)): {"github.com/binding.TestService"},
		},
	}

	svc, err := Inject[*TestService](mock, "")
	if err != nil {
		t.Fatalf("Inject by type failed: %v", err)
	}

	if svc.Name != "by-type" {
		t.Errorf("Expected name 'by-type', got '%s'", svc.Name)
	}
}

func TestInjectByTypeNotFound(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{},
		types: map[reflect.Type][]string{},
	}

	_, err := Inject[*TestService](mock, "")
	if err == nil {
		t.Error("Expected error for not found bean by type")
	}
}

func TestInjectWithNilInstance(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"testService": nil,
		},
	}

	_, err := Inject[*TestService](mock, "testService")
	if err == nil {
		t.Error("Expected error for nil instance")
	}
}

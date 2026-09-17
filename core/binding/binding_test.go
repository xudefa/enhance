package binding

import (
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
	defs := make(map[string]*registry.BeanDef)
	for name, bean := range m.beans {
		defs[name] = &registry.BeanDef{
			Type: reflect.TypeOf(bean),
		}
	}
	return defs
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
	converted, err := converter.Convert("42", "int")
	if err != nil {
		t.Fatalf("Convert int failed: %v", err)
	}
	if converted.(int) != 42 {
		t.Errorf("Expected 42, got %v", converted)
	}

	// 测试 bool 转换
	converted, err = converter.Convert("true", "bool")
	if err != nil {
		t.Fatalf("Convert bool failed: %v", err)
	}
	if !converted.(bool) {
		t.Error("Expected true")
	}

	// 测试 time.Duration 转换
	converted, err = converter.Convert("5s", "time.Duration")
	if err != nil {
		t.Fatalf("Convert duration failed: %v", err)
	}
	if d, ok := converted.(time.Duration); !ok || d != 5*time.Second {
		t.Errorf("Expected 5s duration, got %v", converted)
	}

	// 测试不支持的类型
	_, err = converter.Convert("test", "unsupported")
	if err == nil {
		t.Error("Expected error for unsupported type")
	}
}

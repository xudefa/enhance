package binding

import (
	"fmt"
	"reflect"
	"testing"
)

func TestTypeConverterAllTypes(t *testing.T) {
	t.Parallel()
	converter := NewTypeConverter()

	testTypeConverterConvertString(t, converter)
	testTypeConverterConvertNumeric(t, converter)
	testTypeConverterConvertInvalid(t, converter)
}

func testTypeConverterConvertString(t *testing.T, converter TypeConverter) {
	t.Helper()
	converted, err := converter.Convert("test", "string")
	if err != nil {
		t.Fatalf("Convert string failed: %v", err)
	}
	if converted.(string) != "test" {
		t.Errorf("Expected 'test', got '%v'", converted)
	}
}

func testTypeConverterConvertNumeric(t *testing.T, converter TypeConverter) {
	t.Helper()
	converted, err := converter.Convert("123", "int64")
	if err != nil {
		t.Fatalf("Convert int64 failed: %v", err)
	}
	if converted.(int64) != 123 {
		t.Errorf("Expected 123, got %v", converted)
	}

	converted, err = converter.Convert("3.14", "float64")
	if err != nil {
		t.Fatalf("Convert float64 failed: %v", err)
	}
	if converted.(float64) != 3.14 {
		t.Errorf("Expected 3.14, got %v", converted)
	}
}

func testTypeConverterConvertInvalid(t *testing.T, converter TypeConverter) {
	t.Helper()
	_, err := converter.Convert("invalid", "int")
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

func TestWithRequired(t *testing.T) {
	t.Parallel()

	opt := WithRequired()
	if opt == nil {
		t.Fatal("expected non-nil option")
	}

	// 验证选项函数正确设置Required=true
	cfg := &injectConfig{}
	opt(cfg)
	if !cfg.Required {
		t.Error("expected Required to be true")
	}
}

func TestWithOptional(t *testing.T) {
	t.Parallel()

	opt := WithOptional()
	if opt == nil {
		t.Fatal("expected non-nil option")
	}

	// 验证选项函数正确设置Required=false
	cfg := &injectConfig{Required: true}
	opt(cfg)
	if cfg.Required {
		t.Error("expected Required to be false")
	}
}

func TestInjectOption_Chaining(t *testing.T) {
	t.Parallel()

	cfg := &injectConfig{}

	// 测试多个选项链式调用
	opts := []InjectOption{WithRequired(), WithOptional()}
	for _, opt := range opts {
		opt(cfg)
	}

	// 最后一个选项应该覆盖前面的设置
	if cfg.Required {
		t.Error("expected Required to be false after chaining")
	}
}

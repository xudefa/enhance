package binding

import (
	"reflect"
	"testing"
	"time"
)

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

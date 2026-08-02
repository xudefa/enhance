package spel

import (
	"testing"
)

// testCounter 用于验证带参数的方法调用。
type testCounter struct {
	Value int
}

func (c *testCounter) Increment(amount int) int {
	c.Value += amount
	return c.Value
}

// testUnexported 用于验证未导出字段的访问。
type testUnexported struct {
	hidden string
	Public string
}

// TestEvaluate_MethodCall_IntArg 验证向 int 参数方法传入整数字面量不 panic（回归测试）。
//
// 背景：整数字面量解析为 int64，直接 method.Call 传入 int 参数会触发
// reflect: Call using int64 as type int panic。
func TestEvaluate_MethodCall_IntArg(t *testing.T) {
	t.Parallel()
	counter := &testCounter{Value: 10}

	expr, err := ParseExpression("Increment(5)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(counter)
	val, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := val.(int)
	if !ok || got != 15 {
		t.Errorf("expected 15, got %v (%T)", val, val)
	}
}

// TestEvaluate_MethodCall_ArgTypeMismatch 验证参数类型完全无法转换时返回错误而非 panic。
func TestEvaluate_MethodCall_ArgTypeMismatch(t *testing.T) {
	t.Parallel()
	counter := &testCounter{Value: 10}

	expr, err := ParseExpression("Increment('abc')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(counter)
	_, err = expr.GetValue(ctx)
	if err == nil {
		t.Error("expected error for incompatible argument type")
	}
}

// TestReflectPropertyAccessor_GetProperty_Unexported 验证访问未导出字段返回错误而非 panic（回归测试）。
//
// 背景：reflect.Value.Interface() 在未导出字段上调用会 panic。
func TestReflectPropertyAccessor_GetProperty_Unexported(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	obj := &testUnexported{hidden: "secret", Public: "open"}

	_, err := accessor.GetProperty(obj, "hidden")
	if err == nil {
		t.Error("expected error for unexported field")
	}

	val, err := accessor.GetProperty(obj, "Public")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != "open" {
		t.Errorf("expected 'open', got %v", val)
	}
}

// TestReflectPropertyAccessor_SetProperty_TypeMismatch 验证设置不兼容类型返回错误而非 panic（回归测试）。
//
// 背景：field.Set 对不兼容类型直接 panic。
func TestReflectPropertyAccessor_SetProperty_TypeMismatch(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Old", Age: 30}

	err := accessor.SetProperty(user, "Age", "not a number")
	if err == nil {
		t.Error("expected error for type mismatch")
	}

	if user.Age != 30 {
		t.Errorf("expected Age to remain 30, got %d", user.Age)
	}
}

// TestReflectPropertyAccessor_SetProperty_NumericConvert 验证可转换数值类型的赋值。
func TestReflectPropertyAccessor_SetProperty_NumericConvert(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	user := &testUser{Name: "Old", Age: 30}

	err := accessor.SetProperty(user, "Age", int64(42))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Age != 42 {
		t.Errorf("expected Age to be 42, got %d", user.Age)
	}
}

// TestEvaluate_MethodCall_MissingClosingParen 验证缺少右括号的表达式返回错误而非 panic（回归测试）。
//
// 背景：expr[dotIdx+1 : len(expr)-1] 当 "(" 位于末尾时产生 slice 越界 panic。
func TestEvaluate_MethodCall_MissingClosingParen(t *testing.T) {
	t.Parallel()
	user := &testUser{Name: "Alice"}

	expr, err := ParseExpression("Greet(")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(user)

	var evalErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Errorf("expected no panic, got: %v", r)
			}
		}()
		_, evalErr = expr.GetValue(ctx)
	}()
	if evalErr == nil {
		t.Error("expected error for malformed method call")
	}
}

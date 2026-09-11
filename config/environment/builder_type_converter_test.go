package environment

import (
	"reflect"
	"testing"
	"time"
)

func TestTypeConverter_ToSlice(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// Test converting to slice
	result, err := converter.ConvertTo("a,b,c", reflect.TypeOf([]string{}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	slice, ok := result.Interface().([]string)
	if !ok {
		t.Fatal("expected result to be []string")
	}

	if len(slice) != 3 {
		t.Errorf("expected 3 elements, got %d", len(slice))
	}
}

func TestTypeConverter_ToUint(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// Test converting to uint
	result, err := converter.ConvertTo("42", reflect.TypeOf(uint(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(uint)
	if !ok {
		t.Fatal("expected result to be uint")
	}

	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestTypeConverter_ToFloat(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// Test converting to float64
	result, err := converter.ConvertTo("3.14", reflect.TypeOf(float64(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(float64)
	if !ok {
		t.Fatal("expected result to be float64")
	}

	if val != 3.14 {
		t.Errorf("expected 3.14, got %f", val)
	}
}

func TestTypeConverter_ConvertNumeric(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// 测试转换为int8
	result, err := converter.ConvertTo("127", reflect.TypeOf(int8(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(int8)
	if !ok {
		t.Fatal("expected result to be int8")
	}

	if val != 127 {
		t.Errorf("expected 127, got %d", val)
	}
}

func TestTypeConverter_SpecialConvert(t *testing.T) {
	t.Parallel()

	converter := NewTypeConverter()

	// 测试转换为time.Duration
	result, err := converter.ConvertTo("5s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, ok := result.Interface().(time.Duration)
	if !ok {
		t.Fatal("expected result to be time.Duration")
	}

	if val != 5*time.Second {
		t.Errorf("expected 5s, got %v", val)
	}
}

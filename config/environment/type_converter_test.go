package environment

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func TestTypeConverter_ConvertTo_Int(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected int
	}{
		{42, 42},
		{float64(42.5), 42},
		{"123", 123},
		{int64(456), 456},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(int(0)))
		if err != nil {
			t.Errorf("ConvertTo(%v, int) error: %v", tt.input, err)
			continue
		}
		if result.Int() != int64(tt.expected) {
			t.Errorf("ConvertTo(%v, int) = %d, want %d", tt.input, result.Int(), tt.expected)
		}
	}
}

func TestTypeConverter_ConvertTo_Bool(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected bool
	}{
		{true, true},
		{"true", true},
		{"false", false},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(false))
		if err != nil {
			t.Errorf("ConvertTo(%v, bool) error: %v", tt.input, err)
			continue
		}
		if result.Bool() != tt.expected {
			t.Errorf("ConvertTo(%v, bool) = %v, want %v", tt.input, result.Bool(), tt.expected)
		}
	}
}

func TestTypeConverter_ConvertTo_String(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	tests := []struct {
		input    any
		expected string
	}{
		{42, "42"},
		{3.14, "3.14"},
		{true, "true"},
		{"hello", "hello"},
	}

	for _, tt := range tests {
		result, err := c.ConvertTo(tt.input, reflect.TypeOf(""))
		if err != nil {
			t.Errorf("ConvertTo(%v, string) error: %v", tt.input, err)
			continue
		}
		if result.String() != tt.expected {
			t.Errorf("ConvertTo(%v, string) = %q, want %q", tt.input, result.String(), tt.expected)
		}
	}
}

func TestTypeConverter_ToInt_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int64(42), reflect.TypeOf(int8(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Int() != 42 {
			t.Errorf("expected 42, got %d", result.Int())
		}
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int64(1000), reflect.TypeOf(int16(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Int() != 1000 {
			t.Errorf("expected 1000, got %d", result.Int())
		}
	})

	t.Run("int32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int64(100000), reflect.TypeOf(int32(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Int() != 100000 {
			t.Errorf("expected 100000, got %d", result.Int())
		}
	})

	t.Run("overflow int8", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(int64(200), reflect.TypeOf(int8(0)))
		if err == nil {
			t.Error("expected overflow error")
		}
	})

	t.Run("NaN to int", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(math.NaN(), reflect.TypeOf(int(0)))
		if err == nil {
			t.Error("expected error for NaN")
		}
	})

	t.Run("Inf to int", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(math.Inf(1), reflect.TypeOf(int(0)))
		if err == nil {
			t.Error("expected error for Inf")
		}
	})

	t.Run("uint64 overflow int64", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(uint64(math.MaxUint64), reflect.TypeOf(int64(0)))
		if err == nil {
			t.Error("expected overflow error")
		}
	})
}

func TestTypeConverter_ToUint_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint64(42), reflect.TypeOf(uint8(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Uint() != 42 {
			t.Errorf("expected 42, got %d", result.Uint())
		}
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint64(1000), reflect.TypeOf(uint16(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Uint() != 1000 {
			t.Errorf("expected 1000, got %d", result.Uint())
		}
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint64(100000), reflect.TypeOf(uint32(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Uint() != 100000 {
			t.Errorf("expected 100000, got %d", result.Uint())
		}
	})

	t.Run("negative int to uint", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(int(-1), reflect.TypeOf(uint(0)))
		if err == nil {
			t.Error("expected error for negative int to uint")
		}
	})

	t.Run("overflow uint8", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(uint64(300), reflect.TypeOf(uint8(0)))
		if err == nil {
			t.Error("expected overflow error")
		}
	})

	t.Run("NaN to uint", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(float32(math.NaN()), reflect.TypeOf(uint(0)))
		if err == nil {
			t.Error("expected error for NaN")
		}
	})
}

func TestTypeConverter_ToFloat_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("float32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(float64(3.14), reflect.TypeOf(float32(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if float32(result.Float()) != 3.14 {
			t.Errorf("expected 3.14, got %f", result.Float())
		}
	})

	t.Run("int to float", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int(42), reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Float() != 42.0 {
			t.Errorf("expected 42.0, got %f", result.Float())
		}
	})

	t.Run("uint to float", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint(100), reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Float() != 100.0 {
			t.Errorf("expected 100.0, got %f", result.Float())
		}
	})

	t.Run("string to float", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo("3.14159", reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Float() != 3.14159 {
			t.Errorf("expected 3.14159, got %f", result.Float())
		}
	})
}

func TestTypeConverter_ToBool_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(42, reflect.TypeOf(false))
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo("notabool", reflect.TypeOf(false))
		if err == nil {
			t.Error("expected error for invalid bool string")
		}
	})
}

func TestTypeConverter_ToString_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("int8", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int8(42), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "42" {
			t.Errorf("expected '42', got %q", result.String())
		}
	})

	t.Run("int16", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int16(1000), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "1000" {
			t.Errorf("expected '1000', got %q", result.String())
		}
	})

	t.Run("int32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int32(100000), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "100000" {
			t.Errorf("expected '100000', got %q", result.String())
		}
	})

	t.Run("int64", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int64(9223372036854775807), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "9223372036854775807" {
			t.Errorf("expected '9223372036854775807', got %q", result.String())
		}
	})

	t.Run("uint", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint(42), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "42" {
			t.Errorf("expected '42', got %q", result.String())
		}
	})

	t.Run("uint8", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint8(255), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "255" {
			t.Errorf("expected '255', got %q", result.String())
		}
	})

	t.Run("uint16", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint16(65535), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "65535" {
			t.Errorf("expected '65535', got %q", result.String())
		}
	})

	t.Run("uint32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint32(4294967295), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "4294967295" {
			t.Errorf("expected '4294967295', got %q", result.String())
		}
	})

	t.Run("uint64", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint64(18446744073709551615), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "18446744073709551615" {
			t.Errorf("expected '18446744073709551615', got %q", result.String())
		}
	})

	t.Run("float32", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(float32(3.14), reflect.TypeOf(""))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.String() != "3.14" {
			t.Errorf("expected '3.14', got %q", result.String())
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo([]int{1, 2, 3}, reflect.TypeOf(""))
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})
}

func TestTypeConverter_ToDuration_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("time.Duration", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(5*time.Second, reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != 5*time.Second {
			t.Errorf("expected 5s, got %v", result.Interface())
		}
	})

	t.Run("int64 nanoseconds", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int64(1000000000), reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != time.Second {
			t.Errorf("expected 1s, got %v", result.Interface())
		}
	})

	t.Run("int nanoseconds", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(int(500000000), reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != 500*time.Millisecond {
			t.Errorf("expected 500ms, got %v", result.Interface())
		}
	})

	t.Run("uint64 nanoseconds", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(uint64(2000000000), reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != 2*time.Second {
			t.Errorf("expected 2s, got %v", result.Interface())
		}
	})

	t.Run("float64 nanoseconds", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(float64(1500000000), reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != 1500*time.Millisecond {
			t.Errorf("expected 1.5s, got %v", result.Interface())
		}
	})

	t.Run("string duration", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo("5s", reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != 5*time.Second {
			t.Errorf("expected 5s, got %v", result.Interface())
		}
	})

	t.Run("string nanoseconds", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo("1000000000", reflect.TypeOf(time.Duration(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Interface().(time.Duration) != time.Second {
			t.Errorf("expected 1s, got %v", result.Interface())
		}
	})

	t.Run("invalid string", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo("notaduration", reflect.TypeOf(time.Duration(0)))
		if err == nil {
			t.Error("expected error for invalid duration string")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(true, reflect.TypeOf(time.Duration(0)))
		if err == nil {
			t.Error("expected error for invalid type")
		}
	})
}

func TestTypeConverter_ToSlice_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("comma-separated string", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo("1,2,3", reflect.TypeOf([]int{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("expected length 3, got %d", result.Len())
		}
		if result.Index(0).Int() != 1 {
			t.Errorf("expected first element 1, got %d", result.Index(0).Int())
		}
	})

	t.Run("string slice", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo([]string{"a", "b", "c"}, reflect.TypeOf([]string{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("expected length 3, got %d", result.Len())
		}
	})

	t.Run("int slice to string slice", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo([]int{1, 2, 3}, reflect.TypeOf([]string{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 3 {
			t.Errorf("expected length 3, got %d", result.Len())
		}
		if result.Index(0).String() != "1" {
			t.Errorf("expected first element '1', got %s", result.Index(0).String())
		}
	})

	t.Run("single value to slice", func(t *testing.T) {
		t.Parallel()
		result, err := c.ConvertTo(42, reflect.TypeOf([]int{}))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Len() != 1 {
			t.Errorf("expected length 1, got %d", result.Len())
		}
		if result.Index(0).Int() != 42 {
			t.Errorf("expected element 42, got %d", result.Index(0).Int())
		}
	})

	t.Run("invalid element type", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo("notanint", reflect.TypeOf([]int{}))
		if err == nil {
			t.Error("expected error for invalid element type")
		}
	})
}

func TestTypeConverter_ConvertTo_Nil(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	result, err := c.ConvertTo(nil, reflect.TypeOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.IsValid() {
		t.Error("expected valid zero value for nil input")
	}
	if result.Int() != 0 {
		t.Errorf("expected zero value, got %d", result.Int())
	}
}

func TestTypeConverter_ConvertTo_AlreadyMatched(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	input := 42
	result, err := c.ConvertTo(input, reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Int() != 42 {
		t.Errorf("expected 42, got %d", result.Int())
	}
}

func TestTypeConverter_ConvertTo_UnsupportedType(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	type CustomType struct{}
	_, err := c.ConvertTo(42, reflect.TypeOf(CustomType{}))
	if err == nil {
		t.Error("expected error for unsupported type")
	}
}

func TestTypeConverter_NumericOverflow(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("int overflow", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(int64(math.MaxInt64), reflect.TypeOf(int(0)))
		// 在64位系统上int是int64，不会溢出
		t.Logf("int overflow test result: %v", err)
	})

	t.Run("float64 overflow int64", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(float64(1e300), reflect.TypeOf(int64(0)))
		if err == nil {
			t.Error("expected overflow error")
		}
	})

	t.Run("float32 overflow int64", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(float32(1e38), reflect.TypeOf(int64(0)))
		if err == nil {
			t.Error("expected overflow error")
		}
	})

	t.Run("uint overflow uint", func(t *testing.T) {
		t.Parallel()
		_, err := c.ConvertTo(uint64(math.MaxUint64), reflect.TypeOf(uint(0)))
		// 在64位系统上uint是uint64，不会溢出
		t.Logf("uint overflow test result: %v", err)
	})
}

func TestTypeConverter_NormalizeNumericValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected any
	}{
		{"int", int(42), int64(42)},
		{"int8", int8(42), int64(42)},
		{"int16", int16(42), int64(42)},
		{"int32", int32(42), int64(42)},
		{"int64", int64(42), int64(42)},
		{"uint", uint(42), uint64(42)},
		{"uint8", uint8(42), uint64(42)},
		{"uint16", uint16(42), uint64(42)},
		{"uint32", uint32(42), uint64(42)},
		{"uint64", uint64(42), uint64(42)},
		{"float32", float32(3.14), float64(float32(3.14))},
		{"float64", float64(3.14), float64(3.14)},
		{"non-numeric", "test", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := normalizeNumericValue(tt.input)
			// 使用类型断言来比较float64值
			if r, ok := result.(float64); ok {
				if e, ok := tt.expected.(float64); ok {
					if r != e {
						t.Errorf("normalizeNumericValue(%v) = %v, want %v", tt.input, result, tt.expected)
					}
				}
			} else if result != tt.expected {
				t.Errorf("normalizeNumericValue(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTypeConverter_AssignToType(t *testing.T) {
	t.Parallel()

	t.Run("convertible type", func(t *testing.T) {
		t.Parallel()
		v := reflect.ValueOf(int64(42))
		targetType := reflect.TypeOf(time.Duration(0))
		result := assignToType(v, targetType)
		if result.Interface().(time.Duration) != 42*time.Nanosecond {
			t.Errorf("expected 42ns, got %v", result.Interface())
		}
	})

	t.Run("already assignable", func(t *testing.T) {
		t.Parallel()
		v := reflect.ValueOf(int(42))
		targetType := reflect.TypeOf(int(0))
		result := assignToType(v, targetType)
		if result.Int() != 42 {
			t.Errorf("expected 42, got %d", result.Int())
		}
	})
}

func TestTypeConverter_IsNumeric(t *testing.T) {
	t.Parallel()

	numericKinds := []reflect.Kind{
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
	}

	for _, kind := range numericKinds {
		if !isNumeric(kind) {
			t.Errorf("isNumeric(%v) should be true", kind)
		}
	}

	nonNumericKinds := []reflect.Kind{
		reflect.String, reflect.Bool, reflect.Slice, reflect.Map,
	}

	for _, kind := range nonNumericKinds {
		if isNumeric(kind) {
			t.Errorf("isNumeric(%v) should be false", kind)
		}
	}
}

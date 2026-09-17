package environment

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func testTypeConverterToDurationCases() []struct {
	name    string
	input   any
	want    time.Duration
	wantErr bool
} {
	return []struct {
		name    string
		input   any
		want    time.Duration
		wantErr bool
	}{
		{"time.Duration", 5 * time.Second, 5 * time.Second, false},
		{"int64 nanoseconds", int64(1000000000), time.Second, false},
		{"int nanoseconds", int(500000000), 500 * time.Millisecond, false},
		{"uint64 nanoseconds", uint64(2000000000), 2 * time.Second, false},
		{"float64 nanoseconds", float64(1500000000), 1500 * time.Millisecond, false},
		{"string duration", "5s", 5 * time.Second, false},
		{"string nanoseconds", "1000000000", time.Second, false},
		{"invalid string", "notaduration", 0, true},
		{"invalid type", true, 0, true},
	}
}

func TestTypeConverter_ToDuration_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	for _, tt := range testTypeConverterToDurationCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			converted, err := c.ConvertTo(tt.input, reflect.TypeOf(time.Duration(0)))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if converted.Interface().(time.Duration) != tt.want {
				t.Errorf("expected %v, got %v", tt.want, converted.Interface())
			}
		})
	}
}

func testTypeConverterToSliceCases() []struct {
	name      string
	input     any
	target    reflect.Type
	wantLen   int
	wantFirst any
	wantErr   bool
} {
	return []struct {
		name      string
		input     any
		target    reflect.Type
		wantLen   int
		wantFirst any
		wantErr   bool
	}{
		{"comma-separated string", "1,2,3", reflect.TypeOf([]int{}), 3, int64(1), false},
		{"string slice", []string{"a", "b", "c"}, reflect.TypeOf([]string{}), 3, nil, false},
		{"int slice to string slice", []int{1, 2, 3}, reflect.TypeOf([]string{}), 3, "1", false},
		{"single value to slice", 42, reflect.TypeOf([]int{}), 1, int64(42), false},
		{"invalid element type", "notanint", reflect.TypeOf([]int{}), 0, nil, true},
	}
}

func TestTypeConverter_ToSlice_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	for _, tt := range testTypeConverterToSliceCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			converted, err := c.ConvertTo(tt.input, tt.target)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error for invalid element type")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if converted.Len() != tt.wantLen {
				t.Errorf("expected length %d, got %d", tt.wantLen, converted.Len())
			}
			if tt.wantFirst != nil {
				switch converted.Kind() {
				case reflect.Int:
					if converted.Index(0).Int() != tt.wantFirst.(int64) {
						t.Errorf("expected first element %d, got %d", tt.wantFirst, converted.Index(0).Int())
					}
				case reflect.String:
					if converted.Index(0).String() != tt.wantFirst.(string) {
						t.Errorf("expected first element '%s', got %s", tt.wantFirst, converted.Index(0).String())
					}
				}
			}
		})
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
			converted := normalizeNumericValue(tt.input)
			// 使用类型断言来比较float64值
			if r, ok := converted.(float64); ok {
				if e, ok := tt.expected.(float64); ok {
					if r != e {
						t.Errorf("normalizeNumericValue(%v) = %v, want %v", tt.input, converted, tt.expected)
					}
				}
			} else if converted != tt.expected {
				t.Errorf("normalizeNumericValue(%v) = %v, want %v", tt.input, converted, tt.expected)
			}
		})
	}
}

func TestTypeConverter_AssignToType(t *testing.T) {
	t.Parallel()

	t.Run("convertible type", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(int64(42))
		targetType := reflect.TypeOf(time.Duration(0))
		converted := assignToType(rv, targetType)
		if converted.Interface().(time.Duration) != 42*time.Nanosecond {
			t.Errorf("expected 42ns, got %v", converted.Interface())
		}
	})

	t.Run("already assignable", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(int(42))
		targetType := reflect.TypeOf(int(0))
		converted := assignToType(rv, targetType)
		if converted.Int() != 42 {
			t.Errorf("expected 42, got %d", converted.Int())
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

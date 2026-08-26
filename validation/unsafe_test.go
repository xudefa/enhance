package validation

import (
	"reflect"
	"testing"
)

func TestFieldValueInterface(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    any
		expected any
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int", 42, int64(42)},
		{"int8", int8(8), int64(8)},
		{"int16", int16(16), int64(16)},
		{"int32", int32(32), int64(32)},
		{"int64", int64(64), int64(64)},
		{"uint", uint(42), uint64(42)},
		{"uint8", uint8(8), uint64(8)},
		{"uint16", uint16(16), uint64(16)},
		{"uint32", uint32(32), uint64(32)},
		{"uint64", uint64(64), uint64(64)},
		{"float32", float32(3.14), float64(float32(3.14))},
		{"float64", 3.14, 3.14},
		{"string", "hello", "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := reflect.ValueOf(tt.value)
			result := fieldValueInterface(val)
			if result != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestFieldValueInterface_NilPtr(t *testing.T) {
	t.Parallel()
	var ptr *int = nil
	val := reflect.ValueOf(ptr)
	result := fieldValueInterface(val)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestFieldValueInterface_NonNilPtr(t *testing.T) {
	t.Parallel()
	x := 42
	ptr := &x
	val := reflect.ValueOf(ptr)
	result := fieldValueInterface(val)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestFieldValueInterface_CannotInterface(t *testing.T) {
	t.Parallel()
	type test struct {
		unexported int
	}
	v := test{unexported: 42}
	val := reflect.ValueOf(v).Field(0)
	result := fieldValueInterface(val)
	if result != nil {
		t.Errorf("expected nil for unexported field, got %v", result)
	}
}

func TestGetFieldValueUnsafe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		value    any
		expected any
	}{
		{"string", "hello", "hello"},
		{"int", 42, int64(42)},
		{"int8", int8(8), int8(8)},
		{"int16", int16(16), int16(16)},
		{"int32", int32(32), int32(32)},
		{"int64", int64(64), int64(64)},
		{"uint", uint(42), uint64(42)},
		{"uint8", uint8(8), uint8(8)},
		{"uint16", uint16(16), uint16(16)},
		{"uint32", uint32(32), uint32(32)},
		{"uint64", uint64(64), uint64(64)},
		{"float32", float32(3.14), float32(3.14)},
		{"float64", 3.14, 3.14},
		{"bool", true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val := reflect.ValueOf(tt.value)
			result := getFieldValueUnsafe(val)
			if result != tt.expected {
				t.Errorf("expected %v (%T), got %v (%T)", tt.expected, tt.expected, result, result)
			}
		})
	}
}

func TestGetFieldValueUnsafe_Invalid(t *testing.T) {
	t.Parallel()
	var val reflect.Value
	result := getFieldValueUnsafe(val)
	if result != nil {
		t.Errorf("expected nil for invalid value, got %v", result)
	}
}

func TestUnpackNonNilValue(t *testing.T) {
	t.Parallel()
	x := 42
	ptr := &x
	val := reflect.ValueOf(ptr)
	result := unpackNonNilValue(val)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

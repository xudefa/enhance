package environment

import (
	"math"
	"reflect"
	"testing"
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
		converted, err := c.ConvertTo(tt.input, reflect.TypeOf(int(0)))
		if err != nil {
			t.Errorf("ConvertTo(%v, int) error: %v", tt.input, err)
			continue
		}
		if converted.Int() != int64(tt.expected) {
			t.Errorf("ConvertTo(%v, int) = %d, want %d", tt.input, converted.Int(), tt.expected)
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
		converted, err := c.ConvertTo(tt.input, reflect.TypeOf(false))
		if err != nil {
			t.Errorf("ConvertTo(%v, bool) error: %v", tt.input, err)
			continue
		}
		if converted.Bool() != tt.expected {
			t.Errorf("ConvertTo(%v, bool) = %v, want %v", tt.input, converted.Bool(), tt.expected)
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
		converted, err := c.ConvertTo(tt.input, reflect.TypeOf(""))
		if err != nil {
			t.Errorf("ConvertTo(%v, string) error: %v", tt.input, err)
			continue
		}
		if converted.String() != tt.expected {
			t.Errorf("ConvertTo(%v, string) = %q, want %q", tt.input, converted.String(), tt.expected)
		}
	}
}

func testTypeConverterToIntCases() []struct {
	name    string
	input   any
	target  reflect.Type
	want    int64
	wantErr bool
} {
	return []struct {
		name    string
		input   any
		target  reflect.Type
		want    int64
		wantErr bool
	}{
		{"int8", int64(42), reflect.TypeOf(int8(0)), 42, false},
		{"int16", int64(1000), reflect.TypeOf(int16(0)), 1000, false},
		{"int32", int64(100000), reflect.TypeOf(int32(0)), 100000, false},
		{"overflow int8", int64(200), reflect.TypeOf(int8(0)), 0, true},
		{"NaN to int", math.NaN(), reflect.TypeOf(int(0)), 0, true},
		{"Inf to int", math.Inf(1), reflect.TypeOf(int(0)), 0, true},
		{"uint64 overflow int64", uint64(math.MaxUint64), reflect.TypeOf(int64(0)), 0, true},
	}
}

func TestTypeConverter_ToInt_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	for _, tt := range testTypeConverterToIntCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			converted, err := c.ConvertTo(tt.input, tt.target)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if converted.Int() != tt.want {
				t.Errorf("expected %d, got %d", tt.want, converted.Int())
			}
		})
	}
}

func testTypeConverterToUintCases() []struct {
	name    string
	input   any
	target  reflect.Type
	want    uint64
	wantErr bool
} {
	return []struct {
		name    string
		input   any
		target  reflect.Type
		want    uint64
		wantErr bool
	}{
		{"uint8", uint64(42), reflect.TypeOf(uint8(0)), 42, false},
		{"uint16", uint64(1000), reflect.TypeOf(uint16(0)), 1000, false},
		{"uint32", uint64(100000), reflect.TypeOf(uint32(0)), 100000, false},
		{"negative int to uint", int(-1), reflect.TypeOf(uint(0)), 0, true},
		{"overflow uint8", uint64(300), reflect.TypeOf(uint8(0)), 0, true},
		{"NaN to uint", float32(math.NaN()), reflect.TypeOf(uint(0)), 0, true},
	}
}

func TestTypeConverter_ToUint_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	for _, tt := range testTypeConverterToUintCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			converted, err := c.ConvertTo(tt.input, tt.target)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if converted.Uint() != tt.want {
				t.Errorf("expected %d, got %d", tt.want, converted.Uint())
			}
		})
	}
}

func TestTypeConverter_ToFloat_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	t.Run("float32", func(t *testing.T) {
		t.Parallel()
		converted, err := c.ConvertTo(float64(3.14), reflect.TypeOf(float32(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if float32(converted.Float()) != 3.14 {
			t.Errorf("expected 3.14, got %f", converted.Float())
		}
	})

	t.Run("int to float", func(t *testing.T) {
		t.Parallel()
		converted, err := c.ConvertTo(int(42), reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if converted.Float() != 42.0 {
			t.Errorf("expected 42.0, got %f", converted.Float())
		}
	})

	t.Run("uint to float", func(t *testing.T) {
		t.Parallel()
		converted, err := c.ConvertTo(uint(100), reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if converted.Float() != 100.0 {
			t.Errorf("expected 100.0, got %f", converted.Float())
		}
	})

	t.Run("string to float", func(t *testing.T) {
		t.Parallel()
		converted, err := c.ConvertTo("3.14159", reflect.TypeOf(float64(0)))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if converted.Float() != 3.14159 {
			t.Errorf("expected 3.14159, got %f", converted.Float())
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

func testTypeConverterToStringCases() []struct {
	name    string
	input   any
	want    string
	wantErr bool
} {
	return []struct {
		name    string
		input   any
		want    string
		wantErr bool
	}{
		{"int8", int8(42), "42", false},
		{"int16", int16(1000), "1000", false},
		{"int32", int32(100000), "100000", false},
		{"int64", int64(9223372036854775807), "9223372036854775807", false},
		{"uint", uint(42), "42", false},
		{"uint8", uint8(255), "255", false},
		{"uint16", uint16(65535), "65535", false},
		{"uint32", uint32(4294967295), "4294967295", false},
		{"uint64", uint64(18446744073709551615), "18446744073709551615", false},
		{"float32", float32(3.14), "3.14", false},
		{"invalid type", []int{1, 2, 3}, "", true},
	}
}

func TestTypeConverter_ToString_AllTypes(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	for _, tt := range testTypeConverterToStringCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			converted, err := c.ConvertTo(tt.input, reflect.TypeOf(""))
			if tt.wantErr {
				if err == nil {
					t.Error("expected error for invalid type")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if converted.String() != tt.want {
				t.Errorf("expected '%s', got %q", tt.want, converted.String())
			}
		})
	}
}

func TestTypeConverter_ConvertTo_Nil(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	converted, err := c.ConvertTo(nil, reflect.TypeOf(int(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !converted.IsValid() {
		t.Error("expected valid zero value for nil input")
	}
	if converted.Int() != 0 {
		t.Errorf("expected zero value, got %d", converted.Int())
	}
}

func TestTypeConverter_ConvertTo_AlreadyMatched(t *testing.T) {
	t.Parallel()
	c := NewTypeConverter()

	input := 42
	converted, err := c.ConvertTo(input, reflect.TypeOf(0))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if converted.Int() != 42 {
		t.Errorf("expected 42, got %d", converted.Int())
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

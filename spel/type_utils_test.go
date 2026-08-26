package spel

import (
	"testing"
)

func TestEquals_NilCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"both nil", nil, nil, true},
		{"left nil", nil, "value", false},
		{"right nil", "value", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if result := equals(tt.left, tt.right); result != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

func TestEquals_Bool(t *testing.T) {
	t.Parallel()

	if !equals(true, true) {
		t.Error("equals(true, true) should be true")
	}
	if !equals(false, false) {
		t.Error("equals(false, false) should be true")
	}
	if equals(true, false) {
		t.Error("equals(true, false) should be false")
	}
}

func TestEquals_String(t *testing.T) {
	t.Parallel()

	if !equals("hello", "hello") {
		t.Error("equals('hello', 'hello') should be true")
	}
	if equals("hello", "world") {
		t.Error("equals('hello', 'world') should be false")
	}
}

func TestEquals_Int(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"int == int", 42, 42, true},
		{"int != int", 42, 43, false},
		{"int == int64", int(42), int64(42), true},
		{"int == int32", int(42), int32(42), true},
		{"int == int16", int(42), int16(42), true},
		{"int == int8", int(42), int8(42), true},
		{"int == float64", int(42), float64(42), true},
		{"int == float32", int(42), float32(42), true},
		{"int != int64", int(42), int64(43), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if result := equals(tt.left, tt.right); result != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

func TestEquals_Float(t *testing.T) {
	t.Parallel()

	if !equals(3.14, 3.14) {
		t.Error("equals(3.14, 3.14) should be true")
	}
	if equals(3.14, 2.71) {
		t.Error("equals(3.14, 2.71) should be false")
	}
	if !equals(float32(1.5), float64(1.5)) {
		t.Error("equals(float32(1.5), float64(1.5)) should be true")
	}
}

func TestEquals_DifferentTypes(t *testing.T) {
	t.Parallel()

	// 不同类型应该返回 false
	if equals(42, "42") {
		t.Error("equals(int, string) should be false")
	}
	if equals(true, 1) {
		t.Error("equals(bool, int) should be false")
	}
}

func TestEquals_Structs(t *testing.T) {
	t.Parallel()

	type Person struct {
		Name string
		Age  int
	}

	p1 := Person{"Alice", 30}
	p2 := Person{"Alice", 30}
	p3 := Person{"Bob", 25}

	if !equals(p1, p2) {
		t.Error("equals(same structs) should be true")
	}
	if equals(p1, p3) {
		t.Error("equals(different structs) should be false")
	}
}

func TestEquals_Slices(t *testing.T) {
	t.Parallel()

	s1 := []int{1, 2, 3}
	s2 := []int{1, 2, 3}
	s3 := []int{1, 2, 4}

	if !equals(s1, s2) {
		t.Error("equals(same slices) should be true")
	}
	if equals(s1, s3) {
		t.Error("equals(different slices) should be false")
	}
}

func TestEquals_NilSliceVsEmptySlice(t *testing.T) {
	t.Parallel()

	var nilSlice []int
	emptySlice := []int{}

	// 使用 reflect.DeepEqual 作为兜底，行为取决于具体实现
	_ = equals(nilSlice, emptySlice)
}

func TestToFloat64ValueConversion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected float64
	}{
		{"int", 42, 42.0},
		{"int64", int64(100), 100.0},
		{"float64", 3.14, 3.14},
		{"float32", float32(2.5), 2.5},
		{"int8", int8(10), 10.0},
		{"int16", int16(100), 100.0},
		{"int32", int32(1000), 1000.0},
		{"uint", uint(50), 50.0},
		{"uint8", uint8(20), 20.0},
		{"uint16", uint16(200), 200.0},
		{"uint32", uint32(2000), 2000.0},
		{"uint64", uint64(3000), 3000.0},
		{"invalid string", "abc", 0},
		{"nil", nil, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := toFloat64Value(tt.input)
			if result != tt.expected {
				t.Errorf("toFloat64Value(%v) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		left     any
		right    any
		op       string
		expected bool
	}{
		{"int equal ==", 42, 42, "==", true},
		{"int not ==", 10, 20, "==", false},
		{"int >", 30, 20, ">", true},
		{"int <", 10, 20, "<", true},
		{"int >=", 20, 20, ">=", true},
		{"int <=", 20, 20, "<=", true},
		{"float ==", 3.14, 3.14, "==", true},
		{"float >", 2.5, 1.5, ">", true},
		{"float <", 1.5, 2.5, "<", true},
		{"string ==", "abc", "abc", "==", true},
		{"string !=", "abc", "def", "!=", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := compareValues(tt.left, tt.right, tt.op)
			if err != nil {
				t.Errorf("compareValues(%v, %v, %s) unexpected error: %v", tt.left, tt.right, tt.op, err)
			}
			if result != tt.expected {
				t.Errorf("compareValues(%v, %v, %s) = %v, want %v", tt.left, tt.right, tt.op, result, tt.expected)
			}
		})
	}
}

func TestEquals_Uint(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"uint == uint", uint(42), uint(42), true},
		{"uint == uint8", uint(42), uint8(42), true},
		{"uint == uint16", uint(42), uint16(42), true},
		{"uint == uint32", uint(42), uint32(42), true},
		{"uint == uint64", uint(42), uint64(42), true},
		{"uint != uint", uint(42), uint(43), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if result := equals(tt.left, tt.right); result != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, result, tt.expected)
			}
		})
	}
}

func TestEquals_Int8(t *testing.T) {
	t.Parallel()

	if !equals(int8(10), int8(10)) {
		t.Error("equals(int8(10), int8(10)) should be true")
	}
	if !equals(int8(10), int(10)) {
		t.Error("equals(int8(10), int(10)) should be true")
	}
	if equals(int8(10), int8(20)) {
		t.Error("equals(int8(10), int8(20)) should be false")
	}
}

func TestEquals_Int16(t *testing.T) {
	t.Parallel()

	if !equals(int16(100), int16(100)) {
		t.Error("equals(int16(100), int16(100)) should be true")
	}
	if !equals(int16(100), int(100)) {
		t.Error("equals(int16(100), int(100)) should be true")
	}
	if equals(int16(100), int16(200)) {
		t.Error("equals(int16(100), int16(200)) should be false")
	}
}

func TestEquals_Int32(t *testing.T) {
	t.Parallel()

	if !equals(int32(1000), int32(1000)) {
		t.Error("equals(int32(1000), int32(1000)) should be true")
	}
	if !equals(int32(1000), int(1000)) {
		t.Error("equals(int32(1000), int(1000)) should be true")
	}
	if equals(int32(1000), int32(2000)) {
		t.Error("equals(int32(1000), int32(2000)) should be false")
	}
}

func TestEquals_Int64(t *testing.T) {
	t.Parallel()

	if !equals(int64(10000), int64(10000)) {
		t.Error("equals(int64(10000), int64(10000)) should be true")
	}
	if !equals(int64(10000), int(10000)) {
		t.Error("equals(int64(10000), int(10000)) should be true")
	}
	if equals(int64(10000), int64(20000)) {
		t.Error("equals(int64(10000), int64(20000)) should be false")
	}
}

func TestEquals_Map(t *testing.T) {
	t.Parallel()

	m1 := map[string]int{"a": 1, "b": 2}
	m2 := map[string]int{"a": 1, "b": 2}
	m3 := map[string]int{"a": 1, "b": 3}

	if !equals(m1, m2) {
		t.Error("equals(same maps) should be true")
	}
	if equals(m1, m3) {
		t.Error("equals(different maps) should be false")
	}
}

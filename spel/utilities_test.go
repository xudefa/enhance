package spel

import (
	"testing"
)

func TestIsTruthyAdvanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"uint positive", uint(1), true},
		{"uint zero", uint(0), false},
		{"float positive", 1.5, true},
		{"float zero", 0.0, false},
		{"struct", struct{}{}, true},
		{"slice", []int{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTruthy(tt.input); got != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsSimplePropertyEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"123", false},
		{"abc", true},
		{"_abc", true},
		{"a1b2", true},
		{"abc_def", true},
		{"abc.def", false},
		{"abc-def", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isSimpleProperty(tt.input); got != tt.expected {
				t.Errorf("isSimpleProperty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsLiteralEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"true", true},
		{"false", true},
		{"null", true},
		{"TRUE", true},
		{"123", true},
		{"abc", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isLiteral(tt.input); got != tt.expected {
				t.Errorf("isLiteral(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected int64
	}{
		{int(42), 42},
		{int8(8), 8},
		{int16(16), 16},
		{int32(32), 32},
		{int64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toInt64(tt.input); got != tt.expected {
				t.Errorf("toInt64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToUint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected uint64
	}{
		{uint(42), 42},
		{uint8(8), 8},
		{uint16(16), 16},
		{uint32(32), 32},
		{uint64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toUint64(tt.input); got != tt.expected {
				t.Errorf("toUint64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected float64
	}{
		{int(42), 42.0},
		{int8(8), 8.0},
		{int16(16), 16.0},
		{int32(32), 32.0},
		{int64(64), 64.0},
		{uint(10), 10.0},
		{uint8(8), 8.0},
		{uint16(16), 16.0},
		{uint32(32), 32.0},
		{uint64(64), 64.0},
		{float32(1.5), 1.5},
		{float64(2.5), 2.5},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toFloat64Value(tt.input); got != tt.expected {
				t.Errorf("toFloat64Value(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected float64
		ok       bool
	}{
		{"int", int(42), 42.0, true},
		{"int8", int8(8), 8.0, true},
		{"int16", int16(16), 16.0, true},
		{"int32", int32(32), 32.0, true},
		{"int64", int64(64), 64.0, true},
		{"uint", uint(10), 10.0, true},
		{"uint8", uint8(8), 8.0, true},
		{"uint16", uint16(16), 16.0, true},
		{"uint32", uint32(32), 32.0, true},
		{"uint64", uint64(64), 64.0, true},
		{"float32", float32(1.5), 1.5, true},
		{"float64", float64(2.5), 2.5, true},
		{"string", "abc", 0, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.expected {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEqualsCrossTypeComparisons(t *testing.T) {
	t.Parallel()
	tests := append(testEqualsCrossTypeCasesInt(), testEqualsCrossTypeCasesFloat()...)

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
			}
		})
	}
}

type equalsCrossTypeCase struct {
	name     string
	left     any
	right    any
	expected bool
}

func testEqualsCrossTypeCasesInt() []equalsCrossTypeCase {
	return []equalsCrossTypeCase{
		{"int vs int8", int(8), int8(8), true},
		{"int vs int16", int(16), int16(16), true},
		{"int vs int32", int(32), int32(32), true},
		{"int vs int64", int(64), int64(64), true},
		{"int vs uint8", int(8), uint8(8), true},
		{"int vs uint16", int(16), uint16(16), true},
		{"int vs uint32", int(32), uint32(32), true},
		{"int vs float32", int(1), float32(1), true},
		{"int vs float64", int(1), float64(1), true},
		{"int64 vs int", int64(42), int(42), true},
		{"int64 vs int8", int64(8), int8(8), true},
		{"int64 vs int16", int64(16), int16(16), true},
		{"int64 vs int32", int64(32), int32(32), true},
		{"int64 vs uint8", int64(8), uint8(8), true},
		{"int64 vs uint16", int64(16), uint16(16), true},
		{"int64 vs uint32", int64(32), uint32(32), true},
		{"int64 vs float32", int64(1), float32(1), true},
		{"int64 vs float64", int64(1), float64(1), true},
		{"uint vs uint", uint(42), uint(42), true},
		{"uint vs uint8", uint(8), uint8(8), true},
		{"uint vs uint16", uint(16), uint16(16), true},
		{"uint vs uint32", uint(32), uint32(32), true},
		{"uint vs uint64", uint(64), uint64(64), true},
		{"uint vs float32", uint(1), float32(1), true},
		{"uint vs float64", uint(1), float64(1), true},
		{"uint64 vs int8", uint64(8), int8(8), true},
		{"uint64 vs int16", uint64(16), int16(16), true},
		{"uint64 vs int32", uint64(32), int32(32), true},
		{"uint64 vs uint", uint64(42), uint(42), true},
		{"uint64 vs uint8", uint64(8), uint8(8), true},
		{"uint64 vs uint16", uint64(16), uint16(16), true},
		{"uint64 vs uint32", uint64(32), uint32(32), true},
		{"uint64 vs uint64", uint64(64), uint64(64), true},
		{"uint64 vs float32", uint64(1), float32(1), true},
		{"uint64 vs float64", uint64(1), float64(1), true},
	}
}

func testEqualsCrossTypeCasesFloat() []equalsCrossTypeCase {
	return []equalsCrossTypeCase{
		{"float32 vs int", float32(42), int(42), true},
		{"float32 vs int8", float32(8), int8(8), true},
		{"float32 vs int16", float32(16), int16(16), true},
		{"float32 vs int32", float32(32), int32(32), true},
		{"float32 vs int64", float32(64), int64(64), true},
		{"float32 vs uint", float32(42), uint(42), true},
		{"float32 vs uint8", float32(8), uint8(8), true},
		{"float32 vs uint16", float32(16), uint16(16), true},
		{"float32 vs uint32", float32(32), uint32(32), true},
		{"float32 vs uint64", float32(64), uint64(64), true},
		{"float32 vs float64", float32(1.5), float64(1.5), true},
		{"float64 vs int", float64(42), int(42), true},
		{"float64 vs int8", float64(8), int8(8), true},
		{"float64 vs int16", float64(16), int16(16), true},
		{"float64 vs int32", float64(32), int32(32), true},
		{"float64 vs int64", float64(64), int64(64), true},
		{"float64 vs uint", float64(42), uint(42), true},
		{"float64 vs uint8", float64(8), uint8(8), true},
		{"float64 vs uint16", float64(16), uint16(16), true},
		{"float64 vs uint32", float64(32), uint32(32), true},
		{"float64 vs uint64", float64(64), uint64(64), true},
		{"float64 vs float32", float64(1.5), float32(1.5), true},
		{"int8 vs int", int8(42), int(42), true},
		{"int16 vs int", int16(42), int(42), true},
		{"int32 vs int", int32(42), int(42), true},
		{"uint8 vs uint", uint8(42), uint(42), true},
		{"uint16 vs uint", uint16(42), uint(42), true},
		{"uint32 vs uint", uint32(42), uint(42), true},
		{"float64 vs string", float64(1), "a", false},
	}
}

func TestCompareValuesUnsupported(t *testing.T) {
	t.Parallel()
	_, err := compareValues("a", "b", "~")
	if err == nil {
		t.Error("expected error for unsupported operator")
	}
}

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
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
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
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
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
			got := toFloat64Value(tt.input)
			if got != tt.expected {
				t.Errorf("toFloat64Value(%v) = %v, want %v", tt.input, got, tt.expected)
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
			got, err := compareValues(tt.left, tt.right, tt.op)
			if err != nil {
				t.Errorf("compareValues(%v, %v, %s) unexpected error: %v", tt.left, tt.right, tt.op, err)
			}
			if got != tt.expected {
				t.Errorf("compareValues(%v, %v, %s) = %v, want %v", tt.left, tt.right, tt.op, got, tt.expected)
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
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
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

func TestEvaluate_MethodCall_IntArg(t *testing.T) {
	t.Parallel()
	counter := &testCounter{Value: 10}

	expr, err := ParseExpression("Increment(5)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx := NewStandardEvaluationContext(counter)
	evaluated, err := expr.GetValue(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, ok := evaluated.(int)
	if !ok || got != 15 {
		t.Errorf("expected 15, got %v (%T)", evaluated, evaluated)
	}
}

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

func TestReflectPropertyAccessor_GetProperty_Unexported(t *testing.T) {
	t.Parallel()
	accessor := NewReflectPropertyAccessor()
	target := &testUnexported{hidden: "secret", Public: "open"}

	_, err := accessor.GetProperty(target, "hidden")
	if err == nil {
		t.Error("expected error for unexported field")
	}

	got, err := accessor.GetProperty(target, "Public")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "open" {
		t.Errorf("expected 'open', got %v", got)
	}
}

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

func TestEvaluate_MethodCall_MissingClosingParenRegression(t *testing.T) {
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
			if rec := recover(); rec != nil {
				t.Errorf("expected no panic, got: %v", rec)
			}
		}()
		_, evalErr = expr.GetValue(ctx)
	}()
	if evalErr == nil {
		t.Error("expected error for malformed method call")
	}
}

type testCounter struct {
	Value int
}

func (c *testCounter) Increment(amount int) int {
	c.Value += amount
	return c.Value
}

type testUnexported struct {
	hidden string
	Public string
}

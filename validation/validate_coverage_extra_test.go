package validation

import (
	"testing"
)

// TestValidate_Coverage 测试 Validate 函数
func TestValidate_Coverage(t *testing.T) {
	t.Parallel()

	// 测试 valid
	err := Validate("test", "required,min=2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 测试 invalid
	err = Validate("", "required")
	if err == nil {
		t.Error("Expected error for empty string with required rule")
	}

	// 测试空规则
	err = Validate("test", "")
	if err != nil {
		t.Errorf("Expected no error for empty rules, got %v", err)
	}

	// 测试多个规则（包含空规则）
	err = Validate("test", "required,,min=2")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

// TestValidateStruct_Extra 测试 ValidateStruct 函数
func TestValidateStruct_Extra(t *testing.T) {
	t.Parallel()

	type TestStruct struct {
		Name string `validate:"required"`
	}

	// 测试 valid
	s := TestStruct{Name: "test"}
	err := ValidateStruct(s)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// 测试 invalid
	s2 := TestStruct{Name: ""}
	err = ValidateStruct(s2)
	if err == nil {
		t.Error("Expected error for empty name")
	}
}

// TestIsRequiredValidForValue 测试 isRequiredValidForValue
func testIsRequiredValidForValueCases() []struct {
	name   string
	value  any
	want   bool
	errMsg string
} {
	value := 1
	ptr := (*int)(nil)
	return []struct {
		name   string
		value  any
		want   bool
		errMsg string
	}{
		{"nil", nil, false, "Expected nil to be invalid"},
		{"non-empty string", "test", true, "Expected non-empty string to be valid"},
		{"empty string", "", false, "Expected empty string to be invalid"},
		{"non-zero int", 1, true, "Expected non-zero int to be valid"},
		{"zero int", 0, false, "Expected zero int to be invalid"},
		{"non-zero uint", uint(1), true, "Expected non-zero uint to be valid"},
		{"zero uint", uint(0), false, "Expected zero uint to be invalid"},
		{"non-zero float", 1.0, true, "Expected non-zero float to be valid"},
		{"zero float", 0.0, false, "Expected zero float to be invalid"},
		{"true", true, true, "Expected true to be valid"},
		{"false", false, false, "Expected false to be invalid"},
		{"non-nil slice", []int{1}, true, "Expected non-nil slice to be valid"},
		{"nil slice", []int(nil), false, "Expected nil slice to be invalid"},
		{"non-nil map", map[string]int{"a": 1}, true, "Expected non-nil map to be valid"},
		{"nil map", (map[string]int)(nil), false, "Expected nil map to be invalid"},
		{"non-nil pointer", &value, true, "Expected non-nil pointer to be valid"},
		{"nil pointer", ptr, false, "Expected nil pointer to be invalid"},
	}
}

func TestIsRequiredValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	for _, tt := range testIsRequiredValidForValueCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isRequiredValidForValue(tt.value); got != tt.want {
				t.Errorf("%s", tt.errMsg)
			}
		})
	}
}

// TestIsMinValidForValue 测试 isMinValidForValue
func TestIsMinValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isMinValidForValue("test", "3") {
		t.Error("Expected 'test' to satisfy min=3")
	}
	if isMinValidForValue("ab", "3") {
		t.Error("Expected 'ab' to not satisfy min=3")
	}

	// int
	if !isMinValidForValue(25, "18") {
		t.Error("Expected 25 to satisfy min=18")
	}
	if isMinValidForValue(10, "18") {
		t.Error("Expected 10 to not satisfy min=18")
	}

	// uint
	if !isMinValidForValue(uint(10), "5") {
		t.Error("Expected uint(10) to satisfy min=5")
	}

	// float
	if !isMinValidForValue(15.0, "10") {
		t.Error("Expected 15.0 to satisfy min=10")
	}

	// invalid min
	if isMinValidForValue("test", "abc") {
		t.Error("Expected invalid min value to return false")
	}

	// unsupported type
	if isMinValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported for min")
	}
}

// TestIsMaxValidForValue 测试 isMaxValidForValue
func TestIsMaxValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isMaxValidForValue("test", "10") {
		t.Error("Expected 'test' to satisfy max=10")
	}
	if isMaxValidForValue("this is a very long string", "10") {
		t.Error("Expected long string to not satisfy max=10")
	}

	// int
	if !isMaxValidForValue(50, "100") {
		t.Error("Expected 50 to satisfy max=100")
	}
	if isMaxValidForValue(150, "100") {
		t.Error("Expected 150 to not satisfy max=100")
	}

	// uint
	if !isMaxValidForValue(uint(50), "100") {
		t.Error("Expected uint(50) to satisfy max=100")
	}

	// float
	if !isMaxValidForValue(50.0, "100") {
		t.Error("Expected 50.0 to satisfy max=100")
	}

	// invalid max
	if isMaxValidForValue("test", "abc") {
		t.Error("Expected invalid max value to return false")
	}

	// unsupported type
	if isMaxValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported for max")
	}
}

// TestIsLenValidForValue 测试 isLenValidForValue
func TestIsLenValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isLenValidForValue("1234", "4") {
		t.Error("Expected '1234' to satisfy len=4")
	}
	if isLenValidForValue("123", "4") {
		t.Error("Expected '123' to not satisfy len=4")
	}

	// slice
	if !isLenValidForValue([]int{1, 2, 3}, "3") {
		t.Error("Expected slice with 3 items to satisfy len=3")
	}

	// invalid length
	if isLenValidForValue("test", "abc") {
		t.Error("Expected invalid length value to return false")
	}

	// unsupported type
	if isLenValidForValue(123, "3") {
		t.Error("Expected int to be unsupported for len")
	}
}

// TestIsEmailValidForValue 测试 isEmailValidForValue
func TestIsEmailValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// nil
	if isEmailValidForValue(nil) {
		t.Error("Expected nil to be invalid")
	}

	// valid email
	if !isEmailValidForValue("test@example.com") {
		t.Error("Expected valid email")
	}

	// invalid email
	if isEmailValidForValue("invalid-email") {
		t.Error("Expected invalid email")
	}

	// non-string
	if isEmailValidForValue(123) {
		t.Error("Expected non-string to be invalid")
	}
}

// TestIsRegexpValidForValue 测试 isRegexpValidForValue
func TestIsRegexpValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// nil
	if isRegexpValidForValue(nil, ".*") {
		t.Error("Expected nil to be invalid")
	}

	// valid pattern
	if !isRegexpValidForValue("ABC", "^[A-Z]{3}$") {
		t.Error("Expected pattern to match")
	}

	// invalid pattern match
	if isRegexpValidForValue("abc", "^[A-Z]{3}$") {
		t.Error("Expected pattern to not match")
	}

	// invalid regexp
	if isRegexpValidForValue("test", "[invalid") {
		t.Error("Expected invalid regexp to return false")
	}

	// non-string
	if isRegexpValidForValue(123, ".*") {
		t.Error("Expected non-string to be invalid")
	}
}

// TestIsGtValidForValue 测试 isGtValidForValue
func TestIsGtValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isGtValidForValue("test", "3") {
		t.Error("Expected 'test' (len=4) to satisfy gt=3")
	}
	if isGtValidForValue("abc", "3") {
		t.Error("Expected 'abc' (len=3) to not satisfy gt=3")
	}

	// int
	if !isGtValidForValue(25, "18") {
		t.Error("Expected 25 to satisfy gt=18")
	}
	if isGtValidForValue(18, "18") {
		t.Error("Expected 18 to not satisfy gt=18")
	}

	// uint
	if !isGtValidForValue(uint(10), "5") {
		t.Error("Expected uint(10) to satisfy gt=5")
	}

	// float
	if !isGtValidForValue(15.0, "10") {
		t.Error("Expected 15.0 to satisfy gt=10")
	}

	// invalid value
	if isGtValidForValue("test", "abc") {
		t.Error("Expected invalid value to return false")
	}

	// unsupported type
	if isGtValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported")
	}
}

// TestIsGteValidForValue 测试 isGteValidForValue
func TestIsGteValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isGteValidForValue("test", "4") {
		t.Error("Expected 'test' (len=4) to satisfy gte=4")
	}
	if isGteValidForValue("abc", "4") {
		t.Error("Expected 'abc' (len=3) to not satisfy gte=4")
	}

	// int
	if !isGteValidForValue(18, "18") {
		t.Error("Expected 18 to satisfy gte=18")
	}
	if isGteValidForValue(17, "18") {
		t.Error("Expected 17 to not satisfy gte=18")
	}

	// uint
	if !isGteValidForValue(uint(10), "10") {
		t.Error("Expected uint(10) to satisfy gte=10")
	}

	// float
	if !isGteValidForValue(10.0, "10") {
		t.Error("Expected 10.0 to satisfy gte=10")
	}

	// invalid value
	if isGteValidForValue("test", "abc") {
		t.Error("Expected invalid value to return false")
	}

	// unsupported type
	if isGteValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported")
	}
}

// TestIsLtValidForValue 测试 isLtValidForValue
func TestIsLtValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isLtValidForValue("abc", "4") {
		t.Error("Expected 'abc' (len=3) to satisfy lt=4")
	}
	if isLtValidForValue("test", "4") {
		t.Error("Expected 'test' (len=4) to not satisfy lt=4")
	}

	// int
	if !isLtValidForValue(50, "100") {
		t.Error("Expected 50 to satisfy lt=100")
	}
	if isLtValidForValue(100, "100") {
		t.Error("Expected 100 to not satisfy lt=100")
	}

	// uint
	if !isLtValidForValue(uint(50), "100") {
		t.Error("Expected uint(50) to satisfy lt=100")
	}

	// float
	if !isLtValidForValue(50.0, "100") {
		t.Error("Expected 50.0 to satisfy lt=100")
	}

	// invalid value
	if isLtValidForValue("test", "abc") {
		t.Error("Expected invalid value to return false")
	}

	// unsupported type
	if isLtValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported")
	}
}

// TestIsLteValidForValue 测试 isLteValidForValue
func TestIsLteValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string
	if !isLteValidForValue("test", "4") {
		t.Error("Expected 'test' (len=4) to satisfy lte=4")
	}
	if isLteValidForValue("tests", "4") {
		t.Error("Expected 'tests' (len=5) to not satisfy lte=4")
	}

	// int
	if !isLteValidForValue(100, "100") {
		t.Error("Expected 100 to satisfy lte=100")
	}
	if isLteValidForValue(101, "100") {
		t.Error("Expected 101 to not satisfy lte=100")
	}

	// uint
	if !isLteValidForValue(uint(100), "100") {
		t.Error("Expected uint(100) to satisfy lte=100")
	}

	// float
	if !isLteValidForValue(100.0, "100") {
		t.Error("Expected 100.0 to satisfy lte=100")
	}

	// invalid value
	if isLteValidForValue("test", "abc") {
		t.Error("Expected invalid value to return false")
	}

	// unsupported type
	if isLteValidForValue(true, "1") {
		t.Error("Expected bool to be unsupported")
	}
}

package validation

import (
	"strings"
	"testing"
)

// TestRegexpValidationWithComma 验证包含逗号的正则表达式规则（回归测试）
//
// 背景：规则按逗号分割时，regexp=^[0-9]{1,3}$ 会被拆成 "regexp=^[0-9]{1" 和 "3}$"，
// 导致正则表达式被破坏。
// 注意：Go 的 reflect.StructTag 使用 strconv.Unquote 解析标签值，
// 因此标签中的 \d 必须写成 \\d（否则标签值为空），测试中使用 [0-9] 避免歧义。
func TestRegexpValidationWithComma(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type TestStruct struct {
		Code string `validate:"regexp=^[0-9]{1,3}$"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "123", false},
		{"valid single digit", "7", false},
		{"too many digits", "1234", true},
		{"non digit", "abc", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Validate(TestStruct{Code: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("value %q: got error %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// TestRegexpValidationCombined 正则表达式规则与其他规则组合（回归测试）
func TestRegexpValidationCombined(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type TestStruct struct {
		Code string `validate:"required,regexp=^[A-Z]{1,3}$"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "AB", false},
		{"too long", "ABCD", true},
		{"lowercase", "ab", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Validate(TestStruct{Code: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("value %q: got error %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

// TestRegexpValidationInvalidPattern 无效正则表达式不应导致 panic（回归测试）
//
// 背景：compileRegex 对无效模式返回 nil，isRegexpValid 直接调用 MatchString
// 导致 nil 指针 panic。修复后应返回验证错误而非 panic。
func TestRegexpValidationInvalidPattern(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type TestStruct struct {
		Code string `validate:"regexp=[invalid("`
	}

	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("无效正则表达式不应触发 panic，got: %v", r)
			}
		}()
		err := validator.Validate(TestStruct{Code: "x"})
		if err == nil {
			t.Error("预期无效正则表达式应产生验证错误")
		}
	}()
}

// TestRegexpValidationErrorContainsPattern 验证错误消息应包含原始正则表达式
func TestRegexpValidationErrorContainsPattern(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type TestStruct struct {
		Code string `validate:"regexp=^[0-9]{1,3}$"`
	}

	err := validator.Validate(TestStruct{Code: "1234"})
	if err == nil {
		t.Fatal("预期验证错误")
	}
	if !strings.Contains(err.Error(), "^[0-9]{1,3}$") {
		t.Errorf("错误消息应包含完整正则表达式，got: %v", err)
	}
}

// TestRegexpValidationEscapedBackslash 验证使用 \\d 转义的反斜杠模式（回归测试）
//
// Go 的 StructTag 解析会将 \\d 转义为 \d，确保该场景下规则也能正确解析。
func TestRegexpValidationEscapedBackslash(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type TestStruct struct {
		Code string `validate:"regexp=^\\d{1,3}$"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "123", false},
		{"too many digits", "1234", true},
		{"non digit", "abc", true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validator.Validate(TestStruct{Code: tt.value})
			if (err != nil) != tt.wantErr {
				t.Errorf("value %q: got error %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

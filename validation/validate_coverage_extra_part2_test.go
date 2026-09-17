package validation

import (
	"testing"
)

// TestIsURLValidForValue 测试 isURLValidForValue
func TestIsURLValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// nil
	if isURLValidForValue(nil) {
		t.Error("Expected nil to be invalid")
	}

	// valid URL
	if !isURLValidForValue("https://example.com") {
		t.Error("Expected valid URL")
	}

	// invalid URL
	if isURLValidForValue("not-a-url") {
		t.Error("Expected invalid URL")
	}

	// non-string
	if isURLValidForValue(123) {
		t.Error("Expected non-string to be invalid")
	}
}

// TestIsIPValidForValue 测试 isIPValidForValue
func TestIsIPValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// nil
	if isIPValidForValue(nil) {
		t.Error("Expected nil to be invalid")
	}

	// valid IP
	if !isIPValidForValue("192.168.1.1") {
		t.Error("Expected valid IP")
	}

	// invalid IP
	if isIPValidForValue("not-an-ip") {
		t.Error("Expected invalid IP")
	}

	// non-string
	if isIPValidForValue(123) {
		t.Error("Expected non-string to be invalid")
	}
}

// TestIsOneOfValidForValue 测试 isOneOfValidForValue
func TestIsOneOfValidForValue_Coverage(t *testing.T) {
	t.Parallel()

	// string - valid
	if !isOneOfValidForValue("green", "red green blue") {
		t.Error("Expected 'green' to be in options")
	}

	// string - invalid
	if isOneOfValidForValue("yellow", "red green blue") {
		t.Error("Expected 'yellow' to not be in options")
	}

	// int - valid
	if !isOneOfValidForValue(2, "1 2 3") {
		t.Error("Expected 2 to be in options")
	}

	// int - invalid
	if isOneOfValidForValue(4, "1 2 3") {
		t.Error("Expected 4 to not be in options")
	}

	// nil
	if isOneOfValidForValue(nil, "1 2 3") {
		t.Error("Expected nil to be invalid")
	}
}

// TestValidateRuleWithParam 测试 validateRuleWithParam
func testValidateRuleWithParamCases() []struct {
	name    string
	value   any
	rule    string
	wantErr bool
	errMsg  string
} {
	return []struct {
		name    string
		value   any
		rule    string
		wantErr bool
		errMsg  string
	}{
		{"valid min", "test", "min=3", false, ""},
		{"invalid min", "ab", "min=3", true, "Expected error for invalid min"},
		{"valid max", "test", "max=10", false, ""},
		{"valid len", "test", "len=4", false, ""},
		{"valid email", "test@example.com", "email=true", false, ""},
		{"valid regexp", "ABC", "regexp=^[A-Z]{3}$", false, ""},
		{"valid gt", 25, "gt=18", false, ""},
		{"valid gte", 18, "gte=18", false, ""},
		{"valid lt", 50, "lt=100", false, ""},
		{"valid lte", 100, "lte=100", false, ""},
		{"valid url", "https://example.com", "url=true", false, ""},
		{"valid ip", "192.168.1.1", "ip=true", false, ""},
		{"valid oneof", "green", "oneof=red green blue", false, ""},
		{"unknown rule", "test", "unknown=value", false, ""},
	}
}

func TestValidateRuleWithParam_Coverage(t *testing.T) {
	t.Parallel()

	for _, tt := range testValidateRuleWithParamCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRuleWithParam(tt.value, tt.rule)
			if (err != nil) != tt.wantErr {
				if tt.errMsg != "" {
					t.Errorf("%s", tt.errMsg)
				} else {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

// TestValidateRuleWithoutParam 测试 validateRuleWithoutParam
func testValidateRuleWithoutParamCases() []struct {
	name    string
	value   any
	rule    string
	wantErr bool
	errMsg  string
} {
	return []struct {
		name    string
		value   any
		rule    string
		wantErr bool
		errMsg  string
	}{
		{"valid required", "test", "required", false, ""},
		{"invalid required", "", "required", true, "Expected error for empty string with required"},
		{"valid email", "test@example.com", "email", false, ""},
		{"invalid email", "invalid", "email", true, "Expected error for invalid email"},
		{"valid url", "https://example.com", "url", false, ""},
		{"invalid url", "not-a-url", "url", true, "Expected error for invalid URL"},
		{"valid ip", "192.168.1.1", "ip", false, ""},
		{"invalid ip", "not-an-ip", "ip", true, "Expected error for invalid IP"},
		{"unknown rule", "test", "unknown", false, ""},
	}
}

func TestValidateRuleWithoutParam_Coverage(t *testing.T) {
	t.Parallel()

	for _, tt := range testValidateRuleWithoutParamCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := validateRuleWithoutParam(tt.value, tt.rule)
			if (err != nil) != tt.wantErr {
				if tt.errMsg != "" {
					t.Errorf("%s", tt.errMsg)
				} else {
					t.Errorf("Expected no error, got %v", err)
				}
			}
		})
	}
}

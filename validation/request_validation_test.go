package validation

import (
	"net/http"
	"net/url"
	"testing"
)

func TestRequestValidator_Required(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试缺失必填字段
	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for missing required field")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(result.Errors))
	}

	// 测试存在必填字段
	req = &http.Request{URL: &url.URL{RawQuery: "name=test"}}
	result = validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for present required field")
	}
}

func TestRequestValidator_String(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "string", MinLength: intPtr(2), MaxLength: intPtr(10)},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试有效字符串
	req := &http.Request{URL: &url.URL{RawQuery: "name=test"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid string")
	}

	// 测试字符串过短
	req = &http.Request{URL: &url.URL{RawQuery: "name=a"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for too short string")
	}

	// 测试字符串过长
	req = &http.Request{URL: &url.URL{RawQuery: "name=verylongstring"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for too long string")
	}
}

func TestRequestValidator_Email(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "email", Type: "email"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试有效邮箱
	req := &http.Request{URL: &url.URL{RawQuery: "email=test@example.com"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid email")
	}

	// 测试无效邮箱
	req = &http.Request{URL: &url.URL{RawQuery: "email=invalid"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid email")
	}
}

func TestRequestValidator_Enum(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "status", Type: "enum", In: []string{"active", "inactive"}},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试有效枚举值
	req := &http.Request{URL: &url.URL{RawQuery: "status=active"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid enum value")
	}

	// 测试无效枚举值
	req = &http.Request{URL: &url.URL{RawQuery: "status=unknown"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestRequestValidator_FailFast(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required"},
			{Field: "email", Type: "email"},
		},
		Source:   "query",
		FailFast: true,
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试快速失败 - 应在第一个错误时停止
	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
	if len(result.Errors) != 1 {
		t.Errorf("Expected 1 error with fail fast, got %d", len(result.Errors))
	}
}

func TestRequestValidator_CustomMessage(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "name", Type: "required", Message: "Name is required, please provide it"},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	req := &http.Request{URL: &url.URL{}}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
	if result.Errors[0].Message != "Name is required, please provide it" {
		t.Errorf("Expected custom message, got: %s", result.Errors[0].Message)
	}
}

func TestRequestValidator_Regex(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "phone", Type: "regex", Pattern: `^\d{3}-\d{3}-\d{4}$`},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试有效手机号
	req := &http.Request{URL: &url.URL{RawQuery: "phone=123-456-7890"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid phone")
	}

	// 测试无效手机号
	req = &http.Request{URL: &url.URL{RawQuery: "phone=1234567890"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for invalid phone")
	}
}

func TestRequestValidator_InvalidRegex(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "test", Type: "regex", Pattern: `[invalid`},
		},
		Source: "query",
	}

	_, err := NewRequestValidator(config)
	if err == nil {
		t.Error("Expected error for invalid regex pattern")
	}
}

func TestRequestValidator_Number(t *testing.T) {
	t.Parallel()
	minVal := 1.0
	maxVal := 100.0
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "age", Type: "number", Min: &minVal, Max: &maxVal},
		},
		Source: "query",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试有效数字
	req := &http.Request{URL: &url.URL{RawQuery: "age=25"}}
	result := validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for valid number")
	}

	// 测试超出范围
	req = &http.Request{URL: &url.URL{RawQuery: "age=200"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for out of range number")
	}

	// 测试非数字
	req = &http.Request{URL: &url.URL{RawQuery: "age=abc"}}
	result = validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for non-number")
	}
}

func TestRequestValidator_Header(t *testing.T) {
	t.Parallel()
	config := ValidationConfig{
		Rules: []ValidationRule{
			{Field: "X-Api-Key", Type: "required"},
		},
		Source: "header",
	}

	validator, err := NewRequestValidator(config)
	if err != nil {
		t.Fatalf("Failed to create validator: %v", err)
	}

	// 测试缺失请求头
	req := &http.Request{Header: make(http.Header)}
	result := validator.Validate(req)
	if result.Valid {
		t.Error("Expected validation to fail for missing header")
	}

	// 测试存在请求头
	req = &http.Request{Header: http.Header{"X-Api-Key": []string{"secret-key"}}}
	result = validator.Validate(req)
	if !result.Valid {
		t.Error("Expected validation to pass for present header")
	}
}

func TestValidateJSONBody(t *testing.T) {
	t.Parallel()
	rules := []ValidationRule{
		{Field: "name", Type: "required"},
		{Field: "email", Type: "email"},
	}

	// 测试有效 JSON
	body := []byte(`{"name":"test","email":"test@example.com"}`)
	result := ValidateJSONBody(body, rules)
	if !result.Valid {
		t.Errorf("Expected validation to pass, got errors: %v", result.Errors)
	}

	// 测试无效 JSON
	body = []byte(`{"name":"","email":"invalid"}`)
	result = ValidateJSONBody(body, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}

	// 测试格式错误的 JSON
	body = []byte(`{invalid json}`)
	result = ValidateJSONBody(body, rules)
	if result.Valid {
		t.Error("Expected validation to fail for malformed JSON")
	}
}

func TestValidateHeaders(t *testing.T) {
	t.Parallel()
	rules := []ValidationRule{
		{Field: "X-Request-Id", Type: "required"},
	}

	req := &http.Request{Header: http.Header{"X-Request-Id": []string{"123"}}}
	result := ValidateHeaders(req, rules)
	if !result.Valid {
		t.Error("Expected validation to pass")
	}

	req = &http.Request{Header: make(http.Header)}
	result = ValidateHeaders(req, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
}

func TestValidateQuery(t *testing.T) {
	t.Parallel()
	rules := []ValidationRule{
		{Field: "page", Type: "required"},
	}

	req := &http.Request{URL: &url.URL{RawQuery: "page=1"}}
	result := ValidateQuery(req, rules)
	if !result.Valid {
		t.Error("Expected validation to pass")
	}

	req = &http.Request{URL: &url.URL{}}
	result = ValidateQuery(req, rules)
	if result.Valid {
		t.Error("Expected validation to fail")
	}
}

func intPtr(i int) *int {
	return &i
}

package config

import (
	"testing"
)

func TestValidationError_Error_Coverage(t *testing.T) {
	t.Parallel()
	err := ValidationError{Field: "port", Message: "value below minimum"}
	expected := "field port: value below minimum"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationErrors_Error_Coverage(t *testing.T) {
	t.Parallel()
	errs := ValidationErrors{
		{Field: "name", Message: "required"},
		{Field: "port", Message: "invalid"},
	}
	errorStr := errs.Error()
	if errorStr == "" {
		t.Error("expected non-empty error string")
	}
	if len(errorStr) < 20 {
		t.Errorf("expected substantial error string, got '%s'", errorStr)
	}
}

func TestValidationErrors_Error_Empty(t *testing.T) {
	t.Parallel()
	errs := ValidationErrors{}
	if errs.Error() != "" {
		t.Errorf("expected empty string for empty errors, got '%s'", errs.Error())
	}
}

func TestDefaultValidator_AddMin_Int(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMin("port", 10)

	payload := map[string]any{"port": 5}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for int below minimum")
	}

	payload["port"] = 10
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	payload["port"] = 20
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_Float64(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMin("rate", 5)

	payload := map[string]any{"rate": 3.5}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for float64 below minimum")
	}

	payload["rate"] = 5.0
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_String(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMin("name", 3)

	payload := map[string]any{"name": "ab"}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for string length below minimum")
	}

	payload["name"] = "abc"
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_UnsupportedType(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMin("val", 10)

	payload := map[string]any{"val": true}
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error for unsupported type: %v", err)
	}
}

func TestDefaultValidator_AddMax_Int(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMax("port", 100)

	payload := map[string]any{"port": 200}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for int above maximum")
	}

	payload["port"] = 100
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	payload["port"] = 50
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_Float64(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMax("rate", 10)

	payload := map[string]any{"rate": 15.0}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for float64 above maximum")
	}

	payload["rate"] = 10.0
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_String(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMax("name", 5)

	payload := map[string]any{"name": "toolongname"}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for string length above maximum")
	}

	payload["name"] = "ok"
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_UnsupportedType(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddMax("val", 10)

	payload := map[string]any{"val": []int{1, 2, 3}}
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error for unsupported type: %v", err)
	}
}

func TestDefaultValidator_AddRegex(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRegex("email", `^[a-z]+@[a-z]+\.[a-z]+$`)

	payload := map[string]any{"email": "Test@Example.com"}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for non-matching email (uppercase)")
	}

	payload["email"] = "user@domain.com"
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddRegex_NonString(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRegex("email", `^[a-z]+$`)

	payload := map[string]any{"email": 123}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for non-string value")
	}
}

func TestDefaultValidator_AddRegex_InvalidPattern(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRegex("email", `[invalid`)

	payload := map[string]any{"email": "test"}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestDefaultValidator_AddEnum(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddEnum("status", "active", "inactive")

	payload := map[string]any{"status": "active"}
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	payload["status"] = "deleted"
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for non-matching enum value")
	}
}

func TestDefaultValidator_AddCustomRule(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddCustomRule("age", func(value any) error {
		age, ok := value.(int)
		if !ok {
			return &testValErr{msg: "not an int"}
		}
		if age < 0 || age > 150 {
			return &testValErr{msg: "out of range"}
		}
		return nil
	})

	payload := map[string]any{"age": 25}
	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	payload["age"] = 200
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for age out of range")
	}
}

func TestDefaultValidator_Validate_FieldNotFound(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRequired("missing")

	payload := map[string]any{}
	if err := validator.Validate(payload); err == nil {
		t.Error("expected error for missing field")
	}
}

func TestDefaultValidator_Validate_MultipleErrors(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRequired("name", "email")
	validator.AddMin("age", 0)

	payload := map[string]any{
		"age": -1,
	}

	err := validator.Validate(payload)
	if err == nil {
		t.Fatal("expected validation errors")
	}

	validationErrs, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	if len(validationErrs) != 3 {
		t.Errorf("expected 3 errors, got %d", len(validationErrs))
	}
}

func TestDefaultValidator_Validate_NoErrors(t *testing.T) {
	t.Parallel()
	validator := NewValidator()
	validator.AddRequired("name")
	validator.AddMin("age", 0)
	validator.AddMax("age", 150)

	payload := map[string]any{
		"name": "John",
		"age":  25,
	}

	if err := validator.Validate(payload); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestWatchManager_AddAndGetSource(t *testing.T) {
	t.Parallel()
	mgr := NewWatchManager()
	ch := make(chan WatchEvent, 1)
	mgr.AddSource("nacos", ch)

	got, ok := mgr.GetSource("nacos")
	if !ok {
		t.Fatal("expected to find source 'nacos'")
	}
	if got != ch {
		t.Error("expected channel to match")
	}
}

func TestWatchManager_GetSource_NotFound(t *testing.T) {
	t.Parallel()
	mgr := NewWatchManager()
	_, ok := mgr.GetSource("nonexistent")
	if ok {
		t.Error("expected false for nonexistent source")
	}
}

func TestWatchManager_Close_PreventsNewOps(t *testing.T) {
	t.Parallel()
	mgr := NewWatchManager()
	mgr.Close()

	// 这些操作在关闭后应静默忽略
	mgr.AddSource("test", make(chan WatchEvent))
	mgr.Register("test", func(event WatchEvent) {})
	mgr.Unregister("test")

	// Notify 也不应 panic
	mgr.Notify(WatchEvent{Type: EventModify})
}

func TestWatchManager_Notify_NilCallback(t *testing.T) {
	t.Parallel()
	mgr := NewWatchManager()
	mgr.Register("nil-cb", nil)

	// 不应 panic
	mgr.Notify(WatchEvent{Type: EventModify})
}

func TestConfig_GetAll(t *testing.T) {
	t.Parallel()
	c := NewConfig()
	c.Set("a", 1)
	c.Set("b", "two")

	all := c.GetAll()
	if len(all) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(all))
	}
	if all["a"] != 1 {
		t.Errorf("expected a=1, got %v", all["a"])
	}
	if all["b"] != "two" {
		t.Errorf("expected b=two, got %v", all["b"])
	}
}

func TestConfig_Watch(t *testing.T) {
	t.Parallel()
	c := NewConfig()

	key := "test-watch-key"
	var received bool
	Watch(key, func(event WatchEvent) {
		received = true
	})

	c.Set(key, "value")

	if !received {
		t.Error("expected watch callback to be invoked")
	}
}

type testValErr struct {
	msg string
}

func (e *testValErr) Error() string { return e.msg }

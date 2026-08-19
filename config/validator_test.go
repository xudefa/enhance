package config

import (
	"testing"
)

func TestValidationError_Error(t *testing.T) {
	t.Parallel()
	err := ValidationError{Field: "port", Message: "value below minimum"}
	expected := "field port: value below minimum"
	if err.Error() != expected {
		t.Errorf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestValidationErrors_Error(t *testing.T) {
	t.Parallel()
	errs := ValidationErrors{
		{Field: "name", Message: "required"},
		{Field: "port", Message: "invalid"},
	}
	result := errs.Error()
	if result == "" {
		t.Error("expected non-empty error string")
	}
	if len(result) < 20 {
		t.Errorf("expected substantial error string, got '%s'", result)
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
	v := NewValidator()
	v.AddMin("port", 10)

	data := map[string]any{"port": 5}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for int below minimum")
	}

	data["port"] = 10
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["port"] = 20
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_Float64(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMin("rate", 5)

	data := map[string]any{"rate": 3.5}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for float64 below minimum")
	}

	data["rate"] = 5.0
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_String(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMin("name", 3)

	data := map[string]any{"name": "ab"}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for string length below minimum")
	}

	data["name"] = "abc"
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMin_UnsupportedType(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMin("val", 10)

	data := map[string]any{"val": true}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error for unsupported type: %v", err)
	}
}

func TestDefaultValidator_AddMax_Int(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMax("port", 100)

	data := map[string]any{"port": 200}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for int above maximum")
	}

	data["port"] = 100
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["port"] = 50
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_Float64(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMax("rate", 10)

	data := map[string]any{"rate": 15.0}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for float64 above maximum")
	}

	data["rate"] = 10.0
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_String(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMax("name", 5)

	data := map[string]any{"name": "toolongname"}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for string length above maximum")
	}

	data["name"] = "ok"
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddMax_UnsupportedType(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddMax("val", 10)

	data := map[string]any{"val": []int{1, 2, 3}}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error for unsupported type: %v", err)
	}
}

func TestDefaultValidator_AddRegex(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddRegex("email", `^[a-z]+@[a-z]+\.[a-z]+$`)

	data := map[string]any{"email": "Test@Example.com"}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for non-matching email (uppercase)")
	}

	data["email"] = "user@domain.com"
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestDefaultValidator_AddRegex_NonString(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddRegex("email", `^[a-z]+$`)

	data := map[string]any{"email": 123}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for non-string value")
	}
}

func TestDefaultValidator_AddRegex_InvalidPattern(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddRegex("email", `[invalid`)

	data := map[string]any{"email": "test"}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for invalid regex pattern")
	}
}

func TestDefaultValidator_AddEnum(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddEnum("status", "active", "inactive")

	data := map[string]any{"status": "active"}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["status"] = "deleted"
	if err := v.Validate(data); err == nil {
		t.Error("expected error for non-matching enum value")
	}
}

func TestDefaultValidator_AddCustomRule(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddCustomRule("age", func(value any) error {
		age, ok := value.(int)
		if !ok {
			return &testValErr{msg: "not an int"}
		}
		if age < 0 || age > 150 {
			return &testValErr{msg: "out of range"}
		}
		return nil
	})

	data := map[string]any{"age": 25}
	if err := v.Validate(data); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data["age"] = 200
	if err := v.Validate(data); err == nil {
		t.Error("expected error for age out of range")
	}
}

func TestDefaultValidator_Validate_FieldNotFound(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddRequired("missing")

	data := map[string]any{}
	if err := v.Validate(data); err == nil {
		t.Error("expected error for missing field")
	}
}

func TestDefaultValidator_Validate_MultipleErrors(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	v.AddRequired("name", "email")
	v.AddMin("age", 0)

	data := map[string]any{
		"age": -1,
	}

	err := v.Validate(data)
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
	v := NewValidator()
	v.AddRequired("name")
	v.AddMin("age", 0)
	v.AddMax("age", 150)

	data := map[string]any{
		"name": "John",
		"age":  25,
	}

	if err := v.Validate(data); err != nil {
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

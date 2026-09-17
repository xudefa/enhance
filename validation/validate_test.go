package validation

import (
	"testing"
)

func TestValidateStruct_Basic(t *testing.T) {
	t.Parallel()
	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
	}

	user := User{
		Name:  "John",
		Email: "john@example.com",
	}

	err := ValidateStruct(&user)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateStruct_InvalidStruct(t *testing.T) {
	t.Parallel()
	type User struct {
		Name  string `validate:"required"`
		Email string `validate:"email"`
	}

	user := User{
		Name:  "",
		Email: "invalid-email",
	}

	err := ValidateStruct(&user)
	if err == nil {
		t.Error("expected error for invalid struct")
	}
}

func testValidateStringCases() []struct {
	name    string
	value   any
	rules   string
	wantErr bool
} {
	return []struct {
		name    string
		value   any
		rules   string
		wantErr bool
	}{
		{"valid string", "hello", "required,min=3,max=10", false},
		{"invalid min", "hi", "min=5", true},
		{"invalid max", "hello world", "max=5", true},
		{"valid email", "user@example.com", "email", false},
		{"invalid email", "invalid", "email", true},
		{"valid url", "https://example.com", "url", false},
		{"invalid url", "not-a-url", "url", true},
		{"valid ip", "192.168.1.1", "ip", false},
		{"invalid ip", "999.999.999.999", "ip", true},
		{"valid regexp", "hello123", "regexp=^[a-z]+\\d+$", false},
		{"invalid regexp", "hello", "regexp=^\\d+$", true},
		{"valid oneof", "admin", "oneof=admin user guest", false},
		{"invalid oneof", "superadmin", "oneof=admin user guest", true},
		{"valid gt", 42, "gt=10", false},
		{"invalid gt", 5, "gt=10", true},
		{"valid gte", 10, "gte=10", false},
		{"valid lt", 5, "lt=10", false},
		{"invalid lt", 10, "lt=10", true},
		{"valid lte", 10, "lte=10", false},
		{"valid len", "hello", "len=5", false},
		{"invalid len", "hello", "len=3", true},
		{"required nil", nil, "required", true},
		{"empty rules", "hello", "", false},
		{"unknown rule", "hello", "unknown", false},
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()
	for _, tt := range testValidateStringCases() {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := Validate(tt.value, tt.rules)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate(%v, %q) error = %v, wantErr %v", tt.value, tt.rules, err, tt.wantErr)
			}
		})
	}
}

func TestValidate_IntValues(t *testing.T) {
	t.Parallel()
	t.Run("valid int min", func(t *testing.T) {
		t.Parallel()
		err := Validate(42, "min=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid int max", func(t *testing.T) {
		t.Parallel()
		err := Validate(5, "max=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid uint min", func(t *testing.T) {
		t.Parallel()
		err := Validate(uint(42), "min=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid float min", func(t *testing.T) {
		t.Parallel()
		err := Validate(42.5, "min=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}

func TestValidate_OneOfInt(t *testing.T) {
	t.Parallel()
	err := Validate(1, "oneof=1 2 3")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}

	err = Validate(5, "oneof=1 2 3")
	if err == nil {
		t.Error("expected error for value not in options")
	}
}

func TestValidate_RequiredPtr(t *testing.T) {
	t.Parallel()
	var ptr *int = nil
	err := Validate(ptr, "required")
	if err == nil {
		t.Error("expected error for nil pointer")
	}

	x := 42
	ptr = &x
	err = Validate(ptr, "required")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

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

func TestValidate(t *testing.T) {
	t.Parallel()
	t.Run("valid string", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "required,min=3,max=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid min", func(t *testing.T) {
		t.Parallel()
		err := Validate("hi", "min=5")
		if err == nil {
			t.Error("expected error for string too short")
		}
	})

	t.Run("invalid max", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello world", "max=5")
		if err == nil {
			t.Error("expected error for string too long")
		}
	})

	t.Run("valid email", func(t *testing.T) {
		t.Parallel()
		err := Validate("user@example.com", "email")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid email", func(t *testing.T) {
		t.Parallel()
		err := Validate("invalid", "email")
		if err == nil {
			t.Error("expected error for invalid email")
		}
	})

	t.Run("valid url", func(t *testing.T) {
		t.Parallel()
		err := Validate("https://example.com", "url")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid url", func(t *testing.T) {
		t.Parallel()
		err := Validate("not-a-url", "url")
		if err == nil {
			t.Error("expected error for invalid url")
		}
	})

	t.Run("valid ip", func(t *testing.T) {
		t.Parallel()
		err := Validate("192.168.1.1", "ip")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid ip", func(t *testing.T) {
		t.Parallel()
		err := Validate("999.999.999.999", "ip")
		if err == nil {
			t.Error("expected error for invalid ip")
		}
	})

	t.Run("valid regexp", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello123", "regexp=^[a-z]+\\d+$")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid regexp", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "regexp=^\\d+$")
		if err == nil {
			t.Error("expected error for regexp mismatch")
		}
	})

	t.Run("valid oneof", func(t *testing.T) {
		t.Parallel()
		err := Validate("admin", "oneof=admin user guest")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid oneof", func(t *testing.T) {
		t.Parallel()
		err := Validate("superadmin", "oneof=admin user guest")
		if err == nil {
			t.Error("expected error for invalid oneof")
		}
	})

	t.Run("valid gt", func(t *testing.T) {
		t.Parallel()
		err := Validate(42, "gt=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid gt", func(t *testing.T) {
		t.Parallel()
		err := Validate(5, "gt=10")
		if err == nil {
			t.Error("expected error for value not greater than 10")
		}
	})

	t.Run("valid gte", func(t *testing.T) {
		t.Parallel()
		err := Validate(10, "gte=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid lt", func(t *testing.T) {
		t.Parallel()
		err := Validate(5, "lt=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid lt", func(t *testing.T) {
		t.Parallel()
		err := Validate(10, "lt=10")
		if err == nil {
			t.Error("expected error for value not less than 10")
		}
	})

	t.Run("valid lte", func(t *testing.T) {
		t.Parallel()
		err := Validate(10, "lte=10")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("valid len", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "len=5")
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid len", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "len=3")
		if err == nil {
			t.Error("expected error for wrong length")
		}
	})

	t.Run("required nil", func(t *testing.T) {
		t.Parallel()
		err := Validate(nil, "required")
		if err == nil {
			t.Error("expected error for nil value")
		}
	})

	t.Run("empty rules", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "")
		if err != nil {
			t.Errorf("expected no error for empty rules, got %v", err)
		}
	})

	t.Run("unknown rule", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "unknown")
		if err != nil {
			t.Errorf("expected no error for unknown rule, got %v", err)
		}
	})
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

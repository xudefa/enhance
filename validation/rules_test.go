package validation

import (
	"reflect"
	"testing"
)

func TestIsRequiredValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"non-empty string", "hello", true},
		{"empty string", "", false},
		{"non-zero int", 42, true},
		{"zero int", 0, false},
		{"true bool", true, true},
		{"false bool", false, false},
		{"non-nil ptr", new(int), true},
		{"nil ptr", (*int)(nil), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			rv := reflect.ValueOf(tt.value)
			got := validator.isRequiredValid(rv)
			if got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}

func TestIsMinValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if !validator.isMinValid(rv, "3") {
			t.Error("expected min valid for string length >= 3")
		}
		if validator.isMinValid(rv, "10") {
			t.Error("expected min invalid for string length < 10")
		}
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if !validator.isMinValid(rv, "10") {
			t.Error("expected min valid for int >= 10")
		}
		if validator.isMinValid(rv, "50") {
			t.Error("expected min invalid for int < 50")
		}
	})

	t.Run("invalid min value", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if validator.isMinValid(rv, "abc") {
			t.Error("expected min invalid for non-numeric value")
		}
	})
}

func TestIsMaxValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hi")
		if !validator.isMaxValid(rv, "5") {
			t.Error("expected max valid for string length <= 5")
		}
		if validator.isMaxValid(rv, "1") {
			t.Error("expected max invalid for string length > 1")
		}
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if !validator.isMaxValid(rv, "50") {
			t.Error("expected max valid for int <= 50")
		}
		if validator.isMaxValid(rv, "10") {
			t.Error("expected max invalid for int > 10")
		}
	})
}

func TestIsLenValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if !validator.isLenValid(rv, "5") {
			t.Error("expected len valid for string length == 5")
		}
		if validator.isLenValid(rv, "3") {
			t.Error("expected len invalid for string length != 3")
		}
	})

	t.Run("slice length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf([]int{1, 2, 3})
		if !validator.isLenValid(rv, "3") {
			t.Error("expected len valid for slice length == 3")
		}
	})
}

func TestIsEmailValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	tests := []struct {
		email    string
		expected bool
	}{
		{"user@example.com", true},
		{"invalid-email", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			t.Parallel()
			rv := reflect.ValueOf(tt.email)
			got := validator.isEmailValid(rv)
			if got != tt.expected {
				t.Errorf("email %s: expected %v, got %v", tt.email, tt.expected, got)
			}
		})
	}

	t.Run("non-string", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if validator.isEmailValid(rv) {
			t.Error("expected email invalid for non-string")
		}
	})
}

func TestIsRegexpValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("valid pattern", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello123")
		if !validator.isRegexpValid(rv, `^[a-z]+\d+$`) {
			t.Error("expected regexp valid")
		}
	})

	t.Run("invalid pattern", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if validator.isRegexpValid(rv, `^\d+$`) {
			t.Error("expected regexp invalid")
		}
	})

	t.Run("invalid regex", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if validator.isRegexpValid(rv, `[invalid`) {
			t.Error("expected regexp invalid for bad pattern")
		}
	})
}

func TestIsGtValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if !validator.isGtValid(rv, "3") {
			t.Error("expected gt valid for string length > 3")
		}
		if validator.isGtValid(rv, "5") {
			t.Error("expected gt invalid for string length == 5")
		}
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if !validator.isGtValid(rv, "10") {
			t.Error("expected gt valid for int > 10")
		}
		if validator.isGtValid(rv, "50") {
			t.Error("expected gt invalid for int < 50")
		}
	})
}

func TestIsGteValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hello")
		if !validator.isGteValid(rv, "5") {
			t.Error("expected gte valid for string length >= 5")
		}
	})

	t.Run("int value", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(42)
		if !validator.isGteValid(rv, "42") {
			t.Error("expected gte valid for int == 42")
		}
	})
}

func TestIsLtValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hi")
		if !validator.isLtValid(rv, "5") {
			t.Error("expected lt valid for string length < 5")
		}
		if validator.isLtValid(rv, "2") {
			t.Error("expected lt invalid for string length == 2")
		}
	})
}

func TestIsLteValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string length", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("hi")
		if !validator.isLteValid(rv, "2") {
			t.Error("expected lte valid for string length <= 2")
		}
	})
}

func TestIsURLValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"not-a-url", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			t.Parallel()
			rv := reflect.ValueOf(tt.url)
			got := validator.isURLValid(rv)
			if got != tt.expected {
				t.Errorf("url %s: expected %v, got %v", tt.url, tt.expected, got)
			}
		})
	}
}

func TestIsIPValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	tests := []struct {
		ip       string
		expected bool
	}{
		{"192.168.1.1", true},
		{"255.255.255.255", true},
		{"999.999.999.999", false},
		{"not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			t.Parallel()
			rv := reflect.ValueOf(tt.ip)
			got := validator.isIPValid(rv)
			if got != tt.expected {
				t.Errorf("ip %s: expected %v, got %v", tt.ip, tt.expected, got)
			}
		})
	}
}

func TestIsOneOfValid(t *testing.T) {
	t.Parallel()
	validator := &TagValidator{}

	t.Run("string options", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf("admin")
		if !validator.isOneOfValid(rv, "admin user guest") {
			t.Error("expected oneof valid")
		}
		if validator.isOneOfValid(rv, "user guest") {
			t.Error("expected oneof invalid")
		}
	})

	t.Run("int options", func(t *testing.T) {
		t.Parallel()
		rv := reflect.ValueOf(1)
		if !validator.isOneOfValid(rv, "1 2 3") {
			t.Error("expected oneof valid")
		}
	})
}

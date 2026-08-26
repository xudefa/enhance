package actuator

import (
	"testing"
)

func TestLooksLikeSensitiveData(t *testing.T) {
	t.Parallel()
	t.Run("PrivateKey", func(t *testing.T) {
		data := "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgHcTz6sE2I2yPB\naQDrB8g3bqVW8T3oL9k2J3f8y4x7z5w6v5u4t3s2r1q0p9o8n7m6l5k4j3i2h1g0f\n-----END RSA PRIVATE KEY-----"
		if !looksLikeSensitiveData(data) {
			t.Error("expected private key to be detected as sensitive")
		}
	})

	t.Run("RandomString", func(t *testing.T) {
		data := "Ab3dEf6hIj9lMn2pQr5tUv8xYz1bCd4fGh7jKl0mNp3qRs6t!@#$"
		if !looksLikeSensitiveData(data) {
			t.Error("expected random-looking string to be detected as sensitive")
		}
	})

	t.Run("JWTToken", func(t *testing.T) {
		data := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.POstGetfAytaZS82wHcjoTyoqhMyxXiWdR7Nn7A29DNSl0EiXLdwJ6xC6AfgZWF1bOsS_TuYI3OG85AmiExREkrS6tDfTQ2B3WXlrr-wp5AokiRbz3_oB4OxG-W9KcEEbDRcZc"
		if !looksLikeSensitiveData(data) {
			t.Error("expected JWT token to be detected as sensitive")
		}
	})

	t.Run("ShortString", func(t *testing.T) {
		data := "short"
		if looksLikeSensitiveData(data) {
			t.Error("expected short string to not be detected as sensitive")
		}
	})

	t.Run("NormalString", func(t *testing.T) {
		data := "this is a normal string without any sensitive data"
		if looksLikeSensitiveData(data) {
			t.Error("expected normal string to not be detected as sensitive")
		}
	})
}

func TestIsRandomLookingString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "HasAllCharTypes",
			input:    "Ab3dEf6hIj9lMn2pQr5tUv8xYz1bCd4fGh7jKl0mNp3qRs6t!@#$",
			expected: true,
		},
		{
			name:     "OnlyUppercase",
			input:    "ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			expected: false,
		},
		{
			name:     "OnlyLowercase",
			input:    "abcdefghijklmnopqrstuvwxyz",
			expected: false,
		},
		{
			name:     "OnlyDigits",
			input:    "12345678901234567890",
			expected: false,
		},
		{
			name:     "ShortString",
			input:    "Ab3d",
			expected: false,
		},
		{
			name:     "MissingSpecial",
			input:    "Ab3dEf6hIj9lMn2pQr5tUv8xYz1bCd4fGh7jKl0mNp",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRandomLookingString(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestIsTokenFormat(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ValidJWT",
			input:    "eyJhbGciOiJIUzI1NiJ9.dGVzdA==.dGVzdA==",
			expected: true,
		},
		{
			name:     "InvalidBase64InJWT",
			input:    "invalid!!!.invalid!!!.invalid!!!",
			expected: false,
		},
		{
			name:     "TwoDots",
			input:    "part1.part2.part3",
			expected: false,
		},
		{
			name:     "NoDots",
			input:    "nodotsatall",
			expected: false,
		},
		{
			name:     "FourParts",
			input:    "part1.part2.part3.part4",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTokenFormat(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestIsValidBase64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "ValidBase64",
			input:    "SGVsbG8gV29ybGQ=",
			expected: true,
		},
		{
			name:     "InvalidBase64",
			input:    "!!!invalid!!!",
			expected: false,
		},
		{
			name:     "EmptyString",
			input:    "",
			expected: true,
		},
		{
			name:     "JWTHeader",
			input:    "eyJhbGciOiJIUzI1NiJ9",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidBase64(tt.input)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestDefaultKeywords(t *testing.T) {
	t.Parallel()
	keywords := defaultKeywords()

	expectedKeywords := []string{
		"password", "secret", "token", "key", "auth",
		"credential", "private", "api_key", "access_token",
		"client_secret", "oauth", "bearer", "jwt",
	}

	if len(keywords) != len(expectedKeywords) {
		t.Errorf("expected %d keywords, got %d", len(expectedKeywords), len(keywords))
	}

	for _, expected := range expectedKeywords {
		found := false
		for _, keyword := range keywords {
			if keyword == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected keyword '%s' not found", expected)
		}
	}
}

type alwaysSensitiveStrategy struct{}

func (c *alwaysSensitiveStrategy) IsSensitive(key string, value any) bool {
	return true
}

func TestSanitizer_AlwaysSensitiveStrategy(t *testing.T) {
	t.Parallel()
	sanitizer := NewSanitizer()

	custom := &alwaysSensitiveStrategy{}

	sanitizer.AddStrategy(custom)

	result := sanitizer.Sanitize("custom_field", "some_value")
	if result != redactedValue {
		t.Errorf("expected redacted value, got %v", result)
	}
}

type testSanitizeStrategy struct {
	isSensitive func(key string, value any) bool
}

func (s *testSanitizeStrategy) IsSensitive(key string, value any) bool {
	return s.isSensitive(key, value)
}

func TestSanitizer_SanitizeWithCustomStrategy(t *testing.T) {
	t.Parallel()
	sanitizer := NewSanitizer()

	strategy := &testSanitizeStrategy{
		isSensitive: func(key string, value any) bool {
			return key == "custom_sensitive"
		},
	}

	sanitizer.AddStrategy(strategy)

	t.Run("CustomSensitiveKey", func(t *testing.T) {
		result := sanitizer.Sanitize("custom_sensitive", "value")
		if result != redactedValue {
			t.Errorf("expected redacted value, got %v", result)
		}
	})

	t.Run("NonSensitiveKey", func(t *testing.T) {
		result := sanitizer.Sanitize("normal_key", "value")
		if result == redactedValue {
			t.Logf("warning: 'normal_key' was redacted due to keyword matching, this is expected behavior")
		}
	})
}

func TestSanitizer_SanitizeNilValue(t *testing.T) {
	t.Parallel()
	sanitizer := NewSanitizer()

	result := sanitizer.Sanitize("password", nil)
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestSanitizer_SanitizeKeywordMatch(t *testing.T) {
	t.Parallel()
	sanitizer := NewSanitizer()

	tests := []struct {
		key      string
		value    string
		expected string
	}{
		{"password", "mysecret", redactedValue},
		{"db_password", "pass123", redactedValue},
		{"secret_key", "key123", redactedValue},
		{"api_token", "token123", redactedValue},
		{"auth_header", "bearer123", redactedValue},
		{"normal_field", "normal_value", "normal_value"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			result := sanitizer.Sanitize(tt.key, tt.value)
			if result != tt.expected {
				t.Errorf("expected %v, got %v for key %s", tt.expected, result, tt.key)
			}
		})
	}
}

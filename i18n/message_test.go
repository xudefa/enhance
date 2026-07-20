package i18n

import "testing"

func TestLocale_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		locale   Locale
		expected string
	}{
		{"with country", Locale{Language: "zh", Country: "CN"}, "zh_CN"},
		{"without country", Locale{Language: "en"}, "en"},
		{"with variant", Locale{Language: "en", Country: "US", Variant: "POSIX"}, "en_US"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.locale.String(); got != tt.expected {
				t.Errorf("Locale.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestStaticMessageSource_GetMessage_Found(t *testing.T) {
	t.Parallel()
	src := NewStaticMessageSource()
	src.AddMessage("greeting", "Hello, %s!")

	got := src.GetMessage("greeting", "World")
	if got != "Hello, World!" {
		t.Errorf("GetMessage() = %v, want %v", got, "Hello, World!")
	}
}

func TestStaticMessageSource_GetMessage_NotFound(t *testing.T) {
	t.Parallel()
	src := NewStaticMessageSource()

	got := src.GetMessage("unknown")
	if got != "unknown" {
		t.Errorf("GetMessage() should return code when not found, got %v", got)
	}
}

func TestStaticMessageSource_AddMessages(t *testing.T) {
	t.Parallel()
	src := NewStaticMessageSource()
	src.AddMessages(map[string]string{
		"hello": "Hello",
		"bye":   "Goodbye",
	})

	if src.GetMessage("hello") != "Hello" {
		t.Error("AddMessages failed for 'hello'")
	}
	if src.GetMessage("bye") != "Goodbye" {
		t.Error("AddMessages failed for 'bye'")
	}
}

func TestStaticMessageSource_GetMessageWithLocale(t *testing.T) {
	t.Parallel()
	src := NewStaticMessageSource()
	src.AddMessage("msg", "test %s")

	got := src.GetMessageWithLocale("msg", Locale{Language: "zh"}, "value")
	if got != "test value" {
		t.Errorf("GetMessageWithLocale() = %v, want %v", got, "test value")
	}
}

func TestResourceBundleMessageSource_ExactMatch(t *testing.T) {
	t.Parallel()
	src := NewResourceBundleMessageSource()
	src.AddResourceBundle(Locale{Language: "zh", Country: "CN"}, map[string]string{
		"greeting": "你好",
	})

	got := src.GetMessageWithLocale("greeting", Locale{Language: "zh", Country: "CN"})
	if got != "你好" {
		t.Errorf("expected '你好', got %v", got)
	}
}

func TestResourceBundleMessageSource_FallbackToLanguage(t *testing.T) {
	t.Parallel()
	src := NewResourceBundleMessageSource()
	src.AddResourceBundle(Locale{Language: "zh"}, map[string]string{
		"greeting": "你好",
	})

	got := src.GetMessageWithLocale("greeting", Locale{Language: "zh", Country: "CN"})
	if got != "你好" {
		t.Errorf("expected '你好', got %v", got)
	}
}

func TestResourceBundleMessageSource_FallbackToParent(t *testing.T) {
	t.Parallel()
	parent := NewStaticMessageSource()
	parent.AddMessage("msg", "from parent")

	src := NewResourceBundleMessageSource()
	src.SetFallback(parent)

	got := src.GetMessage("msg")
	if got != "from parent" {
		t.Errorf("expected 'from parent', got %v", got)
	}
}

func TestResourceBundleMessageSource_ReturnsCode(t *testing.T) {
	t.Parallel()
	src := NewResourceBundleMessageSource()

	got := src.GetMessage("unknown.code")
	if got != "unknown.code" {
		t.Errorf("expected code itself, got %v", got)
	}
}

func TestResourceBundleMessageSource_FormatMessage(t *testing.T) {
	t.Parallel()
	src := NewResourceBundleMessageSource()
	src.AddResourceBundle(Locale{Language: "en"}, map[string]string{
		"welcome": "Welcome, %s! You have %d messages.",
	})

	got := src.GetMessage("welcome", "Alice", 5)
	want := "Welcome, Alice! You have 5 messages."
	if got != want {
		t.Errorf("GetMessage() = %v, want %v", got, want)
	}
}

func TestDelegatingMessageSource_FindsInChild(t *testing.T) {
	t.Parallel()
	src := NewDelegatingMessageSource()
	child := NewStaticMessageSource()
	child.AddMessage("msg", "from child")
	src.AddChild(child)

	got := src.GetMessage("msg")
	if got != "from child" {
		t.Errorf("expected 'from child', got %v", got)
	}
}

func TestDelegatingMessageSource_FallbackToParent(t *testing.T) {
	t.Parallel()
	src := NewDelegatingMessageSource()
	parent := NewStaticMessageSource()
	parent.AddMessage("msg", "from parent")
	src.SetParent(parent)

	got := src.GetMessage("msg")
	if got != "from parent" {
		t.Errorf("expected 'from parent', got %v", got)
	}
}

func TestDelegatingMessageSource_ReturnsCode(t *testing.T) {
	t.Parallel()
	src := NewDelegatingMessageSource()

	got := src.GetMessage("unknown")
	if got != "unknown" {
		t.Errorf("expected code itself, got %v", got)
	}
}

func TestResourceBundleMessageSource_GetMessage_DefaultLocale(t *testing.T) {
	t.Parallel()
	src := NewResourceBundleMessageSource()
	src.AddResourceBundle(Locale{Language: "en"}, map[string]string{
		"hello": "Hello",
	})

	got := src.GetMessage("hello")
	if got != "Hello" {
		t.Errorf("expected 'Hello', got %v", got)
	}
}

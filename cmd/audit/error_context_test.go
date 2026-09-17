package main

import (
	"strings"
	"testing"
)

func TestAnalyzeErrorContext_BareReturn(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() error {
	err := doSomething()
	if err != nil {
		return err
	}
	return nil
}`, false)
	findings := analyzeErrorContext(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条裸返回，得到 %v", findings)
	}
	if findings[0].Category != CategoryErrorContext {
		t.Errorf("类别错误: %v", findings[0].Category)
	}
	if !strings.Contains(findings[0].Message, "err") {
		t.Errorf("消息应包含 err: %s", findings[0].Message)
	}
}

func TestAnalyzeErrorContext_MissingPercentW(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "fmt"
func F() error {
	err := doSomething()
	return fmt.Errorf("load failed: %v", err)
}`, false)
	findings := analyzeErrorContext(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条 %%v 误用，得到 %+v", findings)
	}
	if !strings.Contains(findings[0].Suggestion, "%w") {
		t.Errorf("建议应提到 %%w: %s", findings[0].Suggestion)
	}
}

func TestAnalyzeErrorContext_HasPercentW(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "fmt"
func F() error {
	err := doSomething()
	return fmt.Errorf("load failed: %w", err)
}`, false)
	if findings := analyzeErrorContext(f); len(findings) != 0 {
		t.Fatalf("使用 %%w 不应报告: %+v", findings)
	}
}

func TestAnalyzeErrorContext_SkipsNonErrorArg(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "fmt"
func F(n int) error {
	return fmt.Errorf("write %d bytes", n)
}`, false)
	if findings := analyzeErrorContext(f); len(findings) != 0 {
		t.Fatalf("非 error 参数不应报告: %+v", findings)
	}
}

func TestIsErrorName_GoConvention(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		want bool
	}{
		{"err", true},
		{"error", true},
		{"e", true},
		{"errNotFound", true},
		{"errConfig", true},
		{"errs", false},
		{"errmsg", false},
		{"errlist", false},
		{"value", false},
		{"", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isErrorName(tt.name); got != tt.want {
				t.Errorf("isErrorName(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func TestAnalyzeErrorContext_SkipsPluralErr(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "fmt"
func F() error {
	errs := []string{"a"}
	return fmt.Errorf("required config missing: %v", errs)
}`, false)
	if findings := analyzeErrorContext(f); len(findings) != 0 {
		t.Fatalf("errs 非 error 值，不应报告: %+v", findings)
	}
}

func TestAnalyzeErrorContext_SkipsTests(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "testing"
func TestX(t *testing.T) {
	err := doSomething()
	if err != nil {
		t.Fatal(err)
	}
}`, true)
	if findings := analyzeErrorContext(f); len(findings) != 0 {
		t.Fatalf("测试文件不应报告: %+v", findings)
	}
}

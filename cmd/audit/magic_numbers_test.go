package main

import (
	"strings"
	"testing"
)

func TestAnalyzeMagicNumbers_Condition(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(x int) bool {
	return x > 60
}`, false)
	findings := analyzeMagicNumbers(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条魔法数字，得到 %+v", findings)
	}
	if findings[0].Category != CategoryMagicNumbers {
		t.Errorf("类别错误: %v", findings[0].Category)
	}
	if !strings.Contains(findings[0].Message, "60") {
		t.Errorf("消息应包含 60: %s", findings[0].Message)
	}
}

func TestAnalyzeMagicNumbers_SwitchAndReturn(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(x int) int {
	switch x {
	case 2000:
		return 4000
	}
	return 0
}`, false)
	findings := analyzeMagicNumbers(f)
	if len(findings) != 2 {
		t.Fatalf("期望 2 条（switch 2000 + 返回 4000），得到 %+v", findings)
	}
}

func TestAnalyzeMagicNumbers_SkipsTime(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "time"
func F() time.Duration {
	return 30 * time.Second
}`, false)
	if findings := analyzeMagicNumbers(f); len(findings) != 0 {
		t.Fatalf("与 time 常量运算的数字不应报告: %+v", findings)
	}
}

func TestAnalyzeMagicNumbers_SkipsPortsAndPowers(t *testing.T) {
	t.Parallel()
	for _, code := range []string{
		"package p\nfunc F() int { return 8080 }",
		"package p\nfunc F(x int) int { if x&1024 == 0 { return 1 }; return 0 }",
	} {
		f := parseSnippet(t, code, false)
		if findings := analyzeMagicNumbers(f); len(findings) != 0 {
			t.Fatalf("端口/2的幂不应报告 %q: %+v", code, findings)
		}
	}
}

func TestAnalyzeMagicNumbers_AllowsHTTPStatusAndSmall(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(x int) int {
	if x == 0 {
		return 0
	}
	if x == 404 {
		return 1
	}
	return -1
}`, false)
	if findings := analyzeMagicNumbers(f); len(findings) != 0 {
		t.Fatalf("0/1/-1/404 不应报告: %+v", findings)
	}
}

func TestAnalyzeMagicNumbers_NegativeValue(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(x int) int {
	if x < -30 {
		return -30
	}
	return 0
}`, false)
	findings := analyzeMagicNumbers(f)
	if len(findings) != 2 {
		t.Fatalf("期望 -30 出现 2 处被报告，得到 %+v", findings)
	}
}

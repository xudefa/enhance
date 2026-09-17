package main

import (
	"strings"
	"testing"
)

func TestAnalyzeDocumentation_MissingFuncDoc(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() int { return 0 }
func unexported() int { return 0 }`, false)
	findings := analyzeDocumentation(f)
	if len(findings) != 1 {
		t.Fatalf("期望仅 F 缺注释（内部函数不检测），得到 %+v", findings)
	}
	if !strings.Contains(findings[0].Message, "F") {
		t.Errorf("消息应包含函数名 F: %s", findings[0].Message)
	}
}

func TestAnalyzeDocumentation_MissingTypeVarDoc(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
type Foo struct{ X int }
var Bar = 1`, false)
	findings := analyzeDocumentation(f)
	if len(findings) != 2 {
		t.Fatalf("期望 Foo 与 Bar 均缺注释，得到 %+v", findings)
	}
}

func TestAnalyzeDocumentation_HasDoc(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p

// F 返回 0。
func F() int { return 0 }

// Foo 是示例类型。
type Foo struct{ X int }`, false)
	if findings := analyzeDocumentation(f); len(findings) != 0 {
		t.Fatalf("有注释不应报告: %+v", findings)
	}
}

func TestAnalyzeDocumentation_EnglishComment(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p

// Register registers a bean factory into the container.
func Register() {}`, false)
	findings := analyzeDocumentation(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条英文注释发现，得到 %+v", findings)
	}
	if !strings.Contains(findings[0].Message, "英文") {
		t.Errorf("消息应提示中文注释规范: %s", findings[0].Message)
	}
}

func TestAnalyzeDocumentation_SkipsTestFuncs(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "testing"
func TestX(t *testing.T) {}`, true)
	if findings := analyzeDocumentation(f); len(findings) != 0 {
		t.Fatalf("测试函数不应报告: %+v", findings)
	}
}

func TestAnalyzeDocumentation_GroupDoc(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p

// 常量定义。
const (
	A = 1
	B = 2
)`, false)
	if findings := analyzeDocumentation(f); len(findings) != 0 {
		t.Fatalf("分组注释覆盖的常量不应报告: %+v", findings)
	}
}

package main

import (
	"strings"
	"testing"
)

func TestAnalyzeBadNames_SingleLetter(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() int {
	n := 42
	n = n + 1
	return n
}`, false)
	findings := analyzeBadNames(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条发现，得到 %+v", findings)
	}
	if findings[0].Category != CategoryBadNames {
		t.Errorf("类别错误: %v", findings[0].Category)
	}
	if !strings.Contains(findings[0].Message, "n") {
		t.Errorf("消息应包含变量名 n: %s", findings[0].Message)
	}
}

func TestAnalyzeBadNames_LoopVarExcluded(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() int {
	total := 0
	for i := 0; i < 10; i++ {
		total += i
	}
	return total
}`, false)
	findings := analyzeBadNames(f)
	for _, fd := range findings {
		if strings.Contains(fd.Message, `"i"`) {
			t.Fatalf("循环变量 i 不应被报告: %+v", findings)
		}
	}
}

func TestAnalyzeBadNames_AllowsCommon(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() int {
	c := 0 // container 惯例缩写，允许
	c = c + 1
	return c
}`, false)
	if findings := analyzeBadNames(f); len(findings) != 0 {
		t.Fatalf("c 与 err 不应被报告: %+v", findings)
	}
}

func TestAnalyzeBadNames_VagueName(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F(v int) int {
	tmp := v * 2
	return tmp
}`, false)
	findings := analyzeBadNames(f)
	if len(findings) != 1 {
		t.Fatalf("期望 1 条模糊命名发现，得到 %+v", findings)
	}
	if !strings.Contains(findings[0].Message, "tmp") {
		t.Errorf("消息应包含 tmp: %s", findings[0].Message)
	}
}

func TestAnalyzeBadNames_TestParamsAllowed(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
import "testing"
func TestX(t *testing.T) {
	t.Log("a")
	t.Log("b")
}`, true)
	if findings := analyzeBadNames(f); len(findings) != 0 {
		t.Fatalf("测试参数 t/b 不应被报告: %+v", findings)
	}
}

func TestAnalyzeBadNames_BlankAllowed(t *testing.T) {
	t.Parallel()
	f := parseSnippet(t, `package p
func F() error {
	_, err := something()
	if err != nil {
		return err
	}
	return nil
}`, false)
	for _, fd := range analyzeBadNames(f) {
		if strings.Contains(fd.Message, `"_"`) {
			t.Fatalf("空标识符不应被报告")
		}
	}
}

package aop

import (
	"reflect"
	"testing"
)

func TestMatchAll(t *testing.T) {
	t.Parallel()
	pc := MatchAll()

	// MatchAll 应该匹配任何方法
	if !pc.Matches(nil, "AnyMethod") {
		t.Error("MatchAll should match any method")
	}

	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("MatchAll should match any method with target")
	}
}

func TestMatchByName(t *testing.T) {
	t.Parallel()
	pc := MatchByName("DoSomething")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName should match DoSomething")
	}

	if pc.Matches(nil, "DoAnother") {
		t.Error("MatchByName should not match DoAnother")
	}
}

func TestMatchByNamePrefix(t *testing.T) {
	t.Parallel()
	pc := MatchByName("Do*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName with Do* should match methods with Do prefix")
	}

	if pc.Matches(nil, "GetValue") {
		t.Error("MatchByName with Do* should not match methods without Do prefix")
	}
}

func TestMatchByRegex(t *testing.T) {
	t.Parallel()
	pc := MatchByName("^Do.*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("MatchByName with regex should match methods matching regex")
	}

	if pc.Matches(nil, "GetValue") {
		t.Error("MatchByName with regex should not match methods not matching regex")
	}
}

func TestMatchInterface_NilInput(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(nil)

	// 当 target 为 nil 时，应该返回 false
	if pc.Matches(nil, "DoSomething") {
		t.Error("MatchInterface(nil) should not match nil target")
	}
}

func TestMatchInterface_NonInterfaceInput(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(reflect.TypeFor[string]())

	// string 不是接口，不应该匹配
	if pc.Matches("test", "DoSomething") {
		t.Error("MatchInterface with non-interface should not match")
	}
}

func TestMatchInterface(t *testing.T) {
	t.Parallel()
	pc := MatchInterface(reflect.TypeFor[TestInterfaceForMatch]())

	if !pc.Matches(&TestImplForMatch{}, "DoSomething") {
		t.Error("MatchInterface should match implementing struct pointer")
	}

	if pc.Matches("string", "DoSomething") {
		t.Error("MatchInterface should not match non-implementing type")
	}
}

type TestInterfaceForMatch interface {
	DoSomething()
}

type TestImplForMatch struct{}

func (t *TestImplForMatch) DoSomething() {}

func TestPointCut_Expression(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pointCut PointCut
		expected string
	}{
		{"MatchByName", MatchByName("DoSomething"), "DoSomething"},
		{"MatchAll", MatchAll(), "*"},
		{"MatchByRegex", MatchByName("^Do.*"), "^Do.*"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.pointCut.Expression() != tt.expected {
				t.Errorf("Expression() = %v, want %v", tt.pointCut.Expression(), tt.expected)
			}
		})
	}
}

func TestMatchByPackage(t *testing.T) {
	t.Parallel()
	pc := MatchByPackage("github.com/xudefa/enhance")

	// 使用实际的对象测试包匹配
	userService := &TestUserService{}
	if !pc.Matches(userService, "DoSomething") {
		t.Error("MatchByPackage should match types in the specified package")
	}

	// nil target 应该返回 false
	if pc.Matches(nil, "DoSomething") {
		t.Error("MatchByPackage should not match nil target")
	}
}

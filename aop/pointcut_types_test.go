package aop

import (
	"reflect"
	"strings"
	"testing"
)

func TestPointCutImpl_MatchesNilTarget_Variants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pc        *pointCutImpl
		method    string
		wantMatch bool
	}{
		{
			name:      "nil class and nil method matcher matches anything",
			pc:        &pointCutImpl{},
			method:    "AnyMethod",
			wantMatch: true,
		},
		{
			name:      "nil class but with method matcher",
			pc:        &pointCutImpl{methodMatcher: func(m reflect.Method) bool { return m.Name == "Do" }},
			method:    "Do",
			wantMatch: true,
		},
		{
			name:      "nil class but method matcher fails",
			pc:        &pointCutImpl{methodMatcher: func(m reflect.Method) bool { return m.Name == "Do" }},
			method:    "Skip",
			wantMatch: false,
		},
		{
			name:      "class matcher present with nil target returns false",
			pc:        &pointCutImpl{classMatcher: func(t reflect.Type) bool { return true }},
			method:    "Do",
			wantMatch: false,
		},
		{
			name:      "interface matcher present with nil target returns false",
			pc:        &pointCutImpl{interfaceType: reflect.TypeFor[TestInterfaceForMatch]()},
			method:    "DoSomething",
			wantMatch: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pc.Matches(nil, tt.method)
			if got != tt.wantMatch {
				t.Errorf("Matches(nil, %q) = %v, want %v", tt.method, got, tt.wantMatch)
			}
		})
	}
}

func TestPointCutImpl_MatchesWithTarget_Variants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pc        *pointCutImpl
		target    any
		method    string
		wantMatch bool
	}{
		{
			name: "class matcher matches",
			pc: &pointCutImpl{
				classMatcher: func(rt reflect.Type) bool { return rt.Name() == "TestUserService" },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: true,
		},
		{
			name: "class matcher rejects",
			pc: &pointCutImpl{
				classMatcher: func(rt reflect.Type) bool { return rt.Name() == "Other" },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: false,
		},
		{
			name:   "interface matcher matches pointer",
			pc:     &pointCutImpl{interfaceType: reflect.TypeFor[TestInterfaceForMatch]()},
			target: &TestImplForMatch{},
			method: "DoSomething",
			wantMatch: true,
		},
		{
			name:   "interface matcher rejects non-implementer",
			pc:     &pointCutImpl{interfaceType: reflect.TypeFor[TestInterfaceForMatch]()},
			target: "string value",
			method: "DoSomething",
			wantMatch: false,
		},
		{
			name:   "nil method matcher matches all methods",
			pc:     &pointCutImpl{classMatcher: func(rt reflect.Type) bool { return true }},
			target: &TestUserService{},
			method: "DoSomething",
			wantMatch: true,
		},
		{
			name: "method matcher matches via value receiver",
			pc: &pointCutImpl{
				methodMatcher: func(m reflect.Method) bool { return m.Name == "DoSomething" },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: true,
		},
		{
			name: "method not found on target",
			pc: &pointCutImpl{
				methodMatcher: func(m reflect.Method) bool { return true },
			},
			target:    &TestUserService{},
			method:    "NonExistent",
			wantMatch: false,
		},
		{
			name: "pointer type target dereferenced correctly",
			pc: &pointCutImpl{
				classMatcher: func(rt reflect.Type) bool { return rt.Name() == "TestUserService" },
			},
			target:    (*TestUserService)(nil),
			method:    "DoSomething",
			wantMatch: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pc.Matches(tt.target, tt.method)
			if got != tt.wantMatch {
				t.Errorf("Matches(%v, %q) = %v, want %v", tt.target, tt.method, got, tt.wantMatch)
			}
		})
	}
}

func TestPointCutImpl_MatchClass_Variants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pc        *pointCutImpl
		inputType reflect.Type
		want      bool
	}{
		{
			name:      "no matchers always matches",
			pc:        &pointCutImpl{},
			inputType: reflect.TypeOf(""),
			want:      true,
		},
		{
			name:      "class matcher rejects pointer by dereferencing",
			pc:        &pointCutImpl{classMatcher: func(rt reflect.Type) bool { return rt.Name() == "string" }},
			inputType: reflect.PointerTo(reflect.TypeOf("")),
			want:      true,
		},
		{
			name:      "interface type matches",
			pc:        &pointCutImpl{interfaceType: reflect.TypeFor[TestInterfaceForMatch]()},
			inputType: reflect.TypeOf((*TestInterfaceForMatch)(nil)).Elem(),
			want:      true,
		},
		{
			name:      "interface type rejects non-implementer",
			pc:        &pointCutImpl{interfaceType: reflect.TypeFor[TestInterfaceForMatch]()},
			inputType: reflect.TypeOf(""),
			want:      false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pc.MatchClass(tt.inputType)
			if got != tt.want {
				t.Errorf("MatchClass(%v) = %v, want %v", tt.inputType, got, tt.want)
			}
		})
	}
}

func TestCompositePointCut_MatchesAdvanced(t *testing.T) {
	t.Parallel()
	truePC := PointCutFunc(func(m reflect.Method) bool { return true })
	falsePC := PointCutFunc(func(m reflect.Method) bool { return false })

	t.Run("AND all match", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{truePC, truePC}, and: true}
		if !c.Matches(nil, "Any") {
			t.Error("expected AND(true, true) to match")
		}
	})
	t.Run("AND one fails", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{truePC, falsePC}, and: true}
		if c.Matches(nil, "Any") {
			t.Error("expected AND(true, false) to not match")
		}
	})
	t.Run("OR one matches", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{falsePC, truePC}, and: false}
		if !c.Matches(nil, "Any") {
			t.Error("expected OR(false, true) to match")
		}
	})
	t.Run("OR none match", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{falsePC, falsePC}, and: false}
		if c.Matches(nil, "Any") {
			t.Error("expected OR(false, false) to not match")
		}
	})
	t.Run("AND empty matches all", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{}, and: true}
		if !c.Matches(nil, "Any") {
			t.Error("expected AND() empty to match")
		}
	})
	t.Run("OR empty matches none", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{}, and: false}
		if c.Matches(nil, "Any") {
			t.Error("expected OR() empty to not match")
		}
	})
}

func TestCompositePointCut_MatchClassAdvanced(t *testing.T) {
	t.Parallel()
	matchAll := &pointCutImpl{}
	matchNone := PointCutFunc(func(m reflect.Method) bool { return false })

	t.Run("AND all match", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{matchAll, matchAll}, and: true}
		if !c.MatchClass(reflect.TypeOf("")) {
			t.Error("expected AND class match")
		}
	})
	t.Run("OR one matches", func(t *testing.T) {
		t.Parallel()
		c := &compositePointCut{pointcuts: []PointCut{matchNone, matchAll}, and: false}
		if !c.MatchClass(reflect.TypeOf("")) {
			t.Error("expected OR class match")
		}
	})
}

func TestPointCutFunc_MatchesNilTarget_Variants(t *testing.T) {
	t.Parallel()
	pc := PointCutFunc(func(m reflect.Method) bool {
		return m.Name == "DoSomething"
	})
	if !pc.Matches(nil, "DoSomething") {
		t.Error("PointCutFunc should match dummy method with matching name")
	}
	if pc.Matches(nil, "Other") {
		t.Error("PointCutFunc should not match dummy method with wrong name")
	}
}

func TestPointCutFunc_MatchesWithTarget_Variants(t *testing.T) {
	t.Parallel()
	pc := PointCutFunc(func(m reflect.Method) bool {
		return m.Name == "DoSomething"
	})
	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("expected match via value receiver lookup")
	}
	if pc.Matches(&TestUserService{}, "NonExistent") {
		t.Error("expected no match for non-existent method")
	}
}

func TestPointCutWithClass_MatchesAdvanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		pc        PointCutWithClass
		target    any
		method    string
		wantMatch bool
	}{
		{
			name:      "nil target always matches",
			pc:        PointCutWithClass{Class: func(rt reflect.Type) bool { return false }},
			target:    nil,
			method:    "Do",
			wantMatch: true,
		},
		{
			name: "class rejects",
			pc: PointCutWithClass{
				Class: func(rt reflect.Type) bool { return rt.Name() == "Other" },
				Match: func(m reflect.Method) bool { return true },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: false,
		},
		{
			name: "nil Match matches all methods",
			pc: PointCutWithClass{
				Class: func(rt reflect.Type) bool { return rt.Name() == "TestUserService" },
			},
			target:    &TestUserService{},
			method:    "Anything",
			wantMatch: true,
		},
		{
			name: "both class and method match",
			pc: PointCutWithClass{
				Class: func(rt reflect.Type) bool { return rt.Name() == "TestUserService" },
				Match: func(m reflect.Method) bool { return m.Name == "DoSomething" },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: true,
		},
		{
			name: "class matches but method rejects",
			pc: PointCutWithClass{
				Class: func(rt reflect.Type) bool { return rt.Name() == "TestUserService" },
				Match: func(m reflect.Method) bool { return m.Name == "Other" },
			},
			target:    &TestUserService{},
			method:    "DoSomething",
			wantMatch: false,
		},
		{
			name: "method not found on target",
			pc: PointCutWithClass{
				Class: func(rt reflect.Type) bool { return true },
				Match: func(m reflect.Method) bool { return true },
			},
			target:    &TestUserService{},
			method:    "NonExistent",
			wantMatch: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pc.Matches(tt.target, tt.method)
			if got != tt.wantMatch {
				t.Errorf("Matches(%v, %q) = %v, want %v", tt.target, tt.method, got, tt.wantMatch)
			}
		})
	}
}

func TestPointcutStringsAdvanced(t *testing.T) {
	t.Parallel()
	pc1 := MatchByName("Foo")
	pc2 := MatchByName("Bar")
	result := pointcutStrings([]PointCut{pc1, pc2})
	if len(result) != 2 || result[0] != "Foo" || result[1] != "Bar" {
		t.Errorf("unexpected result: %v", result)
	}
	if r := pointcutStrings(nil); len(r) != 0 {
		t.Errorf("expected empty slice for nil input, got %v", r)
	}
}

func TestPointCutImpl_Expression_AllBranches(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		pc       *pointCutImpl
		expected string
	}{
		{
			name:     "regex pattern",
			pc:       &pointCutImpl{regexPattern: "^Do"},
			expected: "^Do",
		},
		{
			name:     "name",
			pc:       &pointCutImpl{name: "DoSomething"},
			expected: "DoSomething",
		},
		{
			name:     "package path",
			pc:       &pointCutImpl{packagePath: "github.com/example"},
			expected: "package:github.com/example",
		},
		{
			name: "interface type",
			pc: &pointCutImpl{
				interfaceType: reflect.TypeFor[TestInterfaceForMatch](),
			},
			expected: "ByInterface(",
		},
		{
			name: "class and method",
			pc: &pointCutImpl{
				classMatcher:  func(rt reflect.Type) bool { return true },
				methodMatcher: func(m reflect.Method) bool { return true },
			},
			expected: "ByClassAndMethod",
		},
		{
			name:     "class only",
			pc:       &pointCutImpl{classMatcher: func(rt reflect.Type) bool { return true }},
			expected: "ByClass",
		},
		{
			name:     "method only",
			pc:       &pointCutImpl{methodMatcher: func(m reflect.Method) bool { return true }},
			expected: "ByMethod",
		},
		{
			name:     "empty",
			pc:       &pointCutImpl{},
			expected: "*",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.pc.Expression()
			if tt.name == "interface type" {
				if !strings.HasPrefix(got, "ByInterface(") {
					t.Errorf("Expression() = %q, want prefix %q", got, tt.expected)
				}
			} else if got != tt.expected {
				t.Errorf("Expression() = %q, want %q", got, tt.expected)
			}
		})
	}
}

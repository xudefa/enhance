package aop

import (
	"reflect"
	"testing"
)

type covPcfSvc struct{}

func (covPcfSvc) ValMethod() string { return "val" }

func (*covPcfSvc) PtrMethod() string { return "ptr" }

type covPcfIface interface {
	ValMethod() string
}

type covCacheAnno struct{}

type covOtherSvc struct{}

type errStub struct{}

func (*errStub) Error() string { return "stub" }

func matcherOf(t *testing.T, pc PointCut) MethodMatcher {
	t.Helper()
	impl, ok := pc.(*pointCutImpl)
	if !ok || impl.methodMatcher == nil {
		t.Fatalf("expected pointCutImpl with method matcher, got %T", pc)
	}
	return impl.methodMatcher
}

func classMatcherOf(t *testing.T, pc PointCut) ClassMatcher {
	t.Helper()
	impl, ok := pc.(*pointCutImpl)
	if !ok || impl.classMatcher == nil {
		t.Fatalf("expected pointCutImpl with class matcher, got %T", pc)
	}
	return impl.classMatcher
}

func TestPointCutFunc_Matches(t *testing.T) {
	t.Parallel()

	fn := PointCutFunc(func(m reflect.Method) bool { return m.Name == "ValMethod" })

	tests := []struct {
		name       string
		target     any
		methodName string
		want       bool
	}{
		{"nil target matching name", nil, "ValMethod", true},
		{"nil target non-matching name", nil, "OtherMethod", false},
		{"value receiver method on value target", covPcfSvc{}, "ValMethod", true},
		{"ptr-only method lookup path", &covPcfSvc{}, "PtrMethod", false},
		{"missing method", &covPcfSvc{}, "Missing", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := fn.Matches(tt.target, tt.methodName); got != tt.want {
				t.Errorf("Matches(%v, %q) = %v, want %v", tt.target, tt.methodName, got, tt.want)
			}
		})
	}
}

func TestPointCutWithClass_Matches(t *testing.T) {
	t.Parallel()

	matcher := func(m reflect.Method) bool { return m.Name == "ValMethod" }

	tests := []struct {
		name       string
		pc         PointCutWithClass
		target     any
		methodName string
		want       bool
	}{
		{"nil target returns true", PointCutWithClass{Match: matcher}, nil, "ValMethod", true},
		{
			"class mismatch rejects",
			PointCutWithClass{Class: func(reflect.Type) bool { return false }, Match: matcher},
			covPcfSvc{},
			"ValMethod",
			false,
		},
		{
			"nil match matcher accepts all methods",
			PointCutWithClass{Class: func(reflect.Type) bool { return true }},
			covPcfSvc{},
			"Whatever",
			true,
		},
		{"method match on value target", PointCutWithClass{Match: matcher}, covPcfSvc{}, "ValMethod", true},
		{"missing method rejected", PointCutWithClass{Match: matcher}, covPcfSvc{}, "Missing", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pc.Matches(tt.target, tt.methodName); got != tt.want {
				t.Errorf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCompositePointCut_MatchClassAndMatches(t *testing.T) {
	t.Parallel()

	matchAll := MatchClass(func(reflect.Type) bool { return true })
	matchNone := MatchClass(func(reflect.Type) bool { return false })
	typ := reflect.TypeOf(covPcfSvc{})
	target := covPcfSvc{}

	tests := []struct {
		name            string
		pc              PointCut
		wantMatchClass  bool
		wantMatchesWork bool
	}{
		{"AND both match", Compose(matchAll, matchAll), true, true},
		{"AND one fails class", Compose(matchAll, matchNone), false, false},
		{"OR one matches class", ComposeOr(matchNone, matchAll), true, true},
		{"OR none match class", ComposeOr(matchNone, matchNone), false, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pc.MatchClass(typ); got != tt.wantMatchClass {
				t.Errorf("MatchClass = %v, want %v", got, tt.wantMatchClass)
			}
			if got := tt.pc.Matches(target, "Any"); got != tt.wantMatchesWork {
				t.Errorf("Matches = %v, want %v", got, tt.wantMatchesWork)
			}
		})
	}
}

func TestMatchByAnnotation_MatcherTable(t *testing.T) {
	t.Parallel()
	annoType := reflect.TypeOf(covCacheAnno{})

	tests := []struct {
		name   string
		method string
		want   bool
	}{
		{"prefix form", "covCacheAnno_GetUser", true},
		{"contains form", "do_covCacheAnno_thing", true},
		{"suffix form", "save_covCacheAnno", true},
		{"no annotation", "GetUser", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			matcher := matcherOf(t, MatchByAnnotation(annoType))
			got := matcher(reflect.Method{Name: tt.method})
			if got != tt.want {
				t.Errorf("annotation match for %q = %v, want %v", tt.method, got, tt.want)
			}
		})
	}
}

func TestMatchInterface_Variants(t *testing.T) {
	t.Parallel()

	ifaceType := reflect.TypeOf((*covPcfIface)(nil)).Elem()

	tests := []struct {
		name  string
		input any
		check func(t *testing.T, pc PointCut)
	}{
		{
			name:  "reflect.Type non-interface rejected",
			input: reflect.TypeOf(int(0)),
			check: func(t *testing.T, pc PointCut) {
				if pc.MatchClass(ifaceType) {
					t.Error("expected no match for non-interface type input")
				}
				if pc.Matches(&covPcfSvc{}, "X") {
					t.Error("expected Matches to reject when only class matcher exists and it fails")
				}
			},
		},
		{
			name:  "nil any input rejected",
			input: nil,
			check: func(t *testing.T, pc PointCut) {
				if pc.MatchClass(ifaceType) {
					t.Error("expected no match for nil input")
				}
			},
		},
		{
			name:  "concrete struct value rejected",
			input: covPcfSvc{},
			check: func(t *testing.T, pc PointCut) {
				if pc.MatchClass(reflect.TypeOf(covPcfSvc{})) {
					t.Error("expected classMatcher to reject for non-interface input")
				}
			},
		},
		{
			name:  "valid interface type matches implementers",
			input: ifaceType,
			check: func(t *testing.T, pc PointCut) {
				if !pc.MatchClass(reflect.TypeOf(covPcfSvc{})) {
					t.Error("expected value type implementing interface to match")
				}
				if !pc.MatchClass(reflect.TypeOf((*covPcfSvc)(nil))) {
					t.Error("expected pointer type to be unwrapped and match")
				}
				if classMatcherOf(t, pc)(nil) {
					t.Error("nil type must not match")
				}
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.check(t, MatchInterface(tt.input))
		})
	}
}

func TestMatchInterface_MatchesRejectsNonImplementer(t *testing.T) {
	t.Parallel()

	pc := MatchInterface(reflect.TypeOf((*covPcfIface)(nil)).Elem())

	if pc.Matches(covOtherSvc{}, "Anything") {
		t.Error("target not implementing interface must not match")
	}
	if pc.MatchClass(reflect.TypeOf(covOtherSvc{})) {
		t.Error("non-implementing class must not pass MatchClass")
	}
	if pc.MatchClass(reflect.TypeOf(covPcfSvc{})) == false {
		t.Error("implementing class must pass MatchClass")
	}
}

func TestMatchByMethodSignature_Table(t *testing.T) {
	t.Parallel()

	intType := reflect.TypeOf(int(0))
	strType := reflect.TypeOf("")
	svcPtr := reflect.TypeOf((*covPcfSvc)(nil))
	getUser, _ := svcPtr.MethodByName("PtrMethod")

	tests := []struct {
		name    string
		matcher MethodMatcher
		method  reflect.Method
		want    bool
	}{
		{"name only matches", matcherOf(t, MatchByMethodSignature("PtrMethod")), getUser, true},
		{"name mismatch", matcherOf(t, MatchByMethodSignature("Other")), getUser, false},
		{"arity mismatch", matcherOf(t, MatchByMethodSignature("PtrMethod", intType, strType)), getUser, false},
		{"param type mismatch", matcherOf(t, MatchByMethodSignature("PtrMethod", strType)), getUser, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.matcher(tt.method); got != tt.want {
				t.Errorf("signature match = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchByReturnType_Table(t *testing.T) {
	t.Parallel()

	errType := reflect.TypeOf((*error)(nil)).Elem()
	intType := reflect.TypeOf(int(0))

	noReturn := reflect.Method{Name: "A", Type: reflect.TypeOf(func() {})}
	exactErr := reflect.Method{Name: "B", Type: reflect.TypeOf(func() error { return nil })}
	concreteImpl := reflect.Method{Name: "C", Type: reflect.TypeOf(func() *errStub { return nil })}
	wrongType := reflect.Method{Name: "D", Type: reflect.TypeOf(func() string { return "" })}

	tests := []struct {
		name    string
		m       reflect.Method
		matcher MethodMatcher
		want    bool
	}{
		{"no return values", noReturn, matcherOf(t, MatchByReturnType(errType)), false},
		{"exact interface match", exactErr, matcherOf(t, MatchByReturnType(errType)), true},
		{"concrete implements interface", concreteImpl, matcherOf(t, MatchByReturnType(errType)), true},
		{"return type mismatch", wrongType, matcherOf(t, MatchByReturnType(intType)), false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.matcher(tt.m); got != tt.want {
				t.Errorf("return-type match = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMatchByClassName_Table(t *testing.T) {
	t.Parallel()

	svcValue := reflect.TypeOf(covPcfSvc{})
	svcPtr := reflect.TypeOf((*covPcfSvc)(nil))

	tests := []struct {
		name      string
		className string
		typ       reflect.Type
		useNil    bool
		want      bool
	}{
		{"exact name value type", "covPcfSvc", svcValue, false, true},
		{"exact name pointer unwrapped", "covPcfSvc", svcPtr, false, true},
		{"exact name mismatch", "Other", svcValue, false, false},
		{"exact nil type via matcher", "covPcfSvc", nil, true, false},
		{"wildcard suffix", "*Svc", svcPtr, false, true},
		{"wildcard prefix", "cov*", svcValue, false, true},
		{"wildcard mismatch", "zzz*", svcValue, false, false},
		{"wildcard nil type via matcher", "*", nil, true, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pc := MatchByClassName(tt.className)
			if tt.useNil {
				if classMatcherOf(t, pc)(nil) != tt.want {
					t.Errorf("nil-type class match = %v, want %v", !tt.want, tt.want)
				}
				return
			}
			if got := pc.MatchClass(tt.typ); got != tt.want {
				t.Errorf("MatchByClassName(%q).MatchClass(%v) = %v, want %v", tt.className, tt.typ, got, tt.want)
			}
		})
	}
}

func TestMatchByPackage_NilGuard(t *testing.T) {
	t.Parallel()

	pc := MatchByPackage("github.com/xudefa/enhance/aop")

	if !pc.MatchClass(reflect.TypeOf(covPcfSvc{})) {
		t.Error("type from same package must match")
	}
	if classMatcherOf(t, pc)(nil) {
		t.Error("nil type must not match package pointcut")
	}
}

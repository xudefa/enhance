package aop

import (
	"reflect"
	"testing"
)

func TestMatchByName_ExactMatch(t *testing.T) {
	t.Parallel()

	pc := MatchByName("GetUser")

	if !pc.Matches(nil, "GetUser") {
		t.Error("should match exact name")
	}
	if pc.Matches(nil, "GetUsers") {
		t.Error("should not match partial name")
	}
}

func TestMatchByName_WildcardStar(t *testing.T) {
	t.Parallel()

	pc := MatchByName("Get*")

	if !pc.Matches(nil, "GetUser") {
		t.Error("should match Get prefix")
	}
	if !pc.Matches(nil, "Get") {
		t.Error("should match just Get")
	}
	if pc.Matches(nil, "SaveUser") {
		t.Error("should not match non-Get")
	}
}

func TestMatchByName_WildcardQuestion(t *testing.T) {
	t.Parallel()

	pc := MatchByName("Do?")

	if !pc.Matches(nil, "DoX") {
		t.Error("should match Do followed by one char")
	}
	if pc.Matches(nil, "Do") {
		t.Error("should not match Do without extra char")
	}
	if pc.Matches(nil, "DoIt") {
		t.Error("should not match Do with more than one extra char")
	}
}

func TestMatchByName_RegexWithSpecialChars(t *testing.T) {
	t.Parallel()

	// Pattern containing regex metacharacters should be treated as regex
	pc := MatchByName("^Do.*")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("should match regex pattern")
	}
	if pc.Matches(nil, "GetSomething") {
		t.Error("should not match non-matching regex")
	}
}

func TestMatchByRegex_ValidPattern(t *testing.T) {
	t.Parallel()

	pc := MatchByRegex("(?i)^do")

	if !pc.Matches(nil, "DoSomething") {
		t.Error("should match case-insensitive pattern")
	}
	if !pc.Matches(nil, "doSomething") {
		t.Error("should match lowercase")
	}
	if pc.Matches(nil, "GetSomething") {
		t.Error("should not match non-matching")
	}
}

func TestMatchByRegex_InvalidPattern(t *testing.T) {
	t.Parallel()

	pc := MatchByRegex("[invalid")

	// Invalid regex should not match anything
	if pc.Matches(nil, "anything") {
		t.Error("invalid regex should not match")
	}
}

func TestMatchByClassName_Exact(t *testing.T) {
	t.Parallel()

	pc := MatchByClassName("PointCutTestService")

	if !pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("should match exact class name")
	}
	if pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("should not match different class")
	}
}

func TestMatchByClassName_Wildcard(t *testing.T) {
	t.Parallel()

	pc := MatchByClassName("*Service")

	if !pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("should match *Service wildcard")
	}
	if !pc.Matches(&TestUserService{}, "DoSomething") {
		t.Error("should match *Service for TestUserService")
	}
}

func TestMatchByClassName_PointerType(t *testing.T) {
	t.Parallel()

	pc := MatchByClassName("PointCutTestService")

	// Pointer type should be unwrapped
	if !pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("should match pointer type after unwrapping")
	}
}

func TestMatchByClassName_NilTarget(t *testing.T) {
	t.Parallel()

	pc := MatchByClassName("AnyName")
	// Nil target with classMatcher returns false (no class to check)
	if pc.Matches(nil, "anything") {
		t.Error("nil target with classMatcher should return false")
	}
}

func TestMatchByMethodSignature_NameOnly(t *testing.T) {
	t.Parallel()

	pc := MatchByMethodSignature("GetUser")

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should match method name")
	}
	if pc.Matches(&TestServiceImpl{}, "DoSomething") {
		t.Error("should not match different name")
	}
}

func TestMatchByMethodSignature_WithParamTypes(t *testing.T) {
	t.Parallel()

	int64Type := reflect.TypeOf(int64(0))
	pc := MatchByMethodSignature("GetUser", int64Type)

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should match method with correct signature")
	}
}

func TestMatchByMethodSignature_WrongParamCount(t *testing.T) {
	t.Parallel()

	int64Type := reflect.TypeOf(int64(0))
	strType := reflect.TypeOf("")
	pc := MatchByMethodSignature("GetUser", int64Type, strType)

	// TestServiceImpl.GetUser has only 1 param
	if pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should not match with wrong param count")
	}
}

func TestMatchByReturnType_ErrorType(t *testing.T) {
	t.Parallel()

	errType := reflect.TypeOf((*error)(nil)).Elem()
	pc := MatchByReturnType(errType)

	// TestServiceImpl.GetUser returns string, not error
	if pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should not match when return type differs")
	}
}

func TestMatchByReturnType_StringType(t *testing.T) {
	t.Parallel()

	strType := reflect.TypeOf("")
	pc := MatchByReturnType(strType)

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should match string return type")
	}
}

func TestMatchByPackage_SamePackage(t *testing.T) {
	t.Parallel()

	pc := MatchByPackage("github.com/xudefa/enhance/aop")

	// PointCutTestService is in aop package
	if !pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("should match type in same package")
	}
}

func TestMatchByPackage_DifferentPackage(t *testing.T) {
	t.Parallel()

	pc := MatchByPackage("github.com/nonexistent/package")

	if pc.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("should not match type in different package")
	}
}

func TestMatchInterface_PtrToInterface(t *testing.T) {
	t.Parallel()

	pc := MatchInterface((*TestServiceInterface)(nil))

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should match implementer")
	}
	if pc.Matches("string", "anything") {
		t.Error("should not match non-implementer")
	}
}

func TestMatchInterface_ReflectType_Interface(t *testing.T) {
	t.Parallel()

	ifaceType := reflect.TypeOf((*TestServiceInterface)(nil)).Elem()
	pc := MatchInterface(ifaceType)

	if !pc.Matches(&TestServiceImpl{}, "GetUser") {
		t.Error("should match implementer via reflect.Type")
	}
}

func TestMatchInterface_ReflectType_NonInterface(t *testing.T) {
	t.Parallel()

	nonIfaceType := reflect.TypeOf("")
	pc := MatchInterface(nonIfaceType)

	if pc.Matches("test", "anything") {
		t.Error("non-interface reflect.Type should not match")
	}
}

func TestMatchInterface_Nil(t *testing.T) {
	t.Parallel()

	pc := MatchInterface(nil)

	if pc.Matches(nil, "anything") {
		t.Error("nil input should not match")
	}
}

func TestPointCutFunc_Matches_NilTarget(t *testing.T) {
	t.Parallel()

	fn := PointCutFunc(func(m reflect.Method) bool {
		return m.Name == "DoSomething"
	})

	if !fn.Matches(nil, "DoSomething") {
		t.Error("should match method name with nil target")
	}
	if fn.Matches(nil, "OtherMethod") {
		t.Error("should not match non-matching name with nil target")
	}
}

func TestPointCutFunc_MatchClass_AlwaysTrue(t *testing.T) {
	t.Parallel()

	fn := PointCutFunc(func(m reflect.Method) bool { return true })

	if !fn.MatchClass(reflect.TypeOf("")) {
		t.Error("PointCutFunc.MatchClass should always return true")
	}
	if !fn.MatchClass(reflect.TypeOf(0)) {
		t.Error("PointCutFunc.MatchClass should always return true")
	}
}

func TestPointCutWithClass_NilMatch(t *testing.T) {
	t.Parallel()

	pwc := PointCutWithClass{
		Class: func(t reflect.Type) bool { return true },
	}

	// nil Match should match all methods
	if !pwc.Matches(&PointCutTestService{}, "Anything") {
		t.Error("nil Match should match all methods")
	}
}

func TestPointCutWithClass_NilTarget(t *testing.T) {
	t.Parallel()

	pwc := PointCutWithClass{}
	if !pwc.Matches(nil, "Anything") {
		t.Error("nil target should return true")
	}
}

func TestPointCutWithClass_MatchClass_NilClass(t *testing.T) {
	t.Parallel()

	pwc := PointCutWithClass{}
	if !pwc.MatchClass(reflect.TypeOf("")) {
		t.Error("nil Class matcher should match all")
	}
}

func TestPointCutWithClass_MatchClass_PointerType(t *testing.T) {
	t.Parallel()

	pwc := PointCutWithClass{
		Class: func(t reflect.Type) bool { return t.Name() == "PointCutTestService" },
	}

	ptrType := reflect.PointerTo(reflect.TypeOf(PointCutTestService{}))
	if !pwc.MatchClass(ptrType) {
		t.Error("pointer type should be unwrapped")
	}
}

func TestPointCutImpl_Expression_Complete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pc   *pointCutImpl
		want string
	}{
		{
			"regex pattern",
			&pointCutImpl{regexPattern: "Do.*"},
			"Do.*",
		},
		{
			"name",
			&pointCutImpl{name: "myMethod"},
			"myMethod",
		},
		{
			"package path",
			&pointCutImpl{packagePath: "mypkg"},
			"package:mypkg",
		},
		{
			"class and method",
			&pointCutImpl{
				classMatcher:  func(reflect.Type) bool { return true },
				methodMatcher: func(reflect.Method) bool { return true },
			},
			"ByClassAndMethod",
		},
		{
			"class only",
			&pointCutImpl{classMatcher: func(reflect.Type) bool { return true }},
			"ByClass",
		},
		{
			"method only",
			&pointCutImpl{methodMatcher: func(reflect.Method) bool { return true }},
			"ByMethod",
		},
		{
			"empty",
			&pointCutImpl{},
			"*",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.pc.Expression(); got != tt.want {
				t.Errorf("Expression() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPointCutImpl_MatchClass_NilClassMatcher(t *testing.T) {
	t.Parallel()

	pc := &pointCutImpl{}
	if !pc.MatchClass(reflect.TypeOf("")) {
		t.Error("nil classMatcher should pass")
	}
}

func TestPointCutImpl_Matches_NilTarget_NoMatchers(t *testing.T) {
	t.Parallel()

	pc := &pointCutImpl{}
	if !pc.Matches(nil, "Anything") {
		t.Error("nil target with no matchers should match")
	}
}

func TestPointCutImpl_Matches_NilTarget_WithClassMatcher(t *testing.T) {
	t.Parallel()

	pc := &pointCutImpl{classMatcher: func(reflect.Type) bool { return true }}
	if pc.Matches(nil, "Anything") {
		t.Error("nil target with classMatcher should not match")
	}
}

func TestPointCutImpl_Matches_NilTarget_WithMethodMatcher(t *testing.T) {
	t.Parallel()

	pc := &pointCutImpl{methodMatcher: func(m reflect.Method) bool { return m.Name == "OK" }}
	if !pc.Matches(nil, "OK") {
		t.Error("nil target with matching methodMatcher should match")
	}
	if pc.Matches(nil, "NO") {
		t.Error("nil target with non-matching methodMatcher should not match")
	}
}

package aop

import (
	"reflect"
	"strings"
	"testing"
)

func TestCompose_AND_Matches(t *testing.T) {
	t.Parallel()

	pc1 := MatchByName("Get*")
	pc2 := MatchByName("*User")

	composed := Compose(pc1, pc2)

	if !composed.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("AND: should match when both match")
	}
	if composed.Matches(&PointCutTestService{}, "GetItem") {
		t.Error("AND: should not match when only pc1 matches")
	}
	if composed.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("AND: should not match when only pc2 matches")
	}
}

func TestCompose_OR_Matches(t *testing.T) {
	t.Parallel()

	pc1 := MatchByName("GetUser")
	pc2 := MatchByName("SaveUser")

	composed := ComposeOr(pc1, pc2)

	if !composed.Matches(&PointCutTestService{}, "GetUser") {
		t.Error("OR: should match when pc1 matches")
	}
	if !composed.Matches(&PointCutTestService{}, "SaveUser") {
		t.Error("OR: should match when pc2 matches")
	}
	if composed.Matches(&PointCutTestService{}, "DeleteUser") {
		t.Error("OR: should not match when none match")
	}
}

func TestCompose_AND_MatchClass(t *testing.T) {
	t.Parallel()

	pc1 := MatchByClassName("TestUserService")
	pc2 := MatchByClassName("*Service")

	composed := Compose(pc1, pc2)

	if !composed.MatchClass(reflect.TypeOf(TestUserService{})) {
		t.Error("AND MatchClass: should match when both match")
	}
	if composed.MatchClass(reflect.TypeOf(PointCutTestService{})) {
		t.Error("AND MatchClass: should not match when only pc2 matches")
	}
}

func TestCompose_OR_MatchClass(t *testing.T) {
	t.Parallel()

	pc1 := MatchByClassName("TestUserService")
	pc2 := MatchByClassName("NonExistent")

	composed := ComposeOr(pc1, pc2)

	if !composed.MatchClass(reflect.TypeOf(TestUserService{})) {
		t.Error("OR MatchClass: should match when pc1 matches")
	}
	if composed.MatchClass(reflect.TypeOf(PointCutTestService{})) {
		t.Error("OR MatchClass: should not match when none match")
	}
}

func TestCompose_AND_Expression(t *testing.T) {
	t.Parallel()

	pc1 := MatchByName("Get*")
	pc2 := MatchByName("*User")

	composed := Compose(pc1, pc2)
	expr := composed.Expression()

	if !strings.HasPrefix(expr, "AND(") {
		t.Errorf("Expression() = %q, should start with AND(", expr)
	}
	if !strings.HasSuffix(expr, ")") {
		t.Errorf("Expression() = %q, should end with )", expr)
	}
}

func TestCompose_OR_Expression(t *testing.T) {
	t.Parallel()

	pc1 := MatchByName("Get*")
	pc2 := MatchByName("Save*")

	composed := ComposeOr(pc1, pc2)
	expr := composed.Expression()

	if !strings.HasPrefix(expr, "OR(") {
		t.Errorf("Expression() = %q, should start with OR(", expr)
	}
}

func TestCompose_Empty(t *testing.T) {
	t.Parallel()

	// AND with no pointcuts should match everything
	composed := Compose()
	if !composed.Matches(nil, "Anything") {
		t.Error("AND with no pointcuts should match")
	}
	if !composed.MatchClass(reflect.TypeOf("")) {
		t.Error("AND MatchClass with no pointcuts should match")
	}

	// OR with no pointcuts should match nothing
	composedOr := ComposeOr()
	if composedOr.Matches(nil, "Anything") {
		t.Error("OR with no pointcuts should not match")
	}
	if composedOr.MatchClass(reflect.TypeOf("")) {
		t.Error("OR MatchClass with no pointcuts should not match")
	}
}

func TestCompose_SinglePointcut(t *testing.T) {
	t.Parallel()

	pc := MatchByName("Do*")

	and := Compose(pc)
	if !and.Matches(nil, "DoSomething") {
		t.Error("AND with single pointcut should match")
	}
	if and.Matches(nil, "GetSomething") {
		t.Error("AND with single pointcut should not match non-matching")
	}

	or := ComposeOr(pc)
	if !or.Matches(nil, "DoSomething") {
		t.Error("OR with single pointcut should match")
	}
	if or.Matches(nil, "GetSomething") {
		t.Error("OR with single pointcut should not match non-matching")
	}
}

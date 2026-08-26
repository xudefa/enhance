package aop

import (
	"context"
	"reflect"
	"testing"
)

func TestNewWeaver(t *testing.T) {
	t.Parallel()

	w := NewWeaver()
	if w == nil {
		t.Fatal("expected non-nil weaver")
	}
}

func TestWeaver_AddAspects(t *testing.T) {
	t.Parallel()

	w := NewWeaver()
	aspect := &AspectMeta{
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 0),
	}

	w.AddAspects(aspect)
}

func TestWeaver_Weave_NilTarget(t *testing.T) {
	t.Parallel()

	w := NewWeaver()
	result := w.Weave(nil)
	if result != nil {
		t.Error("expected nil result for nil target")
	}
}

func TestWeaver_Weave_NoAspects(t *testing.T) {
	t.Parallel()

	w := NewWeaver()
	target := &testTypedService{Name: "test"}
	result := w.Weave(target)
	if result != target {
		t.Error("expected original target when no aspects registered")
	}
}

func TestWeaver_Weave_WithAspects(t *testing.T) {
	t.Parallel()

	w := NewWeaver()
	aspect := &AspectMeta{
		PointCut: MatchByName("DoWork"),
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 0),
	}
	w.AddAspects(aspect)

	target := &testTypedService{Name: "test"}
	result := w.Weave(target)
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestAopRegistry_NewAndRegister(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}

	aspect := &AspectMeta{
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 0),
	}
	reg.RegisterAspect(aspect)

	aspects := reg.GetAspects()
	if len(aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(aspects))
	}
}

func TestAopRegistry_RegisterWeaver(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	w := NewWeaver()

	reg.RegisterWeaver("testBean", w)

	retrieved, ok := reg.GetWeaver("testBean")
	if !ok {
		t.Fatal("expected to find weaver")
	}
	if retrieved != w {
		t.Error("expected retrieved weaver to match registered one")
	}
}

func TestAopRegistry_GetWeaver_NotFound(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	_, ok := reg.GetWeaver("nonexistent")
	if ok {
		t.Error("expected not found")
	}
}

func TestAopRegistry_WeaveIfNeeded_Found(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	w := NewWeaver()
	reg.RegisterWeaver("testBean", w)

	target := &testTypedService{Name: "test"}
	result := reg.WeaveIfNeeded("testBean", target)
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestAopRegistry_WeaveIfNeeded_NotFound(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	target := &testTypedService{Name: "test"}
	result := reg.WeaveIfNeeded("nonexistent", target)
	if result != target {
		t.Error("expected original target when weaver not found")
	}
}

func TestAopRegistry_MatchAspectsForType(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()

	// 注册匹配的切面
	aspect1 := &AspectMeta{
		PointCut: MatchByName("DoWork"),
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 0),
	}
	reg.RegisterAspect(aspect1)

	// 注册不匹配的切面
	aspect2 := &AspectMeta{
		PointCut: MatchByName("NonExistentMethod"),
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 1),
	}
	reg.RegisterAspect(aspect2)

	// 注册无 PointCut 的切面
	aspect3 := &AspectMeta{
		Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
			return nil, nil
		}, 2),
	}
	reg.RegisterAspect(aspect3)

	targetType := reflect.TypeOf(&testTypedService{})
	matched := reg.MatchAspectsForType(targetType)

	if len(matched) != 1 {
		t.Errorf("expected 1 matched aspect, got %d", len(matched))
	}
	if len(matched) > 0 && matched[0] != aspect1 {
		t.Error("expected matched aspect to be aspect1")
	}
}

func TestAopRegistry_MatchAspectsForType_Empty(t *testing.T) {
	t.Parallel()

	reg := NewAopRegistry()
	targetType := reflect.TypeOf(&testTypedService{})
	matched := reg.MatchAspectsForType(targetType)
	if len(matched) != 0 {
		t.Errorf("expected 0 matched aspects, got %d", len(matched))
	}
}

package aop

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// typedService 测试用的类型化服务。
type typedService struct {
	val int
}

func (s *typedService) Run(v int) int {
	return s.val + v
}

func (s *typedService) Fail() error {
	return errors.New("service error")
}

func (s *typedService) GetValue() int {
	return s.val
}

// mockJoinPointForTyped 测试用的 JoinPoint 模拟实现。
type mockJoinPointForTyped struct {
	target any
	method string
	args   []any
	result any
	err    error
	ctx    context.Context
}

func (j *mockJoinPointForTyped) Target() any           { return j.target }
func (j *mockJoinPointForTyped) Method() string        { return j.method }
func (j *mockJoinPointForTyped) Args() []any           { return j.args }
func (j *mockJoinPointForTyped) Proceed() (any, error) { return j.result, j.err }
func (j *mockJoinPointForTyped) ProceedWithArgs(args []any) (any, error) {
	return j.result, j.err
}
func (j *mockJoinPointForTyped) Context() context.Context {
	if j.ctx == nil {
		return context.Background()
	}
	return j.ctx
}
func (j *mockJoinPointForTyped) GetResult() any     { return j.result }
func (j *mockJoinPointForTyped) GetError() error    { return j.err }
func (j *mockJoinPointForTyped) SetResult(v any)    { j.result = v }
func (j *mockJoinPointForTyped) SetError(err error) { j.err = err }

func TestNewBefore_TypedTarget(t *testing.T) {
	t.Parallel()

	var capturedTarget *typedService
	var capturedMethod string

	advice := NewBefore[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			capturedTarget = target
			capturedMethod = jp.Method()
			return nil
		},
		0,
	)

	if advice.Type() != AdviceTypeBefore {
		t.Fatalf("expected AdviceTypeBefore, got %v", advice.Type())
	}

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 42},
		method: "Run",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTarget == nil {
		t.Fatal("expected captured target to be non-nil")
	}
	if capturedTarget.val != 42 {
		t.Errorf("expected target.val = 42, got %d", capturedTarget.val)
	}
	if capturedMethod != "Run" {
		t.Errorf("expected method = 'Run', got '%s'", capturedMethod)
	}
}

func TestNewBefore_TypeMismatch(t *testing.T) {
	t.Parallel()

	advice := NewBefore[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			return nil
		},
		0,
	)

	jp := &mockJoinPointForTyped{
		target: "not a typedService",
		method: "Run",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err == nil {
		t.Fatal("expected type mismatch error")
	}
	if err.Error() == "" {
		t.Error("expected error message to be non-empty")
	}
}

func TestNewBefore_ReturnsError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("before error")
	advice := NewBefore[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			return expectedErr
		},
		0,
	)

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 1},
		method: "Run",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestNewAfter_TypedTarget(t *testing.T) {
	t.Parallel()

	var capturedTarget *typedService
	var capturedMethod string

	advice := NewAfter[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			capturedTarget = target
			capturedMethod = jp.Method()
			return nil
		},
		0,
	)

	if advice.Type() != AdviceTypeAfter {
		t.Fatalf("expected AdviceTypeAfter, got %v", advice.Type())
	}

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 99},
		method: "GetValue",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTarget == nil {
		t.Fatal("expected captured target to be non-nil")
	}
	if capturedTarget.val != 99 {
		t.Errorf("expected target.val = 99, got %d", capturedTarget.val)
	}
	if capturedMethod != "GetValue" {
		t.Errorf("expected method = 'GetValue', got '%s'", capturedMethod)
	}
}

func TestNewAround_TypedTarget(t *testing.T) {
	t.Parallel()

	var beforeCalled, afterCalled bool
	var capturedTarget *typedService

	advice := NewAround[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint, proceed func() (any, error)) (any, error) {
			beforeCalled = true
			capturedTarget = target
			result, err := proceed()
			afterCalled = true
			return result, err
		},
		0,
	)

	if advice.Type() != AdviceTypeAround {
		t.Fatalf("expected AdviceTypeAround, got %v", advice.Type())
	}

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 7},
		method: "Run",
		result: 10,
	}

	result, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !beforeCalled {
		t.Error("expected before logic to be called")
	}
	if !afterCalled {
		t.Error("expected after logic to be called")
	}
	if capturedTarget == nil || capturedTarget.val != 7 {
		t.Errorf("expected target with val=7, got %v", capturedTarget)
	}
	if result != 10 {
		t.Errorf("expected result = 10, got %v", result)
	}
}

func TestNewAround_CanSkipProceed(t *testing.T) {
	t.Parallel()

	advice := NewAround[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint, proceed func() (any, error)) (any, error) {
			return "short-circuit", nil
		},
		0,
	)

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 1},
		method: "Run",
		result: "should not be returned",
	}

	result, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "short-circuit" {
		t.Errorf("expected 'short-circuit', got %v", result)
	}
}

func TestNewAfterReturning_TypedTarget(t *testing.T) {
	t.Parallel()

	var capturedTarget *typedService
	var capturedResult any

	advice := NewAfterReturning[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint, result any) error {
			capturedTarget = target
			capturedResult = result
			return nil
		},
		0,
	)

	if advice.Type() != AdviceTypeAfterReturning {
		t.Fatalf("expected AdviceTypeAfterReturning, got %v", advice.Type())
	}

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 5},
		method: "GetValue",
		result: 42,
	}

	result, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTarget == nil {
		t.Fatal("expected captured target to be non-nil")
	}
	if capturedResult != 42 {
		t.Errorf("expected captured result = 42, got %v", capturedResult)
	}
	if result != 42 {
		t.Errorf("expected result = 42, got %v", result)
	}
}

func TestNewAfterThrowing_TypedTarget(t *testing.T) {
	t.Parallel()

	var capturedTarget *typedService
	var capturedErr error

	advice := NewAfterThrowing[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint, err error) error {
			capturedTarget = target
			capturedErr = err
			return nil
		},
		0,
	)

	if advice.Type() != AdviceTypeAfterThrowing {
		t.Fatalf("expected AdviceTypeAfterThrowing, got %v", advice.Type())
	}

	expectedErr := errors.New("service error")
	jp := &mockJoinPointForTyped{
		target: &typedService{val: 1},
		method: "Fail",
		err:    expectedErr,
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != expectedErr {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}

	if capturedTarget == nil {
		t.Fatal("expected captured target to be non-nil")
	}
	if capturedErr != expectedErr {
		t.Errorf("expected captured error = %v, got %v", expectedErr, capturedErr)
	}
}

func TestNewAfterThrowing_SkipsOnSuccess(t *testing.T) {
	t.Parallel()

	called := false
	advice := NewAfterThrowing[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint, err error) error {
			called = true
			return nil
		},
		0,
	)

	jp := &mockJoinPointForTyped{
		target: &typedService{val: 1},
		method: "Run",
		result: "success",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Error("expected AfterThrowing to be skipped on success")
	}
}

func TestTypedAdvice_WithIntegration(t *testing.T) {
	t.Parallel()

	var beforeCalled, afterCalled, aroundBefore, aroundAfter bool
	var capturedVal int

	svc := &typedService{val: 10}

	weaver := NewWeaver()
	weaver.AddAspects(
		&AspectMeta{
			PointCut: MatchByName("Run"),
			Advice: NewBefore[*typedService](
				func(ctx context.Context, target *typedService, jp JoinPoint) error {
					beforeCalled = true
					capturedVal = target.val
					return nil
				},
				0,
			),
		},
		&AspectMeta{
			PointCut: MatchByName("Run"),
			Advice: NewAfter[*typedService](
				func(ctx context.Context, target *typedService, jp JoinPoint) error {
					afterCalled = true
					return nil
				},
				1,
			),
		},
		&AspectMeta{
			PointCut: MatchByName("Run"),
			Advice: NewAround[*typedService](
				func(ctx context.Context, target *typedService, jp JoinPoint, proceed func() (any, error)) (any, error) {
					aroundBefore = true
					result, err := proceed()
					aroundAfter = true
					return result, err
				},
				2,
			),
		},
	)

	proxy := weaver.Weave(svc)
	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}

	reflectiveProxy, ok := proxy.(*ReflectiveAopProxy)
	if !ok {
		t.Fatalf("expected *ReflectiveAopProxy, got %T", proxy)
	}

	result, err := reflectiveProxy.Call("Run", 5)
	if err != nil {
		t.Fatalf("Call error = %v", err)
	}

	if !beforeCalled {
		t.Error("expected Before advice to be called")
	}
	if !afterCalled {
		t.Error("expected After advice to be called")
	}
	if !aroundBefore {
		t.Error("expected Around before logic to be called")
	}
	if !aroundAfter {
		t.Error("expected Around after logic to be called")
	}
	if capturedVal != 10 {
		t.Errorf("expected captured val = 10, got %d", capturedVal)
	}

	if result != 15 {
		t.Errorf("expected result = 15, got %v", result)
	}
}

func TestTypedAdvice_InterfaceTarget(t *testing.T) {
	t.Parallel()

	type Runner interface {
		Run(v int) int
	}

	var capturedTarget Runner

	advice := NewBefore[Runner](
		func(ctx context.Context, target Runner, jp JoinPoint) error {
			capturedTarget = target
			return nil
		},
		0,
	)

	svc := &typedService{val: 42}
	jp := &mockJoinPointForTyped{
		target: svc,
		method: "Run",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedTarget == nil {
		t.Fatal("expected captured target to be non-nil")
	}
	if capturedTarget.Run(0) != 42 {
		t.Errorf("expected target.Run(0) = 42, got %d", capturedTarget.Run(0))
	}
}

func TestTypedAdvice_Order(t *testing.T) {
	t.Parallel()

	advice0 := NewBefore[*typedService](func(ctx context.Context, target *typedService, jp JoinPoint) error {
		return nil
	}, 0)

	advice5 := NewBefore[*typedService](func(ctx context.Context, target *typedService, jp JoinPoint) error {
		return nil
	}, 5)

	advice10 := NewBefore[*typedService](func(ctx context.Context, target *typedService, jp JoinPoint) error {
		return nil
	}, 10)

	if advice0.Order() != 0 {
		t.Errorf("expected order 0, got %d", advice0.Order())
	}
	if advice5.Order() != 5 {
		t.Errorf("expected order 5, got %d", advice5.Order())
	}
	if advice10.Order() != 10 {
		t.Errorf("expected order 10, got %d", advice10.Order())
	}
}

func TestTypedAdvice_NilTarget(t *testing.T) {
	t.Parallel()

	advice := NewBefore[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			return nil
		},
		0,
	)

	jp := &mockJoinPointForTyped{
		target: nil,
		method: "Run",
	}

	_, err := advice.Execute(context.Background(), jp)
	if err == nil {
		t.Fatal("expected error for nil target")
	}
}

func TestTypedAdvice_ContextPropagation(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}
	expectedValue := "test-value"

	var capturedValue string

	advice := NewBefore[*typedService](
		func(ctx context.Context, target *typedService, jp JoinPoint) error {
			if v, ok := ctx.Value(ctxKey{}).(string); ok {
				capturedValue = v
			}
			return nil
		},
		0,
	)

	ctx := context.WithValue(context.Background(), ctxKey{}, expectedValue)
	jp := &mockJoinPointForTyped{
		target: &typedService{val: 1},
		method: "Run",
	}

	_, err := advice.Execute(ctx, jp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedValue != expectedValue {
		t.Errorf("expected context value = '%s', got '%s'", expectedValue, capturedValue)
	}
}

func ExampleNewBefore() {
	type UserService struct {
		name string
	}

	advice := NewBefore[*UserService](
		func(ctx context.Context, target *UserService, jp JoinPoint) error {
			fmt.Printf("Before: calling %s on user %s\n", jp.Method(), target.name)
			return nil
		},
		0,
	)

	fmt.Printf("Advice type: %s\n", advice.Type().String())

	// Output:
	// Advice type: before
}

func ExampleNewAround() {
	type OrderService struct{}

	advice := NewAround[*OrderService](
		func(ctx context.Context, target *OrderService, jp JoinPoint, proceed func() (any, error)) (any, error) {
			fmt.Printf("Before: %s\n", jp.Method())
			result, err := proceed()
			fmt.Printf("After: %s\n", jp.Method())
			return result, err
		},
		0,
	)

	fmt.Printf("Advice type: %s\n", advice.Type().String())

	// Output:
	// Advice type: around
}

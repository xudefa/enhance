package chain

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/xudefa/enhance/aop"
)

type mockAdvice struct {
	called bool
}

func (m *mockAdvice) Type() aop.AdviceType {
	return aop.AdviceTypeAround
}

func (m *mockAdvice) Order() int {
	return 0
}

func (m *mockAdvice) Execute(ctx context.Context, jp aop.JoinPoint) (any, error) {
	m.called = true
	return jp.Proceed()
}

type mockTarget struct{}

func (m *mockTarget) DoWork() string {
	return "work"
}

type mockTargetWithArgs struct{}

func (m *mockTargetWithArgs) DoWork(name string, count int) string {
	return name
}

type mockAdviceWithResult struct {
	called bool
	result any
	err    error
}

func (m *mockAdviceWithResult) Type() aop.AdviceType {
	return aop.AdviceTypeAround
}

func (m *mockAdviceWithResult) Order() int {
	return 0
}

func (m *mockAdviceWithResult) Execute(ctx context.Context, jp aop.JoinPoint) (any, error) {
	m.called = true
	return m.result, m.err
}

type errWrapper struct{ err error }

func (e *errWrapper) Error() string { return e.err.Error() }

type orderedAdvice struct {
	name       string
	order      int
	orderSlice *[]string
}

func (o *orderedAdvice) Type() aop.AdviceType {
	return aop.AdviceTypeAround
}

func (o *orderedAdvice) Order() int {
	return o.order
}

func (o *orderedAdvice) Execute(ctx context.Context, jp aop.JoinPoint) (any, error) {
	*o.orderSlice = append(*o.orderSlice, o.name)
	return jp.Proceed()
}

func TestAdviceChain_Matches(t *testing.T) {
	t.Parallel()
	pc := aop.MatchByName("Do*")

	if !pc.Matches(&mockTarget{}, "DoWork") {
		t.Error("expected pointcut to match DoWork")
	}

	if pc.Matches(&mockTarget{}, "OtherMethod") {
		t.Error("expected pointcut not to match OtherMethod")
	}
}

func TestAdvisor_Matches(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}

	advisor := Advisor{
		Pointcut: aop.MatchByName("Do*"),
		Advice:   advice,
	}

	if !advisor.Matches("DoWork") {
		t.Error("expected advisor to match DoWork")
	}

	if advisor.Matches("OtherMethod") {
		t.Error("expected advisor not to match OtherMethod")
	}
}

func TestMethodInvocation(t *testing.T) {
	t.Parallel()
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	invocation := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  []reflect.Value{},
	}

	if invocation.Target() != target {
		t.Error("expected Target to return target object")
	}

	if invocation.Method() != "DoWork" {
		t.Errorf("expected Method to return 'DoWork', got %v", invocation.Method())
	}

	if len(invocation.Args()) != 0 {
		t.Error("expected Args to return empty slice")
	}

	result, err := invocation.Proceed()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
}

func TestNewAdviceChain(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}

	advisors := []Advisor{
		{Pointcut: aop.MatchByName("Do*"), Advice: advice},
	}

	chain := NewAdviceChain(advisors)
	if chain == nil {
		t.Error("expected chain to be created")
	}
}

func TestMethodInvocation_Context(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  context.Context
	}{
		{
			name: "custom context",
			ctx:  context.WithValue(context.Background(), "key", "val"),
		},
		{
			name: "nil context returns Background",
			ctx:  nil,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inv := &MethodInvocation{ctx: tt.ctx}
			got := inv.Context()
			if tt.ctx == nil {
				if got != context.Background() {
					t.Error("expected Background context")
				}
			} else {
				if got.Value("key") != "val" {
					t.Error("expected custom context value")
				}
			}
		})
	}
}

func TestMethodInvocation_ResultAndError(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")
	errOther := errors.New("err")

	tests := []struct {
		name       string
		setResult  any
		setErr     error
		wantResult any
		wantErr    error
	}{
		{
			name:       "set and get result",
			setResult:  "hello",
			setErr:     nil,
			wantResult: "hello",
			wantErr:    nil,
		},
		{
			name:       "set and get error",
			setResult:  nil,
			setErr:     errBoom,
			wantResult: nil,
			wantErr:    errBoom,
		},
		{
			name:       "set both result and error",
			setResult:  42,
			setErr:     errOther,
			wantResult: 42,
			wantErr:    errOther,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			inv := &MethodInvocation{}
			inv.SetResult(tt.setResult)
			inv.SetError(tt.setErr)

			if inv.GetResult() != tt.wantResult {
				t.Errorf("GetResult: got %v, want %v", inv.GetResult(), tt.wantResult)
			}
			if inv.GetError() != tt.wantErr {
				t.Errorf("GetError: got %v, want %v", inv.GetError(), tt.wantErr)
			}
		})
	}
}

func TestMethodInvocation_ArgsWithArgs(t *testing.T) {
	t.Parallel()
	target := &mockTargetWithArgs{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	inv := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  []reflect.Value{reflect.ValueOf("hello"), reflect.ValueOf(5)},
	}

	args := inv.Args()
	if len(args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(args))
	}
	if args[0] != "hello" {
		t.Errorf("expected first arg 'hello', got %v", args[0])
	}
	if args[1] != 5 {
		t.Errorf("expected second arg 5, got %v", args[1])
	}
}

func TestMethodInvocation_ProceedWithArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		proceedFn  func() (any, error)
		newArgs    []any
		wantResult any
		wantErr    bool
	}{
		{
			name:       "without proceed function",
			proceedFn:  nil,
			newArgs:    []any{"custom", 1},
			wantResult: "custom",
			wantErr:    false,
		},
		{
			name:      "with proceed function",
			proceedFn: func() (any, error) { return "proceeded", nil },
			newArgs:   []any{"ignored", 2},
			wantResult: "proceeded",
			wantErr:    false,
		},
		{
			name:      "with proceed function error",
			proceedFn: func() (any, error) { return nil, errors.New("perr") },
			newArgs:   []any{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target := &mockTargetWithArgs{}
			method, _ := reflect.TypeOf(target).MethodByName("DoWork")

			inv := &MethodInvocation{
				TargetObj: target,
				MethodObj: method,
				ArgsList:  []reflect.Value{reflect.ValueOf("old"), reflect.ValueOf(1)},
				proceed:   tt.proceedFn,
			}

			result, err := inv.ProceedWithArgs(tt.newArgs)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result != tt.wantResult {
				t.Errorf("got %v, want %v", result, tt.wantResult)
			}
		})
	}
}

func TestMethodInvocation_ProceedWithArgs_EmptyArgsPanics(t *testing.T) {
	t.Parallel()
	target := &mockTargetWithArgs{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	inv := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  []reflect.Value{reflect.ValueOf("old"), reflect.ValueOf(1)},
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for too few arguments")
		}
	}()
	inv.ProceedWithArgs([]any{})
}

func TestMethodInvocation_ProceedWithArgs_WithCustomArgs(t *testing.T) {
	t.Parallel()
	target := &mockTargetWithArgs{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	inv := &MethodInvocation{
		TargetObj: target,
		MethodObj: method,
		ArgsList:  []reflect.Value{reflect.ValueOf("old"), reflect.ValueOf(1)},
	}

	result, err := inv.ProceedWithArgs([]any{"new", 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "new" {
		t.Errorf("expected 'new', got %v", result)
	}
}

func TestExtractResults(t *testing.T) {
	t.Parallel()

	strVal := func(s string) reflect.Value { return reflect.ValueOf(s) }
	intVal := func(i int) reflect.Value { return reflect.ValueOf(i) }
	errVal := func(e error) reflect.Value {
		if e == nil {
			var nilErr error
			return reflect.ValueOf(&nilErr).Elem()
		}
		return reflect.ValueOf(&errWrapper{err: e})
	}

	tests := []struct {
		name       string
		results    []reflect.Value
		wantVal    any
		wantErr    bool
		wantMulti  bool
		multiLen   int
	}{
		{
			name:    "empty results",
			results: []reflect.Value{},
			wantVal: nil,
		},
		{
			name:    "single string result",
			results: []reflect.Value{strVal("hello")},
			wantVal: "hello",
		},
		{
			name:    "single int result",
			results: []reflect.Value{intVal(42)},
			wantVal: 42,
		},
		{
			name:      "two results string and nil error",
			results:   []reflect.Value{strVal("ok"), errVal(nil)},
			wantVal:   "ok",
			wantMulti: false,
		},
		{
			name:    "two results string and non-nil error",
			results: []reflect.Value{strVal("ok"), errVal(errors.New("fail"))},
			wantErr: true,
		},
		{
			name:      "three results with nil error at end",
			results:   []reflect.Value{strVal("a"), intVal(1), errVal(nil)},
			wantMulti: true,
			multiLen:  2,
		},
		{
			name:      "two non-error results",
			results:   []reflect.Value{strVal("x"), intVal(99)},
			wantMulti: true,
			multiLen:  2,
		},
		{
			name:      "three non-error results",
			results:   []reflect.Value{strVal("x"), intVal(1), intVal(2)},
			wantMulti: true,
			multiLen:  3,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			val, err := extractResults(tt.results)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.wantMulti {
				slice, ok := val.([]any)
				if !ok {
					t.Fatalf("expected []any, got %T", val)
				}
				if len(slice) != tt.multiLen {
					t.Errorf("expected %d results, got %d", tt.multiLen, len(slice))
				}
			} else if val != tt.wantVal {
				t.Errorf("got %v (%T), want %v", val, val, tt.wantVal)
			}
		})
	}
}

func TestAdviceChain_Execute_MatchingAdvisors(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("DoWork"), Advice: advice},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if !advice.called {
		t.Error("expected advisor to be called")
	}
}

func TestAdviceChain_Execute_NoMatchingAdvisors(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("OtherMethod"), Advice: advice},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if advice.called {
		t.Error("expected advisor NOT to be called")
	}
}

func TestAdviceChain_Execute_EmptyAdvisors(t *testing.T) {
	t.Parallel()
	chain := NewAdviceChain(nil)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
}

func TestAdviceChain_Execute_MultipleAdvisorsSomeMatch(t *testing.T) {
	t.Parallel()
	advice1 := &mockAdvice{}
	advice2 := &mockAdvice{}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("OtherMethod"), Advice: advice1},
		{Pointcut: aop.MatchByName("DoWork"), Advice: advice2},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if advice1.called {
		t.Error("expected advice1 NOT to be called")
	}
	if !advice2.called {
		t.Error("expected advice2 to be called")
	}
}

func TestAdviceChain_AdvisorShortCircuitsResult(t *testing.T) {
	t.Parallel()
	advice := &mockAdviceWithResult{result: "short-circuited", err: nil}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("DoWork"), Advice: advice},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "short-circuited" {
		t.Errorf("expected 'short-circuited', got %v", result)
	}
}

func TestAdviceChain_AdvisorReturnsError(t *testing.T) {
	t.Parallel()
	advice := &mockAdviceWithResult{result: nil, err: errors.New("advisor error")}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("DoWork"), Advice: advice},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	_, err := chain.Execute(target, method, nil)
	if err == nil || err.Error() != "advisor error" {
		t.Errorf("expected 'advisor error', got %v", err)
	}
}

func TestAdviceChain_Execute_WithArgs(t *testing.T) {
	t.Parallel()
	advice := &mockAdvice{}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("DoWork"), Advice: advice},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTargetWithArgs{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")
	args := []reflect.Value{reflect.ValueOf("test"), reflect.ValueOf(7)}

	result, err := chain.Execute(target, method, args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "test" {
		t.Errorf("expected 'test', got %v", result)
	}
}

func TestAdviceChain_WithOrder(t *testing.T) {
	t.Parallel()
	var order []string

	advice1 := &orderedAdvice{name: "first", order: 1, orderSlice: &order}
	advice2 := &orderedAdvice{name: "second", order: 2, orderSlice: &order}

	advisors := []Advisor{
		{Pointcut: aop.MatchByName("Do*"), Advice: advice1},
		{Pointcut: aop.MatchByName("Do*"), Advice: advice2},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}

	expected := []string{"first"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Errorf("order[%d] = %s, want %s", i, order[i], v)
		}
	}
}

func TestNewAdviceChain_NilAdvisors(t *testing.T) {
	t.Parallel()
	chain := NewAdviceChain(nil)
	if chain == nil {
		t.Error("expected non-nil chain")
	}
	if len(chain.advisors) != 0 {
		t.Errorf("expected 0 advisors, got %d", len(chain.advisors))
	}
}

func TestMethodInvocation_Target(t *testing.T) {
	t.Parallel()
	target := &mockTarget{}
	inv := &MethodInvocation{TargetObj: target}
	if inv.Target() != target {
		t.Error("expected Target to return the target object")
	}
}

func TestMethodInvocation_Method(t *testing.T) {
	t.Parallel()
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")
	inv := &MethodInvocation{MethodObj: method}
	if inv.Method() != "DoWork" {
		t.Errorf("expected 'DoWork', got %v", inv.Method())
	}
}

func TestAdviceChain_Execute_SkipsNonMatchingFollowedByMatch(t *testing.T) {
	t.Parallel()
	advice1 := &mockAdvice{}
	advice2 := &mockAdvice{}
	advisors := []Advisor{
		{Pointcut: aop.MatchByName("Foo*"), Advice: advice1},
		{Pointcut: aop.MatchByName("Bar*"), Advice: advice2},
		{Pointcut: aop.MatchByName("Do*"), Advice: &mockAdvice{}},
	}
	chain := NewAdviceChain(advisors)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
	if advice1.called {
		t.Error("expected advice1 NOT to be called")
	}
	if advice2.called {
		t.Error("expected advice2 NOT to be called")
	}
}

func TestAdviceChain_Execute_NilProceedNoArgs(t *testing.T) {
	t.Parallel()
	chain := NewAdviceChain(nil)
	target := &mockTarget{}
	method, _ := reflect.TypeOf(target).MethodByName("DoWork")

	result, err := chain.Execute(target, method, []reflect.Value{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "work" {
		t.Errorf("expected 'work', got %v", result)
	}
}

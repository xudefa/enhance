package aop

import (
	"context"
	"errors"
	"testing"
)

// testTypedService 测试用的类型化服务。
type testTypedService struct {
	Name string
}

func (s *testTypedService) DoWork() string {
	return "working: " + s.Name
}

// mockJoinPoint 测试用的 JoinPoint 模拟实现
type mockJoinPoint struct {
	name   string
	ctx    context.Context
	target any
	args   []any
	result any
	err    error
}

func (m *mockJoinPoint) Target() any                             { return m.target }
func (m *mockJoinPoint) Method() string                          { return m.name }
func (m *mockJoinPoint) Args() []any                             { return m.args }
func (m *mockJoinPoint) Proceed() (any, error)                   { return m.result, m.err }
func (m *mockJoinPoint) ProceedWithArgs(args []any) (any, error) { return m.result, m.err }
func (m *mockJoinPoint) Context() context.Context {
	if m.ctx != nil {
		return m.ctx
	}
	return context.Background()
}
func (m *mockJoinPoint) GetResult() any     { return m.result }
func (m *mockJoinPoint) GetError() error    { return m.err }
func (m *mockJoinPoint) SetResult(v any)    { m.result = v }
func (m *mockJoinPoint) SetError(err error) { m.err = err }

func TestClassifyAdvices(t *testing.T) {
	t.Parallel()

	aspects := []*AspectMeta{
		{
			Advice: NewBeforeAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
				return nil, nil
			}, 0),
		},
		{
			Advice: NewAfterAdvice(func(ctx context.Context, jp JoinPoint) (any, error) {
				return nil, nil
			}, 0),
		},
		{
			Advice: NewAroundAdvice(func(ctx context.Context, jp JoinPoint, proceed func() (any, error)) (any, error) {
				return proceed()
			}, 0),
		},
		{
			Advice: NewAfterReturningAdvice(func(ctx context.Context, jp JoinPoint, result any) (any, error) {
				return result, nil
			}, 0),
		},
		{
			Advice: NewAfterThrowingAdvice(func(ctx context.Context, jp JoinPoint, err error) (any, error) {
				return nil, err
			}, 0),
		},
	}

	ca := classifyAdvices(aspects)

	if len(ca.before) != 1 {
		t.Errorf("expected 1 before advice, got %d", len(ca.before))
	}
	if len(ca.after) != 1 {
		t.Errorf("expected 1 after advice, got %d", len(ca.after))
	}
	if len(ca.around) != 1 {
		t.Errorf("expected 1 around advice, got %d", len(ca.around))
	}
	if len(ca.afterReturning) != 1 {
		t.Errorf("expected 1 afterReturning advice, got %d", len(ca.afterReturning))
	}
	if len(ca.afterThrowing) != 1 {
		t.Errorf("expected 1 afterThrowing advice, got %d", len(ca.afterThrowing))
	}
}

func TestClassifyAdvices_Empty(t *testing.T) {
	t.Parallel()

	ca := classifyAdvices(nil)
	if ca == nil {
		t.Fatal("expected non-nil classifiedAdvices")
	}
	if len(ca.before) != 0 {
		t.Errorf("expected 0 before advices, got %d", len(ca.before))
	}
}

func TestClassifyAdvices_NilAspect(t *testing.T) {
	t.Parallel()

	aspects := []*AspectMeta{nil, {Advice: nil}}
	ca := classifyAdvices(aspects)
	if ca == nil {
		t.Fatal("expected non-nil classifiedAdvices")
	}
}

func TestNewChainExecutor(t *testing.T) {
	t.Parallel()

	executor := NewChainExecutor()
	if executor == nil {
		t.Fatal("expected non-nil executor")
	}

	executorWithOpts := NewChainExecutor(WithRecovery(), WithInterceptor(func(inv Invocation, next func(Invocation) any) any {
		return next(inv)
	}))
	if executorWithOpts == nil {
		t.Fatal("expected non-nil executor with options")
	}
}

func TestDefaultChainExecutor(t *testing.T) {
	t.Parallel()

	executor := DefaultChainExecutor()
	if executor == nil {
		t.Fatal("expected non-nil default executor")
	}
}

func TestSetDefaultChainExecutor(t *testing.T) {
	t.Parallel()

	original := DefaultChainExecutor()

	newExecutor := NewChainExecutor()
	SetDefaultChainExecutor(newExecutor)

	if DefaultChainExecutor() != newExecutor {
		t.Error("expected default executor to be updated")
	}

	// 恢复原始执行器
	SetDefaultChainExecutor(original)
}

func TestSetDefaultChainExecutor_Nil(t *testing.T) {
	t.Parallel()

	original := DefaultChainExecutor()

	// 设置 nil 应该被忽略
	SetDefaultChainExecutor(nil)

	if DefaultChainExecutor() != original {
		t.Error("expected default executor to remain unchanged")
	}
}

func TestPanicInfo_Error(t *testing.T) {
	t.Parallel()

	info := &PanicInfo{
		Value: "test panic",
		Stack: []byte("stack trace"),
	}

	errMsg := info.Error()
	if errMsg == "" {
		t.Error("expected non-empty error message")
	}
}

func TestExtractChainError_ErrorResult(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")
	result, err := extractChainError(testErr)
	if err != testErr {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestExtractChainError_NilError(t *testing.T) {
	t.Parallel()

	result, err := extractChainError(nil)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
}

func TestExtractChainError_MultiReturnWithError(t *testing.T) {
	t.Parallel()

	testErr := errors.New("multi return error")
	results := []any{"success", testErr}

	result, err := extractChainError(results)
	if err != testErr {
		t.Errorf("expected error %v, got %v", testErr, err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestExtractChainError_MultiReturnWithNilError(t *testing.T) {
	t.Parallel()

	results := []any{"success", nil}

	result, err := extractChainError(results)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Error("expected non-nil result")
	}
}

func TestExtractChainError_SuccessResult(t *testing.T) {
	t.Parallel()

	result, err := extractChainError("success")
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected result 'success', got %v", result)
	}
}

func TestChainStats(t *testing.T) {
	t.Parallel()

	// 重置统计信息
	GlobalChainStats.TotalExecutions.Store(0)
	GlobalChainStats.TotalPanics.Store(0)
	GlobalChainStats.TotalInterceptors.Store(0)

	updateStats(nil, 0)
	if GlobalChainStats.TotalExecutions.Load() != 1 {
		t.Errorf("expected 1 total execution, got %d", GlobalChainStats.TotalExecutions.Load())
	}

	updateStats("panic", 2)
	if GlobalChainStats.TotalPanics.Load() != 1 {
		t.Errorf("expected 1 total panic, got %d", GlobalChainStats.TotalPanics.Load())
	}
	if GlobalChainStats.TotalInterceptors.Load() != 2 {
		t.Errorf("expected 2 total interceptors, got %d", GlobalChainStats.TotalInterceptors.Load())
	}
}

func TestChainJoinPoint_ProceedWithArgs(t *testing.T) {
	t.Parallel()

	innerJP := &mockJoinPoint{
		name:   "TestMethod",
		ctx:    context.Background(),
		target: &testTypedService{Name: "test"},
	}

	inv := &invocationImpl{
		joinPoint: innerJP,
		args:      []any{"arg1", "arg2"},
	}

	cjp := &chainJoinPoint{
		inner: innerJP,
		inv:   inv,
		proceed: func() (any, error) {
			return "proceeded", nil
		},
	}

	newArgs := []any{"newArg1", "newArg2"}
	result, err := cjp.ProceedWithArgs(newArgs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "proceeded" {
		t.Errorf("expected result 'proceeded', got %v", result)
	}
}

func TestChainJoinPoint_Delegation(t *testing.T) {
	t.Parallel()

	innerJP := &mockJoinPoint{
		name:   "TestMethod",
		ctx:    context.Background(),
		target: &testTypedService{Name: "test"},
		result: "inner result",
		err:    errors.New("inner error"),
	}

	cjp := &chainJoinPoint{
		inner: innerJP,
		proceed: func() (any, error) {
			return "proceeded", nil
		},
	}

	if cjp.Target() != innerJP.Target() {
		t.Error("expected Target to delegate to inner")
	}
	if cjp.Method() != innerJP.Method() {
		t.Error("expected Method to delegate to inner")
	}
	if cjp.GetResult() != innerJP.GetResult() {
		t.Error("expected GetResult to delegate to inner")
	}
	if cjp.GetError() != innerJP.GetError() {
		t.Error("expected GetError to delegate to inner")
	}
}

func TestChainJoinPoint_SetResult(t *testing.T) {
	t.Parallel()

	innerJP := &mockJoinPoint{
		name: "TestMethod",
		ctx:  context.Background(),
	}

	cjp := &chainJoinPoint{
		inner: innerJP,
		proceed: func() (any, error) {
			return nil, nil
		},
	}

	cjp.SetResult("new result")
	if innerJP.GetResult() != "new result" {
		t.Errorf("expected inner result to be updated")
	}
}

func TestChainJoinPoint_SetError(t *testing.T) {
	t.Parallel()

	innerJP := &mockJoinPoint{
		name: "TestMethod",
		ctx:  context.Background(),
	}

	testErr := errors.New("test error")
	cjp := &chainJoinPoint{
		inner: innerJP,
		proceed: func() (any, error) {
			return nil, nil
		},
	}

	cjp.SetError(testErr)
	if innerJP.GetError() != testErr {
		t.Errorf("expected inner error to be updated")
	}
}

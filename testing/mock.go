package testing

import (
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// mockImpl Mock 接口的默认实现。
type mockImpl struct {
	mu           sync.RWMutex
	expectations []Expectation
	callCount    map[string]int
}

// Expectation 表示一个方法调用期望。
type Expectation struct {
	Method    string
	Args      []any
	Result    any
	Error     error
	Times     int
	CallCount int
}

// NewMock 创建一个新的 Mock 对象。
func NewMock() Mock {
	return &mockImpl{
		expectations: make([]Expectation, 0),
		callCount:    make(map[string]int),
	}
}

// Expect 设置方法调用期望，默认期望调用 1 次。
func (m *mockImpl) Expect(method string, args []any, result any, err error) Mock {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = append(m.expectations, Expectation{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Times:  1,
	})

	return m
}

// ExpectTimes 设置方法调用期望，指定期望调用次数。
func (m *mockImpl) ExpectTimes(method string, args []any, result any, err error, times int) Mock {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = append(m.expectations, Expectation{
		Method: method,
		Args:   args,
		Result: result,
		Error:  err,
		Times:  times,
	})

	return m
}

// Call 模拟方法调用，返回匹配的期望结果。
func (m *mockImpl) Call(method string, args ...any) (any, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := methodKey(method, args)
	m.callCount[key]++

	for i, exp := range m.expectations {
		if exp.Method == method && m.argsMatch(exp.Args, args) {
			if exp.CallCount >= exp.Times {
				continue
			}
			m.expectations[i].CallCount++
			return exp.Result, exp.Error
		}
	}

	return nil, fmt.Errorf("未预期的调用: %s，参数 %v", method, args)
}

// Verify 验证所有期望是否满足。
func (m *mockImpl) Verify() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, exp := range m.expectations {
		if exp.CallCount < exp.Times {
			return fmt.Errorf("期望 %s 被调用 %d 次，但实际被调用 %d 次",
				exp.Method, exp.Times, exp.CallCount)
		}
	}

	return nil
}

// Reset 重置 Mock 对象的所有状态。
func (m *mockImpl) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.expectations = make([]Expectation, 0)
	m.callCount = make(map[string]int)
}

// argsMatch 检查实际参数是否与期望参数匹配。
func (m *mockImpl) argsMatch(expected, actual []any) bool {
	if len(expected) != len(actual) {
		return false
	}

	for i := range expected {
		if !reflect.DeepEqual(expected[i], actual[i]) {
			return false
		}
	}

	return true
}

// methodKey 生成方法调用的唯一键。
func methodKey(method string, args []any) string {
	var sb strings.Builder
	sb.WriteString(method)
	for _, arg := range args {
		fmt.Fprintf(&sb, "_%v", arg)
	}
	return sb.String()
}

// MockRecorder Mock 记录器，用于链式设置期望。
type MockRecorder struct {
	mock Mock
}

// NewMockRecorder 创建 Mock 记录器。
func NewMockRecorder(mock Mock) *MockRecorder {
	return &MockRecorder{mock: mock}
}

// Return 设置返回值（链式调用）。
func (r *MockRecorder) Return(result any, err error) Mock {
	return r.mock
}

// Times 设置调用次数（链式调用）。
func (r *MockRecorder) Times(n int) Mock {
	return r.mock
}

// WithMock 使用 Mock 运行测试。
func WithMock(t *testing.T, fn func(ctx TestContext, mock *MockRecorder)) {
	t.Helper()
	ctx := NewTestContext(t)
	defer ctx.Cleanup()

	m := NewMock()
	mock := NewMockRecorder(m)
	fn(ctx, mock)
}

// AssertExpectations 断言 Mock 期望是否满足。
func AssertExpectations(t TestingT, mock Mock) bool {
	t.Helper()
	if err := mock.Verify(); err != nil {
		t.Errorf("mock verification failed: %v", err)
		return false
	}
	return true
}

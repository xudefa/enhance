package proxy

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
)

// ==================== 测试辅助类型 ====================

// TestService 测试服务接口
type TestService interface {
	DoSomething(arg string) (string, error)
	DoAnother(num int) (int, error)
}

// TestServiceImpl 测试服务实现
type TestServiceImpl struct {
	Name string
}

// DoSomething 实现 TestService 接口
func (s *TestServiceImpl) DoSomething(arg string) (string, error) {
	return fmt.Sprintf("%s: %s", s.Name, arg), nil
}

// DoAnother 实现 TestService 接口
func (s *TestServiceImpl) DoAnother(num int) (int, error) {
	return num * 2, nil
}

// TestStruct 测试结构体（用于 CglibProxy）
type TestStruct struct {
	Value int
}

// Add 测试方法
func (s *TestStruct) Add(num int) int {
	return s.Value + num
}

// Multiply 测试方法
func (s *TestStruct) Multiply(num int) int {
	return s.Value * num
}

// ==================== Mock InvocationHandler ====================

// MockHandler 模拟调用处理器
type MockHandler struct {
	CallCount  atomic.Int64
	LastMethod string
	LastArgs   []any
	ReturnVal  any
	ReturnErr  error
	mu         sync.Mutex
}

// Invoke 实现 InvocationHandler 接口
func (h *MockHandler) Invoke(target any, method string, args []any) (any, error) {
	h.CallCount.Add(1)
	h.mu.Lock()
	h.LastMethod = method
	h.LastArgs = args
	h.mu.Unlock()
	return h.ReturnVal, h.ReturnErr
}

// SpyHandler 间谍调用处理器（记录调用并转发）
type SpyHandler struct {
	mu    sync.Mutex
	Calls []CallRecord
}

// CallRecord 调用记录
type CallRecord struct {
	Method string
	Args   []any
}

// Invoke 实现 InvocationHandler 接口
func (h *SpyHandler) Invoke(target any, method string, args []any) (any, error) {
	h.mu.Lock()
	h.Calls = append(h.Calls, CallRecord{Method: method, Args: args})
	h.mu.Unlock()

	// 转发到实际方法
	switch method {
	case "DoSomething":
		return target.(*TestServiceImpl).DoSomething(args[0].(string))
	case "DoAnother":
		return target.(*TestServiceImpl).DoAnother(args[0].(int))
	case "Add":
		return target.(*TestStruct).Add(args[0].(int)), nil
	case "Multiply":
		return target.(*TestStruct).Multiply(args[0].(int)), nil
	default:
		return nil, fmt.Errorf("unknown method: %s", method)
	}
}

// ==================== JdkDynamicProxy 测试 ====================

func testIface() reflect.Type {
	return reflect.TypeOf((*TestService)(nil)).Elem()
}

func TestJdkDynamicProxy_New(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	if p == nil {
		t.Fatal("NewJdkDynamicProxy should return non-nil proxy")
	}
	if p.GetTarget() != svc {
		t.Error("GetTarget should return original target")
	}
	if p.GetHandler() != handler {
		t.Error("GetHandler should return original handler")
	}
}

func TestJdkDynamicProxy_NewPanics_NilTarget(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewJdkDynamicProxy should panic on nil target")
		}
	}()

	handler := &MockHandler{}
	NewJdkDynamicProxy(nil, testIface(), handler)
}

func TestJdkDynamicProxy_NewPanics_NilHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewJdkDynamicProxy should panic on nil handler")
		}
	}()

	var svc TestService = &TestServiceImpl{Name: "test"}
	NewJdkDynamicProxy(svc, testIface(), nil)
}

func TestJdkDynamicProxy_NewPanics_NonInterface(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewJdkDynamicProxy should panic on non-interface type")
		}
	}()

	handler := &MockHandler{}
	NewJdkDynamicProxy(&TestServiceImpl{Name: "test"}, reflect.TypeOf(&TestServiceImpl{}), handler)
}

func TestJdkDynamicProxy_Invoke(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{ReturnVal: "mocked", ReturnErr: nil}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	result, err := p.Invoke(svc, "DoSomething", []any{"hello"})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}
	if result != "mocked" {
		t.Errorf("Invoke result = %v, want 'mocked'", result)
	}
	if handler.CallCount.Load() != 1 {
		t.Errorf("handler.CallCount = %d, want 1", handler.CallCount.Load())
	}
	if handler.LastMethod != "DoSomething" {
		t.Errorf("handler.LastMethod = %s, want 'DoSomething'", handler.LastMethod)
	}
}

func TestJdkDynamicProxy_Invoke_WithArgs(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{ReturnVal: 42, ReturnErr: nil}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	result, err := p.Invoke(svc, "DoAnother", []any{21})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}
	if result != 42 {
		t.Errorf("Invoke result = %v, want 42", result)
	}
	if len(handler.LastArgs) != 1 || handler.LastArgs[0] != 21 {
		t.Errorf("handler.LastArgs = %v, want [21]", handler.LastArgs)
	}
}

func TestJdkDynamicProxy_Invoke_NilHandler(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	p := &JdkDynamicProxy{
		target:      svc,
		handler:     nil,
		iface:       testIface(),
		methodCache: make(map[string]reflect.Method),
	}

	_, err := p.Invoke(svc, "DoSomething", []any{"hello"})
	if err == nil {
		t.Error("Invoke with nil handler should return error")
	}
}

func TestJdkDynamicProxy_GetMethod(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	method, err := p.GetMethod("DoSomething")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}
	if method.Name != "DoSomething" {
		t.Errorf("GetMethod name = %s, want 'DoSomething'", method.Name)
	}
}

func TestJdkDynamicProxy_GetMethod_NotFound(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	_, err := p.GetMethod("NonExistent")
	if err == nil {
		t.Error("GetMethod should return error for non-existent method")
	}
}

func TestJdkDynamicProxy_GetMethod_Cache(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	// 第一次调用
	method1, err := p.GetMethod("DoSomething")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}

	// 第二次调用（应该从缓存获取）
	method2, err := p.GetMethod("DoSomething")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}

	if method1.Name != method2.Name {
		t.Error("Cached method should be the same")
	}
}

func TestJdkDynamicProxy_InvokeMethod(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	result, err := p.InvokeMethod("DoSomething", []any{"hello"})
	if err != nil {
		t.Fatalf("InvokeMethod should not error, got %v", err)
	}

	expected := "test: hello"
	if result != expected {
		t.Errorf("InvokeMethod result = %v, want %v", result, expected)
	}
}

func TestJdkDynamicProxy_InvokeMethod_NotFound(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	_, err := p.InvokeMethod("NonExistent", nil)
	if err == nil {
		t.Error("InvokeMethod should return error for non-existent method")
	}
}

func TestEnsureJdkDynamicProxy(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	proxy, ok := EnsureJdkDynamicProxy(p)
	if !ok {
		t.Error("EnsureJdkDynamicProxy should return true for JdkDynamicProxy")
	}
	if proxy != p {
		t.Error("EnsureJdkDynamicProxy should return the same proxy")
	}
}

func TestEnsureJdkDynamicProxy_WrongType(t *testing.T) {
	t.Parallel()

	_, ok := EnsureJdkDynamicProxy("not a proxy")
	if ok {
		t.Error("EnsureJdkDynamicProxy should return false for non-proxy")
	}
}

func TestIsJdkDynamicProxy(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	if !IsJdkDynamicProxy(p) {
		t.Error("IsJdkDynamicProxy should return true for JdkDynamicProxy")
	}
	if IsJdkDynamicProxy("not a proxy") {
		t.Error("IsJdkDynamicProxy should return false for non-proxy")
	}
}

func TestJdkDynamicProxy_GetIface(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	iface := p.GetIface()
	if iface == nil {
		t.Error("GetIface should return non-nil")
	}
	if iface.Kind() != reflect.Interface {
		t.Errorf("GetIface should return interface type, got %s", iface.Kind())
	}
}

// ==================== CglibProxy 测试 ====================

func TestCglibProxy_New(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	if p == nil {
		t.Fatal("NewCglibProxy should return non-nil proxy")
	}
	if p.GetTarget() != target {
		t.Error("GetTarget should return original target")
	}
	if p.GetHandler() != handler {
		t.Error("GetHandler should return original handler")
	}
}

func TestCglibProxy_NewPanics_NilTarget(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCglibProxy should panic on nil target")
		}
	}()

	handler := &MockHandler{}
	NewCglibProxy(nil, handler)
}

func TestCglibProxy_NewPanics_NilHandler(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCglibProxy should panic on nil handler")
		}
	}()

	target := &TestStruct{Value: 10}
	NewCglibProxy(target, nil)
}

func TestCglibProxy_NewPanics_NonStruct(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("NewCglibProxy should panic on non-struct target")
		}
	}()

	handler := &MockHandler{}
	NewCglibProxy("not a struct", handler)
}

func TestCglibProxy_Invoke(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{ReturnVal: 20, ReturnErr: nil}
	p := NewCglibProxy(target, handler)

	result, err := p.Invoke(target, "Add", []any{5})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}
	if result != 20 {
		t.Errorf("Invoke result = %v, want 20", result)
	}
	if handler.CallCount.Load() != 1 {
		t.Errorf("handler.CallCount = %d, want 1", handler.CallCount.Load())
	}
}

func TestCglibProxy_Invoke_WithArgs(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{ReturnVal: 50, ReturnErr: nil}
	p := NewCglibProxy(target, handler)

	result, err := p.Invoke(target, "Multiply", []any{5})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}
	if result != 50 {
		t.Errorf("Invoke result = %v, want 50", result)
	}
	if handler.LastMethod != "Multiply" {
		t.Errorf("handler.LastMethod = %s, want 'Multiply'", handler.LastMethod)
	}
}

func TestCglibProxy_Invoke_NilHandler(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	p := &CglibProxy{
		target:      target,
		handler:     nil,
		targetType:  reflect.TypeOf(target).Elem(),
		methodCache: make(map[string]reflect.Method),
	}

	_, err := p.Invoke(target, "Add", []any{5})
	if err == nil {
		t.Error("Invoke with nil handler should return error")
	}
}

func TestCglibProxy_GetMethod(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	method, err := p.GetMethod("Add")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}
	if method.Name != "Add" {
		t.Errorf("GetMethod name = %s, want 'Add'", method.Name)
	}
}

func TestCglibProxy_GetMethod_NotFound(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	_, err := p.GetMethod("NonExistent")
	if err == nil {
		t.Error("GetMethod should return error for non-existent method")
	}
}

func TestCglibProxy_GetMethod_Cache(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	// 第一次调用
	method1, err := p.GetMethod("Add")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}

	// 第二次调用（应该从缓存获取）
	method2, err := p.GetMethod("Add")
	if err != nil {
		t.Fatalf("GetMethod should not error, got %v", err)
	}

	if method1.Name != method2.Name {
		t.Error("Cached method should be the same")
	}
}

func TestCglibProxy_InvokeMethod(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	result, err := p.InvokeMethod("Add", []any{5})
	if err != nil {
		t.Fatalf("InvokeMethod should not error, got %v", err)
	}

	if result != 15 {
		t.Errorf("InvokeMethod result = %v, want 15", result)
	}
}

func TestCglibProxy_InvokeMethod_NotFound(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	_, err := p.InvokeMethod("NonExistent", nil)
	if err == nil {
		t.Error("InvokeMethod should return error for non-existent method")
	}
}

func TestEnsureCglibProxy(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	proxy, ok := EnsureCglibProxy(p)
	if !ok {
		t.Error("EnsureCglibProxy should return true for CglibProxy")
	}
	if proxy != p {
		t.Error("EnsureCglibProxy should return the same proxy")
	}
}

func TestEnsureCglibProxy_WrongType(t *testing.T) {
	t.Parallel()

	_, ok := EnsureCglibProxy("not a proxy")
	if ok {
		t.Error("EnsureCglibProxy should return false for non-proxy")
	}
}

func TestIsCglibProxy(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	if !IsCglibProxy(p) {
		t.Error("IsCglibProxy should return true for CglibProxy")
	}
	if IsCglibProxy("not a proxy") {
		t.Error("IsCglibProxy should return false for non-proxy")
	}
}

func TestCglibProxy_GetTargetType(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{}
	p := NewCglibProxy(target, handler)

	targetType := p.GetTargetType()
	if targetType == nil {
		t.Error("GetTargetType should return non-nil")
	}
	if targetType.Name() != "TestStruct" {
		t.Errorf("GetTargetType name = %s, want 'TestStruct'", targetType.Name())
	}
}

// ==================== SpyHandler 测试 ====================

func TestSpyHandler_JdkDynamicProxy(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	spy := &SpyHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), spy)

	result, err := p.Invoke(svc, "DoSomething", []any{"hello"})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}

	expected := "test: hello"
	if result != expected {
		t.Errorf("Invoke result = %v, want %v", result, expected)
	}
	if len(spy.Calls) != 1 {
		t.Errorf("spy.Calls length = %d, want 1", len(spy.Calls))
	}
	if spy.Calls[0].Method != "DoSomething" {
		t.Errorf("spy.Calls[0].Method = %s, want 'DoSomething'", spy.Calls[0].Method)
	}
}

func TestSpyHandler_CglibProxy(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	spy := &SpyHandler{}
	p := NewCglibProxy(target, spy)

	result, err := p.Invoke(target, "Add", []any{5})
	if err != nil {
		t.Fatalf("Invoke should not error, got %v", err)
	}

	if result != 15 {
		t.Errorf("Invoke result = %v, want 15", result)
	}
	if len(spy.Calls) != 1 {
		t.Errorf("spy.Calls length = %d, want 1", len(spy.Calls))
	}
}

// ==================== 并发测试 ====================

func TestJdkDynamicProxy_Concurrent(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	handler := &MockHandler{ReturnVal: "ok", ReturnErr: nil}
	p := NewJdkDynamicProxy(svc, testIface(), handler)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				_, _ = p.Invoke(svc, "DoSomething", []any{"test"})
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if handler.CallCount.Load() != 1000 {
		t.Errorf("handler.CallCount = %d, want 1000", handler.CallCount.Load())
	}
}

func TestCglibProxy_Concurrent(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	handler := &MockHandler{ReturnVal: 20, ReturnErr: nil}
	p := NewCglibProxy(target, handler)

	done := make(chan struct{})
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- struct{}{} }()
			for j := 0; j < 100; j++ {
				_, _ = p.Invoke(target, "Add", []any{5})
			}
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	if handler.CallCount.Load() != 1000 {
		t.Errorf("handler.CallCount = %d, want 1000", handler.CallCount.Load())
	}
}

// ==================== 表驱动测试 ====================

func TestJdkDynamicProxy_Methods(t *testing.T) {
	t.Parallel()

	var svc TestService = &TestServiceImpl{Name: "test"}
	spy := &SpyHandler{}
	p := NewJdkDynamicProxy(svc, testIface(), spy)

	tests := []struct {
		name    string
		method  string
		args    []any
		want    any
		wantErr bool
	}{
		{
			name:    "DoSomething success",
			method:  "DoSomething",
			args:    []any{"hello"},
			want:    "test: hello",
			wantErr: false,
		},
		{
			name:    "DoAnother success",
			method:  "DoAnother",
			args:    []any{21},
			want:    42,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := p.Invoke(svc, tt.method, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("Invoke() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestCglibProxy_Methods(t *testing.T) {
	t.Parallel()

	target := &TestStruct{Value: 10}
	spy := &SpyHandler{}
	p := NewCglibProxy(target, spy)

	tests := []struct {
		name    string
		method  string
		args    []any
		want    any
		wantErr bool
	}{
		{
			name:    "Add success",
			method:  "Add",
			args:    []any{5},
			want:    15,
			wantErr: false,
		},
		{
			name:    "Multiply success",
			method:  "Multiply",
			args:    []any{5},
			want:    50,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result, err := p.Invoke(target, tt.method, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Invoke() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.want {
				t.Errorf("Invoke() = %v, want %v", result, tt.want)
			}
		})
	}
}

package aop

import (
	"reflect"
	"testing"
)

func TestInterfaceProxyWrapper_GetTarget(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	target := &covWrapTarget{}
	wrapper := NewInterfaceProxyWrapper(target, nil, iface)

	if wrapper.GetTarget() != target {
		t.Error("GetTarget() should return original target")
	}
}

func TestInterfaceProxyWrapper_GetAdvisors(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	advisors := []*AspectMeta{{Order: 1}, {Order: 2}}
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, advisors, iface)

	got := wrapper.GetAdvisors()
	if len(got) != 2 {
		t.Errorf("GetAdvisors() length = %d, want 2", len(got))
	}
}

func TestInterfaceProxyWrapper_AdvisorsCopied(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	original := []*AspectMeta{{Order: 1}}
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, original, iface)

	// Modify original slice
	original[0] = &AspectMeta{Order: 999}

	got := wrapper.GetAdvisors()
	if got[0].Order != 1 {
		t.Error("advisors should be copied, not referenced")
	}
}

func TestInterfaceProxyWrapper_Invoke_NonExistentMethod(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, nil, iface)

	_, err := wrapper.Invoke("NonExistentMethod")
	if err == nil {
		t.Error("expected error for non-existent method")
	}
}

func TestInterfaceProxyWrapper_GetMethod_Cache(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, nil, iface)

	// First call - cache miss
	m1, err := wrapper.getMethod("DoCalc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call - cache hit
	m2, err := wrapper.getMethod("DoCalc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if m1.Name != m2.Name {
		t.Error("cached method should be the same")
	}

	// Verify cache contains the method
	wrapper.cacheMu.RLock()
	cached, ok := wrapper.methodCache["DoCalc"]
	wrapper.cacheMu.RUnlock()
	if !ok || cached.Name != "DoCalc" {
		t.Error("method should be in cache")
	}
}

func TestExtractResult_NilResult(t *testing.T) {
	t.Parallel()

	result, err := extractResult(nil)
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestExtractResult_EmptySlice(t *testing.T) {
	t.Parallel()

	result, err := extractResult([]any{})
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestExtractResult_SingleValue(t *testing.T) {
	t.Parallel()

	result, err := extractResult("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got %v", result)
	}
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestExtractResult_SingleError(t *testing.T) {
	t.Parallel()

	sentinel := &testErr{msg: "fail"}
	result, err := extractResult(sentinel)
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err != sentinel {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestExtractResult_ValueAndError(t *testing.T) {
	t.Parallel()

	sentinel := &testErr{msg: "fail"}
	result, err := extractResult([]any{"value", sentinel})
	if result != "value" {
		t.Errorf("expected 'value', got %v", result)
	}
	if err != sentinel {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestExtractResult_MultipleValuesNoError(t *testing.T) {
	t.Parallel()

	result, err := extractResult([]any{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	list, ok := result.([]any)
	if !ok || len(list) != 3 {
		t.Errorf("expected []any{1,2,3}, got %v", result)
	}
}

func TestExtractResult_SoleErrorInSlice(t *testing.T) {
	t.Parallel()

	sentinel := &testErr{msg: "only error"}
	result, err := extractResult([]any{sentinel})
	if result != nil {
		t.Errorf("expected nil result, got %v", result)
	}
	if err != sentinel {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

type testErr struct {
	msg string
}

func (e *testErr) Error() string { return e.msg }

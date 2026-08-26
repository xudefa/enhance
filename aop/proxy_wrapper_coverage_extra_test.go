package aop

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type covWrapIface interface {
	DoCalc(n int) int
	Greet(name string) (string, error)
	Fail() error
	Nothing()
}

type covWrapTarget struct {
	lastArg int
}

func (*covWrapTarget) DoCalc(n int) int { return n * 2 }

func (*covWrapTarget) Greet(name string) (string, error) {
	if name == "" {
		return "", errors.New("empty name")
	}
	return "hi " + name, nil
}

func (*covWrapTarget) Fail() error { return errors.New("boom") }

func (*covWrapTarget) Nothing() {}

type covWrapPartial struct{}

func (*covWrapPartial) DoCalc(n int) int { return n }

func TestExtractResult_Table(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("sentinel")

	tests := []struct {
		name      string
		in        any
		want      any
		wantErr   bool
		errTarget error
	}{
		{"nil result", nil, nil, false, nil},
		{"empty slice", []any{}, nil, false, nil},
		{"sole error in slice", []any{sentinel}, nil, true, sentinel},
		{"value plus error", []any{7, sentinel}, 7, true, sentinel},
		{
			"multiple values with trailing error",
			[]any{1, "two", sentinel},
			[]any{1, "two"},
			true,
			sentinel,
		},
		{"single value no error", []any{"only"}, "only", false, nil},
		{"multiple values no error", []any{1, 2}, []any{1, 2}, false, nil},
		{"bare error result", error(sentinel), nil, true, sentinel},
		{"plain value", 42, 42, false, nil},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := extractResult(tt.in)
			if tt.wantErr {
				if !errors.Is(err, tt.errTarget) {
					t.Fatalf("expected error %v, got %v", tt.errTarget, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			switch want := tt.want.(type) {
			case []any:
				gotList, ok := got.([]any)
				if !ok || len(gotList) != len(want) {
					t.Fatalf("got %v, want list %v", got, want)
				}
				for i := range want {
					if gotList[i] != want[i] {
						t.Errorf("item %d: got %v want %v", i, gotList[i], want[i])
					}
				}
			default:
				if got != tt.want {
					t.Fatalf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestInterfaceProxyWrapper_InvokeContext_Table(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	target := &covWrapTarget{}

	tests := []struct {
		name       string
		method     string
		args       []any
		want       any
		wantErr    bool
		errMessage string
	}{
		{"single return", "DoCalc", []any{21}, 42, false, ""},
		{
			"value and error ok",
			"Greet",
			[]any{"bob"},
			[]any{"hi bob", nil},
			false,
			"",
		},
		{"value and error fail", "Greet", []any{""}, nil, true, "empty name"},
		{"sole error return", "Fail", nil, nil, true, "boom"},
		{"no return values", "Nothing", nil, nil, false, ""},
		{
			"unknown method on interface",
			"Missing",
			nil,
			nil,
			true,
			"method Missing not found on interface covWrapIface",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			wrapper := NewInterfaceProxyWrapper(target, nil, iface)
			got, err := wrapper.InvokeContext(context.Background(), tt.method, tt.args...)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil (result %v)", tt.errMessage, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("got %v (%T), want %v", got, got, tt.want)
			}
		})
	}
}

func TestInterfaceProxyWrapper_Invoke_Convenience(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, nil, iface)

	got, err := wrapper.Invoke("DoCalc", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 10 {
		t.Fatalf("got %v, want 10", got)
	}
}

func TestInterfaceProxyWrapper_CustomExecutorUsed(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, nil, iface)
	wrapper.SetExecutor(&mockChainExecutorForProxyTest{})

	result, err := wrapper.InvokeContext(context.Background(), "DoCalc", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Errorf("mock executor returns nil, got %v", result)
	}
}

func TestInterfaceProxyWrapper_MethodCacheHit(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, nil, iface)

	for range 2 {
		got, err := wrapper.Invoke("DoCalc", 4)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != 8 {
			t.Fatalf("got %v, want 8", got)
		}
	}

	wrapper.cacheMu.RLock()
	cached, ok := wrapper.methodCache["DoCalc"]
	wrapper.cacheMu.RUnlock()
	if !ok || cached.Name != "DoCalc" {
		t.Error("expected DoCalc to be cached after repeated calls")
	}
}

func TestInterfaceProxyWrapper_TargetMethodPanic(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	wrapper := NewInterfaceProxyWrapper(&covWrapPartial{}, nil, iface)

	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic when target lacks interface method")
		}
		info, ok := r.(*PanicInfo)
		if !ok {
			t.Fatalf("expected PanicInfo, got %T: %v", r, r)
		}
		msg, ok := info.Value.(string)
		if !ok || msg != "method Nothing not found on target *aop.covWrapPartial" {
			t.Fatalf("unexpected panic value: %v", info.Value)
		}
	}()

	_, _ = wrapper.Invoke("Nothing")
}

func TestInterfaceProxyWrapper_AdvisorSeesError(t *testing.T) {
	t.Parallel()

	iface := reflect.TypeOf((*covWrapIface)(nil)).Elem()
	threw := false
	advisors := []*AspectMeta{{
		PointCut: MatchByName("Fail"),
		Advice:   AfterThrowing(func(jp JoinPoint, err error) { threw = true }),
	}}
	wrapper := NewInterfaceProxyWrapper(&covWrapTarget{}, advisors, iface)

	_, err := wrapper.Invoke("Fail")
	if err == nil {
		t.Fatal("expected error from Fail")
	}
	if !threw {
		t.Error("expected AfterThrowing advisor to run")
	}
}

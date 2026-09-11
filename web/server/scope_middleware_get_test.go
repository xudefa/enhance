package server

import (
	"context"
	"sync"
	"testing"
)

func TestGetRequestScope_NilCtx(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Error("GetRequestScope(nil) should panic")
		}
	}()
	GetRequestScope(nil)
}

func TestGetRequestScope_Background(t *testing.T) {
	t.Parallel()
	scope := GetRequestScope(context.Background())
	if scope != nil {
		t.Error("expected nil for background context")
	}
}

func TestGetRequestScope_WithScope(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, scope)

	got := GetRequestScope(ctx)
	if got == nil {
		t.Fatal("expected scope from context")
	}
	if got != scope {
		t.Error("should return the same scope instance")
	}
}

func TestGetRequestScope_WrongType(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, "not a scope")

	got := GetRequestScope(ctx)
	if got != nil {
		t.Error("expected nil when context value has wrong type")
	}
}

func TestMustGetRequestScope_Panic(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRequestScope should panic when scope is missing")
		}
	}()

	MustGetRequestScope(context.Background())
}

func TestMustGetRequestScope_Success(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	ctx := context.WithValue(context.Background(), ScopeContextKey{}, scope)
	got := MustGetRequestScope(ctx)
	if got != scope {
		t.Error("should return the same scope")
	}
}

func TestMustGetRequestScope_WrongType(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustGetRequestScope should panic when context value has wrong type")
		}
	}()

	ctx := context.WithValue(context.Background(), ScopeContextKey{}, "not a scope")
	MustGetRequestScope(ctx)
}

func TestRequestScope_Get_DoubleCheckRace(t *testing.T) {
	t.Parallel()

	scope := NewRequestScope()
	var wg sync.WaitGroup
	var start sync.WaitGroup
	start.Add(1)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			start.Wait()
			result := scope.Get("key", func() any {
				return id
			})
			if result == nil {
				t.Errorf("result should not be nil")
			}
		}(i)
	}

	start.Done()
	wg.Wait()
}

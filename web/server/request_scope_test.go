package server

import (
	"sync"
	"testing"
)

func TestRequestScope_New(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	if scope == nil {
		t.Fatal("NewRequestScope returned nil")
	}
}

func TestRequestScope_Set_Get(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	scope.Set("key", "value")
	val := scope.Get("key", func() any { return "default" })
	if val != "value" {
		t.Errorf("expected value, got %v", val)
	}
}

func TestRequestScope_GetDefault(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	val := scope.Get("missing", func() any { return "default" })
	if val != "default" {
		t.Errorf("expected default, got %v", val)
	}
}

func TestRequestScope_Overwrite(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	scope.Set("key", "first")
	val := scope.Get("key", func() any { return "second" })
	if val != "first" {
		t.Errorf("expected first, got %v", val)
	}
}

func TestRequestScope_ClearV2(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	scope.Set("key", "value")
	scope.Clear()

	val := scope.Get("key", func() any { return "default" })
	if val != "default" {
		t.Errorf("after Clear, expected default, got %v", val)
	}
}

func TestRequestScope_ConcurrentV2(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "key"
			scope.Set(key, n)
			_ = scope.Get(key, func() any { return 0 })
		}(i)
	}
	wg.Wait()
}

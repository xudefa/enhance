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
	fetched := scope.Get("key", func() any { return "default" })
	if fetched != "value" {
		t.Errorf("expected value, got %v", fetched)
	}
}

func TestRequestScope_GetDefault(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	fetched := scope.Get("missing", func() any { return "default" })
	if fetched != "default" {
		t.Errorf("expected default, got %v", fetched)
	}
}

func TestRequestScope_Overwrite(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	scope.Set("key", "first")
	fetched := scope.Get("key", func() any { return "second" })
	if fetched != "first" {
		t.Errorf("expected first, got %v", fetched)
	}
}

func TestRequestScope_ClearV2(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()

	scope.Set("key", "value")
	scope.Clear()

	fetched := scope.Get("key", func() any { return "default" })
	if fetched != "default" {
		t.Errorf("after Clear, expected default, got %v", fetched)
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

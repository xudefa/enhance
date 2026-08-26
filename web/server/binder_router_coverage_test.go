package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/xudefa/enhance/web/core"
)

func TestFormBinder_BindJSON_NotPointer(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	body := strings.NewReader(`{"name":"test"}`)
	req := httptest.NewRequest("POST", "/api", body)
	var form testForm
	err := binder.BindJSON(req, form)
	if !errors.Is(err, ErrNotPointer) {
		t.Errorf("BindJSON() with non-pointer should return ErrNotPointer, got %v", err)
	}
}

func TestFormBinder_BindJSON_TrailingWhitespace(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	body := strings.NewReader(`{"name":"Bob"}  `)
	req := httptest.NewRequest("POST", "/api", body)
	form := &testForm{}
	err := binder.BindJSON(req, form)
	_ = err
}

func TestFormBinder_SetFieldValue_InvalidUint(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type uintForm struct {
		Port uint16 `form:"port"`
	}
	req := httptest.NewRequest("GET", "/?port=notanumber", nil)
	form := &uintForm{}
	err := binder.BindQuery(req, form)
	if err == nil {
		t.Error("expected error for invalid uint value")
	}
}

func TestFormBinder_SetFieldValue_InvalidBool(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type boolForm struct {
		Flag bool `form:"flag"`
	}
	req := httptest.NewRequest("GET", "/?flag=maybe", nil)
	form := &boolForm{}
	err := binder.BindQuery(req, form)
	if err == nil {
		t.Error("expected error for invalid bool value")
	}
}

func TestFormBinder_SetFieldValue_InvalidFloat(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type floatForm struct {
		Price float64 `form:"price"`
	}
	req := httptest.NewRequest("GET", "/?price=notanumber", nil)
	form := &floatForm{}
	err := binder.BindQuery(req, form)
	if err == nil {
		t.Error("expected error for invalid float value")
	}
}

func TestFormBinder_SetFieldValue_InvalidInt(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type intForm struct {
		Count int `form:"count"`
	}
	req := httptest.NewRequest("GET", "/?count=notanumber", nil)
	form := &intForm{}
	err := binder.BindQuery(req, form)
	if err == nil {
		t.Error("expected error for invalid int value")
	}
}

func TestFormBinder_SetFieldValue_PointerField(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type ptrForm struct {
		Name *string `form:"name"`
	}
	req := httptest.NewRequest("GET", "/?name=alice", nil)
	form := &ptrForm{}
	err := binder.BindQuery(req, form)
	if err != nil {
		t.Fatalf("BindQuery() error = %v", err)
	}
	if form.Name == nil || *form.Name != "alice" {
		t.Errorf("Name = %v, want pointer to alice", form.Name)
	}
}

func TestFormBinder_Bind_SkippedTag(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type skipForm struct {
		Name string `form:"-"`
		Age  int    `form:"age"`
	}
	req := httptest.NewRequest("GET", "/?name=alice&age=30", nil)
	form := &skipForm{}
	err := binder.BindQuery(req, form)
	if err != nil {
		t.Fatalf("BindQuery() error = %v", err)
	}
	if form.Name != "" {
		t.Errorf("Name should be empty (tag is '-'), got %q", form.Name)
	}
	if form.Age != 30 {
		t.Errorf("Age = %d, want 30", form.Age)
	}
}

func TestFormBinder_BindUintTypes(t *testing.T) {
	t.Parallel()
	binder := NewFormBinder()
	type uintForm struct {
		Port uint16 `form:"port"`
	}
	req := httptest.NewRequest("GET", "/?port=8080", nil)
	form := &uintForm{}
	err := binder.BindQuery(req, form)
	if err != nil {
		t.Fatalf("BindQuery() error = %v", err)
	}
	if form.Port != 8080 {
		t.Errorf("Port = %d, want 8080", form.Port)
	}
}

func TestRequestScope_ConcurrentGetWithCreation(t *testing.T) {
	t.Parallel()
	scope := NewRequestScope()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			scope.Get("shared", func() any {
				return fmt.Sprintf("value-%d", id)
			})
		}(i)
	}
	wg.Wait()

	val := scope.Get("shared", func() any {
		return "should-not-be-called"
	})
	if val == nil {
		t.Error("shared key should have a value")
	}
}

func TestHTTPServerBuilder_MustBuild_Panic(t *testing.T) {
	t.Parallel()

	srv := NewHTTPServerBuilder().MustBuild()
	if srv == nil {
		t.Fatal("MustBuild() returned nil")
	}
}

func TestRouter_Handle_AnyMethod(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	handlerCalled := false

	router.Handle("PROPFIND", "/files", func(ctx core.Context) {
		handlerCalled = true
	})

	req := httptest.NewRequest("PROPFIND", "/files", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if !handlerCalled {
		t.Error("PROPFIND handler was not called")
	}
}

func TestRouter_Group_PrefixNotMatching(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	_ = router.Group("/api")

	req := httptest.NewRequest(http.MethodGet, "/other/path", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestRouter_ServeHTTP_PrefixPathDoesNotMatchBoundary(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	api := router.Group("/api")
	api.GET("/users", func(ctx core.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/apix/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d for /apix/users", w.Code, http.StatusNotFound)
	}
}

func TestRouter_PrefixEmptyPath(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	api := router.Group("/api")
	api.GET("/", func(ctx core.Context) {
		ctx.String(http.StatusOK, "root")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d for /api/", w.Code, http.StatusOK)
	}
}

func TestRouter_PrefixExactMatch(t *testing.T) {
	t.Parallel()
	router := NewRouter()
	api := router.Group("/api")
	api.GET("/resource", func(ctx core.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/api/resource", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d for /api/resource", w.Code, http.StatusOK)
	}
}

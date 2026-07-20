package webtest

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestNewWebTestClient(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	if client == nil {
		t.Fatal("NewWebTestClient should return non-nil client")
	}
}

func TestWebTestClient_Get(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	client.Get("/api/test").Exchange().StatusIsOk()
}

func TestWebTestClient_Post(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
	})

	client := NewWebTestClient(handler)
	client.Post("/api/test").Exchange().StatusIsCreated()
}

func TestWebTestClient_Put(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	client.Put("/api/test").Exchange().StatusIsOk()
}

func TestWebTestClient_Delete(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	})

	client := NewWebTestClient(handler)
	client.Delete("/api/test").Exchange().StatusIsNoContent()
}

func TestWebTestClient_Patch(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	client.Patch("/api/test").Exchange().StatusIsOk()
}

func TestWebTestClient_BaseURL(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/test" {
			t.Errorf("expected path /api/v1/test, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler).BaseURL("/api/v1")
	client.Get("/test").Exchange().StatusIsOk()
}

func TestWebTestClient_Headers(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom-Header") != "custom-value" {
			t.Errorf("expected header X-Custom-Header=custom-value, got %s", r.Header.Get("X-Custom-Header"))
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	client.Get("/test").
		Header("X-Custom-Header", "custom-value").
		Exchange().
		StatusIsOk()
}

func TestWebTestClient_JSONBody(t *testing.T) {
	t.Parallel()
	type Request struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Request
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if req.Name != "Alice" || req.Age != 30 {
			t.Errorf("expected Alice/30, got %s/%d", req.Name, req.Age)
		}
		w.WriteHeader(http.StatusOK)
	})

	client := NewWebTestClient(handler)
	client.Post("/api/test").
		JSON(Request{Name: "Alice", Age: 30}).
		Exchange().
		StatusIsOk()
}

func TestWebTestClient_ResponseBody(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})

	client := NewWebTestClient(handler)
	response := client.Get("/test").Exchange()
	response.StatusIsOk()

	if response.Body() != "Hello, World!" {
		t.Errorf("expected 'Hello, World!', got %s", response.Body())
	}
}

func TestWebTestClient_JSONResponse(t *testing.T) {
	t.Parallel()
	type Response struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(Response{Name: "Alice", Age: 30})
	})

	client := NewWebTestClient(handler)
	var result Response
	client.Get("/api/test").
		Exchange().
		StatusIsOk().
		Header("Content-Type", "application/json").
		JSONBody(&result)

	if result.Name != "Alice" || result.Age != 30 {
		t.Errorf("expected Alice/30, got %s/%d", result.Name, result.Age)
	}
}

func TestWebTestClient_BodyContains(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Hello, World!"))
	})

	client := NewWebTestClient(handler)
	client.Get("/test").
		Exchange().
		StatusIsOk().
		BodyContains("World")
}

func TestWebTestClient_BodyEquals(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("exact content"))
	})

	client := NewWebTestClient(handler)
	client.Get("/test").
		Exchange().
		StatusIsOk().
		BodyEquals("exact content")
}

func TestWebTestClient_StatusAssertions(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		statusCode int
		assertion  func(*ResponseSpec) *ResponseSpec
	}{
		{"StatusIsOk", http.StatusOK, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsOk() }},
		{"StatusIsCreated", http.StatusCreated, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsCreated() }},
		{"StatusIsNoContent", http.StatusNoContent, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsNoContent() }},
		{"StatusIsBadRequest", http.StatusBadRequest, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsBadRequest() }},
		{"StatusIsUnauthorized", http.StatusUnauthorized, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsUnauthorized() }},
		{"StatusIsForbidden", http.StatusForbidden, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsForbidden() }},
		{"StatusIsNotFound", http.StatusNotFound, func(r *ResponseSpec) *ResponseSpec { return r.StatusIsNotFound() }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			})

			client := NewWebTestClient(handler)
			tt.assertion(client.Get("/test").Exchange())
		})
	}
}

func TestWebTestClient_StatusAssertionFailure(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	client := NewWebTestClient(handler)
	defer func() {
		if r := recover(); r == nil {
			t.Error("StatusIsOk should panic on 404")
		}
	}()

	client.Get("/test").Exchange().StatusIsOk()
}

func TestWebTestClient_Recorder(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test"))
	})

	client := NewWebTestClient(handler)
	response := client.Get("/test").Exchange()

	recorder := response.Recorder()
	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}
}

func TestCreateWebTestClient(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	client := CreateWebTestClient(handler)
	if client == nil {
		t.Fatal("CreateWebTestClient should return non-nil client")
	}
}

func TestWebTestClient_ChainedCalls(t *testing.T) {
	t.Parallel()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/users":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("users"))
		case "/posts":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("posts"))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	})

	client := NewWebTestClient(handler)

	client.Get("/users").Exchange().StatusIsOk().BodyContains("users")
	client.Get("/posts").Exchange().StatusIsOk().BodyContains("posts")
	client.Get("/unknown").Exchange().StatusIsNotFound()
}

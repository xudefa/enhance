package server

import (
	"testing"
)

func TestHTTPResponse_IsSuccess_Helper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{200, true},
		{201, true},
		{299, true},
		{199, false},
		{300, false},
		{404, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(httpStatusNameHelper(tt.code), func(t *testing.T) {
			t.Parallel()
			r := &HTTPResponse{StatusCode: tt.code}
			if got := r.IsSuccess(); got != tt.want {
				t.Errorf("IsSuccess() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsRedirect_Helper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{301, true},
		{302, true},
		{399, true},
		{299, false},
		{400, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(httpStatusNameHelper(tt.code), func(t *testing.T) {
			t.Parallel()
			r := &HTTPResponse{StatusCode: tt.code}
			if got := r.IsRedirect(); got != tt.want {
				t.Errorf("IsRedirect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsClientError_Helper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{400, true},
		{404, true},
		{499, true},
		{399, false},
		{500, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(httpStatusNameHelper(tt.code), func(t *testing.T) {
			t.Parallel()
			r := &HTTPResponse{StatusCode: tt.code}
			if got := r.IsClientError(); got != tt.want {
				t.Errorf("IsClientError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_IsServerError_Helper(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{500, true},
		{502, true},
		{599, true},
		{499, false},
		{600, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(httpStatusNameHelper(tt.code), func(t *testing.T) {
			t.Parallel()
			r := &HTTPResponse{StatusCode: tt.code}
			if got := r.IsServerError(); got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPResponse_Bind_Empty_Helper(t *testing.T) {
	t.Parallel()
	r := &HTTPResponse{}
	var v map[string]string
	err := r.Bind(&v)
	if err != nil {
		t.Errorf("Bind on empty body should not error, got %v", err)
	}
}

func TestHTTPResponse_Bind_JSON_Helper(t *testing.T) {
	t.Parallel()
	type User struct {
		Name string `json:"name"`
	}
	r := &HTTPResponse{Body: []byte(`{"name":"Alice"}`)}
	var u User
	err := r.Bind(&u)
	if err != nil {
		t.Fatalf("Bind error: %v", err)
	}
	if u.Name != "Alice" {
		t.Errorf("Name = %s, want Alice", u.Name)
	}
}

func TestHTTPResponse_String_Helper(t *testing.T) {
	t.Parallel()
	r := &HTTPResponse{Body: []byte("hello")}
	if got := r.String(); got != "hello" {
		t.Errorf("String() = %q, want hello", got)
	}
}

func httpStatusNameHelper(code int) string {
	switch {
	case code >= 200 && code < 300:
		return "2xx"
	case code >= 300 && code < 400:
		return "3xx"
	case code >= 400 && code < 500:
		return "4xx"
	case code >= 500 && code < 600:
		return "5xx"
	default:
		return "other"
	}
}

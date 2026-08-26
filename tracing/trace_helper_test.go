package tracing

import (
	"errors"
	"testing"
)

func TestTraceHelperHelper_TraceHTTP_Success(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	called := false
	err := helper.TraceHTTP("GET", "/api/users", func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Error("function should be called")
	}
}

func TestTraceHelperHelper_TraceHTTP_Error(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	wantErr := errors.New("http error")
	err := helper.TraceHTTP("POST", "/api/users", func() error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status != StatusError {
		t.Errorf("expected ERROR status, got %s", spans[0].Status)
	}
}

func TestTraceHelperHelper_TraceDB_Success(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	called := false
	err := helper.TraceDB("SELECT", "SELECT * FROM users", func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Error("function should be called")
	}
}

func TestTraceHelperHelper_TraceDB_Error(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	wantErr := errors.New("db error")
	err := helper.TraceDB("INSERT", "INSERT INTO users", func() error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status != StatusError {
		t.Errorf("expected ERROR status, got %s", spans[0].Status)
	}
}

func TestTraceHelperHelper_TraceRPC_Success(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	called := false
	err := helper.TraceRPC("UserService", "GetUser", func() error {
		called = true
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !called {
		t.Error("function should be called")
	}
}

func TestTraceHelperHelper_TraceRPC_Error(t *testing.T) {
	t.Parallel()
	tracer := NewTracer(WithServiceName("test"))
	helper := NewTraceHelper(tracer)

	wantErr := errors.New("rpc error")
	err := helper.TraceRPC("UserService", "GetUser", func() error {
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Errorf("expected %v, got %v", wantErr, err)
	}

	spans := tracer.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("expected 1 span, got %d", len(spans))
	}
	if spans[0].Status != StatusError {
		t.Errorf("expected ERROR status, got %s", spans[0].Status)
	}
}

func TestNewTraceHelper_Basic(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	helper := NewTraceHelper(tracer)

	if helper == nil {
		t.Fatal("NewTraceHelper returned nil")
	}
	if helper.tracer != tracer {
		t.Error("tracer should be set")
	}
}

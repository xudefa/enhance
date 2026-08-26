package audit

import (
	"errors"
	"testing"
)

func TestAuditInterceptor_NewAuditInterceptor(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	if interceptor == nil {
		t.Fatal("interceptor should not be nil")
	}
}

func TestAuditInterceptor_Intercept_Success(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	interceptor.Intercept("GetUser", []any{"user123"}, map[string]string{"name": "test"}, nil)

	if len(writer.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(writer.events))
	}

	event := writer.events[0]
	if event.Action != EventAccess {
		t.Errorf("expected action Access, got %s", event.Action)
	}
	if event.Resource != "GetUser" {
		t.Errorf("expected resource 'GetUser', got %s", event.Resource)
	}
	if event.Result != "success" {
		t.Errorf("expected result 'success', got %s", event.Result)
	}
	if event.Severity != SeverityInfo {
		t.Errorf("expected severity Info, got %s", event.Severity)
	}

	details, ok := event.Details["args"].([]any)
	if !ok {
		t.Fatal("expected args in details")
	}
	if len(details) != 1 {
		t.Errorf("expected 1 arg, got %d", len(details))
	}
}

func TestAuditInterceptor_Intercept_Error(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	testErr := errors.New("database connection failed")
	interceptor.Intercept("DeleteUser", []any{"user456"}, nil, testErr)

	if len(writer.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(writer.events))
	}

	event := writer.events[0]
	if event.Result != "failure" {
		t.Errorf("expected result 'failure', got %s", event.Result)
	}
	if event.Severity != SeverityError {
		t.Errorf("expected severity Error, got %s", event.Severity)
	}
	if event.ErrorMessage != "database connection failed" {
		t.Errorf("expected error message, got %s", event.ErrorMessage)
	}
}

func TestAuditInterceptor_DefaultActorAndSource(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	interceptor.Intercept("TestMethod", nil, nil, nil)

	event := writer.events[0]
	if event.Actor != "system" {
		t.Errorf("expected default actor 'system', got %s", event.Actor)
	}
	if event.Source != "unknown" {
		t.Errorf("expected default source 'unknown', got %s", event.Source)
	}
}

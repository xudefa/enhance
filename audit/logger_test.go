package audit

import (
	"testing"
)

func TestAuditLogger_NewAuditLogger(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "web-app")
	if logger == nil {
		t.Fatal("logger should not be nil")
	}
}

func TestAuditLogger_Create(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "web-app")
	logger.Create("user", "user:123", map[string]any{
		"name": "test user",
	})

	if len(writer.events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(writer.events))
	}

	event := writer.events[0]
	if event.Action != EventCreate {
		t.Errorf("expected action Create, got %s", event.Action)
	}
	if event.Resource != "user" {
		t.Errorf("expected resource 'user', got %s", event.Resource)
	}
	if event.Target != "user:123" {
		t.Errorf("expected target 'user:123', got %s", event.Target)
	}
	if event.Actor != "admin" {
		t.Errorf("expected actor 'admin', got %s", event.Actor)
	}
	if event.Result != "success" {
		t.Errorf("expected result 'success', got %s", event.Result)
	}
}

func TestAuditLogger_Update(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "web-app")
	logger.Update("user", "user:123", map[string]any{
		"name": "updated user",
	})

	event := writer.events[0]
	if event.Action != EventUpdate {
		t.Errorf("expected action Update, got %s", event.Action)
	}
	if event.Details["name"] != "updated user" {
		t.Errorf("expected updated name in details")
	}
}

func TestAuditLogger_Delete(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "web-app")
	logger.Delete("user", "user:123")

	event := writer.events[0]
	if event.Action != EventDelete {
		t.Errorf("expected action Delete, got %s", event.Action)
	}
	if event.Target != "user:123" {
		t.Errorf("expected target 'user:123', got %s", event.Target)
	}
}

func TestAuditLogger_Login(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "user1", "auth-service")
	logger.Login("user1", map[string]any{
		"ip": "192.168.1.100",
	})

	event := writer.events[0]
	if event.Action != EventLogin {
		t.Errorf("expected action Login, got %s", event.Action)
	}
	if event.Target != "user1" {
		t.Errorf("expected target 'user1', got %s", event.Target)
	}
	if event.Details["ip"] != "192.168.1.100" {
		t.Errorf("expected IP in details")
	}
}

func TestAuditLogger_Severity(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "security", "security-service")
	logger.Severity("user", "user:456", SeverityCritical, map[string]any{
		"reason": "brute force",
	})

	event := writer.events[0]
	if event.Action != EventSecurity {
		t.Errorf("expected action Security, got %s", event.Action)
	}
	if event.Severity != SeverityCritical {
		t.Errorf("expected severity Critical, got %s", event.Severity)
	}
}

func TestAuditLogger_MultipleOperations(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "web-app")
	logger.Create("user", "user:1", nil)
	logger.Update("user", "user:1", nil)
	logger.Delete("user", "user:1")

	if len(writer.events) != 3 {
		t.Errorf("expected 3 events, got %d", len(writer.events))
	}

	actions := []EventType{writer.events[0].Action, writer.events[1].Action, writer.events[2].Action}
	expected := []EventType{EventCreate, EventUpdate, EventDelete}

	for i, action := range actions {
		if action != expected[i] {
			t.Errorf("event %d: expected action %s, got %s", i, expected[i], action)
		}
	}
}

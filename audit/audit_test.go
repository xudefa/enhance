package audit

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"
)

type mockWriter struct {
	mu     sync.Mutex
	events []Event
}

func (w *mockWriter) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
	return nil
}

func (w *mockWriter) Close() error {
	return nil
}

func (w *mockWriter) getEvents() []Event {
	w.mu.Lock()
	defer w.mu.Unlock()
	cp := make([]Event, len(w.events))
	copy(cp, w.events)
	return cp
}

func (w *mockWriter) getEvent(i int) Event {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.events[i]
}

func (w *mockWriter) eventCount() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.events)
}

func TestAuditor_Basic(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:    "user123",
		Action:   EventCreate,
		Resource: "user",
		Target:   "user:456",
	})

	if writer.eventCount() != 1 {
		t.Errorf("expected 1 event, got %d", writer.eventCount())
	}

	event := writer.getEvent(0)
	if event.Actor != "user123" {
		t.Errorf("expected actor user123, got %s", event.Actor)
	}

	if event.Action != EventCreate {
		t.Errorf("expected action CREATE, got %s", event.Action)
	}
}

func TestAuditor_DefaultValues(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:  "user123",
		Action: EventCreate,
	})

	event := writer.getEvent(0)

	if event.Timestamp.IsZero() {
		t.Error("expected timestamp to be set")
	}

	if event.ID == "" {
		t.Error("expected ID to be generated")
	}

	if event.Severity != SeverityInfo {
		t.Errorf("expected severity INFO, got %s", event.Severity)
	}
}

func TestAuditor_LogAction(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:    "user123",
		Action:   EventCreate,
		Resource: "user",
		Target:   "user:456",
		Details: map[string]any{
			"name": "John",
		},
		Result: "success",
	})

	event := writer.getEvent(0)
	if event.Result != "success" {
		t.Errorf("expected result success, got %s", event.Result)
	}

	if event.Details["name"] != "John" {
		t.Errorf("expected detail name John, got %v", event.Details["name"])
	}
}

func TestAuditor_LogSecurity(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:  "user123",
		Action: EventLogin,
		Source: "192.168.1.1",
		Details: map[string]any{
			"reason": "invalid password",
		},
		Severity: SeverityWarning,
		Result:   "failure",
	})

	event := writer.getEvent(0)
	if event.Severity != SeverityWarning {
		t.Errorf("expected severity WARNING, got %s", event.Severity)
	}

	if event.Result != "failure" {
		t.Errorf("expected result failure, got %s", event.Result)
	}
}

func TestAuditor_LogError(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:        "user123",
		Action:       EventCreate,
		Resource:     "user",
		Severity:     SeverityError,
		ErrorMessage: "creation failed",
		Result:       "failure",
	})

	event := writer.getEvent(0)
	if event.Severity != SeverityError {
		t.Errorf("expected severity ERROR, got %s", event.Severity)
	}

	if event.ErrorMessage != "creation failed" {
		t.Errorf("expected error message 'creation failed', got %s", event.ErrorMessage)
	}
}

func TestAuditor_Async(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(10),
	)

	for range 5 {
		auditor.Log(Event{
			Actor:  "user123",
			Action: EventCreate,
		})
	}

	time.Sleep(100 * time.Millisecond)
	_ = auditor.Close()

	if writer.eventCount() != 5 {
		t.Errorf("expected 5 events, got %d", writer.eventCount())
	}
}

func TestAuditor_ConsoleWriter(t *testing.T) {
	t.Parallel()
	writer := NewConsoleWriter()
	defer func() { _ = writer.Close() }()

	err := writer.Write(Event{
		Actor:    "user123",
		Action:   EventCreate,
		Resource: "user",
	})

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

func TestAuditor_FileWriter(t *testing.T) {
	t.Parallel()
	tmpFile := t.TempDir() + "/audit.log"
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}
	defer func() { _ = writer.Close() }()

	err = writer.Write(Event{
		Actor:    "user123",
		Action:   EventCreate,
		Resource: "user",
	})

	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
}

func TestAuditInterceptor_Success(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	interceptor.Intercept("CreateUser", []any{"John", "john@example.com"}, nil, nil)

	if writer.eventCount() != 1 {
		t.Errorf("expected 1 event, got %d", writer.eventCount())
	}

	event := writer.getEvent(0)
	if event.Resource != "CreateUser" {
		t.Errorf("expected resource CreateUser, got %s", event.Resource)
	}
	if event.Result != "success" {
		t.Errorf("expected result success, got %s", event.Result)
	}
}

func TestAuditInterceptor_Failure(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	interceptor := NewAuditInterceptor(auditor)
	interceptor.Intercept("CreateUser", nil, nil, &testError{"failed"})

	if writer.eventCount() != 1 {
		t.Fatalf("expected 1 event, got %d", writer.eventCount())
	}

	event := writer.getEvent(0)
	if event.Result != "failure" {
		t.Errorf("expected result failure, got %s", event.Result)
	}
}

func TestAuditLogger_Create(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "user123", "web-app")
	logger.Create("user", "user:456", map[string]any{"name": "John"})

	event := writer.getEvent(0)
	if event.Action != EventCreate {
		t.Errorf("expected action CREATE, got %s", event.Action)
	}
	if event.Actor != "user123" {
		t.Errorf("expected actor user123, got %s", event.Actor)
	}
}

func TestAuditLogger_Update(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "user123", "web-app")
	logger.Update("user", "user:456", map[string]any{"name": "Jane"})

	event := writer.getEvent(0)
	if event.Action != EventUpdate {
		t.Errorf("expected action UPDATE, got %s", event.Action)
	}
}

func TestAuditLogger_Delete(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "user123", "web-app")
	logger.Delete("user", "user:456")

	event := writer.getEvent(0)
	if event.Action != EventDelete {
		t.Errorf("expected action DELETE, got %s", event.Action)
	}
}

func TestAuditLogger_Login(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "user123", "web-app")
	logger.Login("192.168.1.1", map[string]any{"browser": "Chrome"})

	event := writer.getEvent(0)
	if event.Action != EventLogin {
		t.Errorf("expected action LOGIN, got %s", event.Action)
	}
}

func TestAuditor_MultipleEvents(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{Actor: "user1", Action: EventCreate, Resource: "user", Target: "user:1", Result: "success"})
	auditor.Log(Event{Actor: "user2", Action: EventUpdate, Resource: "user", Target: "user:2", Result: "success"})
	auditor.Log(Event{Actor: "user3", Action: EventDelete, Resource: "user", Target: "user:3", Result: "success"})
	auditor.Log(Event{Actor: "user4", Action: EventLogin, Source: "192.168.1.1", Severity: SeverityWarning})
	auditor.Log(Event{Actor: "user5", Action: EventCreate, Resource: "user", Severity: SeverityError, ErrorMessage: "error", Result: "failure"})

	if writer.eventCount() != 5 {
		t.Errorf("expected 5 events, got %d", writer.eventCount())
	}

	expectedActions := []EventType{EventCreate, EventUpdate, EventDelete, EventLogin, EventCreate}
	events := writer.getEvents()
	for i, event := range events {
		if event.Action != expectedActions[i] {
			t.Errorf("event %d: expected action %s, got %s", i, expectedActions[i], event.Action)
		}
	}
}

func TestEvent_JSONSerialization(t *testing.T) {
	t.Parallel()
	event := Event{
		ID:        "audit-123",
		Timestamp: time.Now(),
		Actor:     "user123",
		Action:    EventCreate,
		Resource:  "user",
		Target:    "user:456",
		Details: map[string]any{
			"name": "John",
			"age":  30,
		},
		Severity: SeverityInfo,
		Source:   "web-app",
		Result:   "success",
		Duration: 100 * time.Millisecond,
		Tags:     []string{"tag1", "tag2"},
	}

	data, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("failed to marshal event: %v", err)
	}

	var decoded Event
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal event: %v", err)
	}

	if decoded.ID != event.ID {
		t.Errorf("expected ID %s, got %s", event.ID, decoded.ID)
	}

	if decoded.Actor != event.Actor {
		t.Errorf("expected actor %s, got %s", event.Actor, decoded.Actor)
	}

	if decoded.Action != event.Action {
		t.Errorf("expected action %s, got %s", event.Action, decoded.Action)
	}
}

func TestAuditor_IDGeneration(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	ids := make(map[string]bool)
	for range 10 {
		auditor.Log(Event{
			Actor:  "user123",
			Action: EventCreate,
		})

		event := writer.getEvent(writer.eventCount() - 1)
		if ids[event.ID] {
			t.Errorf("duplicate ID generated: %s", event.ID)
		}
		ids[event.ID] = true

		if !strings.HasPrefix(event.ID, "audit-") {
			t.Errorf("expected ID to start with 'audit-', got %s", event.ID)
		}
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

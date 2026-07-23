package audit

import (
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type safeMockWriter struct {
	mu    sync.Mutex
	count atomic.Int64
}

func (w *safeMockWriter) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.count.Add(1)
	return nil
}

func (w *safeMockWriter) Close() error {
	return nil
}

func TestAuditLogger_Severity(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	logger := NewAuditLogger(auditor, "admin", "security-service")
	logger.Severity("user", "user:456", SeverityCritical, map[string]any{
		"reason": "brute force",
	})

	event := writer.events[0]
	if event.Action != EventSecurity {
		t.Errorf("expected action SECURITY, got %s", event.Action)
	}
	if event.Severity != SeverityCritical {
		t.Errorf("expected severity CRITICAL, got %s", event.Severity)
	}
	if event.Actor != "admin" {
		t.Errorf("expected actor admin, got %s", event.Actor)
	}
	if event.Source != "security-service" {
		t.Errorf("expected source security-service, got %s", event.Source)
	}
}

func TestAuditor_IsClosed(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))

	if auditor.IsClosed() {
		t.Fatal("auditor should not be closed initially")
	}

	_ = auditor.Close()

	if !auditor.IsClosed() {
		t.Fatal("auditor should be closed after Close()")
	}
}

func TestAuditor_CloseIdempotent(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))

	err1 := auditor.Close()
	err2 := auditor.Close()
	if err1 != nil {
		t.Errorf("first Close() error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close() error: %v", err2)
	}
}

func TestAuditor_AsyncCloseIdempotent(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(10),
	)

	err1 := auditor.Close()
	err2 := auditor.Close()
	if err1 != nil {
		t.Errorf("first async Close() error: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second async Close() error: %v", err2)
	}
}

func TestAuditor_ConcurrentLog(t *testing.T) {
	t.Parallel()
	writer := &safeMockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	var wg sync.WaitGroup
	for range 100 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			auditor.Log(Event{
				Actor:  "user",
				Action: EventCreate,
			})
		}()
	}
	wg.Wait()

	if writer.count.Load() != 100 {
		t.Errorf("expected 100 events, got %d", writer.count.Load())
	}
}

func TestAuditor_AsyncBufferFullFallback(t *testing.T) {
	t.Parallel()
	writer := &safeMockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(2),
	)
	defer func() { _ = auditor.Close() }()

	for range 10 {
		auditor.Log(Event{
			Actor:  "user",
			Action: EventCreate,
		})
	}

	time.Sleep(100 * time.Millisecond)

	if writer.count.Load() != 10 {
		t.Errorf("expected 10 events with fallback, got %d", writer.count.Load())
	}
}

func TestFileWriter_MultipleWrites(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_multi.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}

	for range 5 {
		err := writer.Write(Event{
			Actor:  "user",
			Action: EventCreate,
		})
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
	}

	_ = writer.Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := 0
	for _, b := range data {
		if b == '\n' {
			lines++
		}
	}
	if lines != 5 {
		t.Errorf("expected 5 lines, got %d", lines)
	}
}

func TestFileWriter_CloseFlush(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_flush.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}

	_ = writer.Write(Event{
		Actor:  "user",
		Action: EventCreate,
	})

	err = writer.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected data after close")
	}
}

func TestEvent_ErrorVariables(t *testing.T) {
	t.Parallel()
	if ErrWriterClosed == nil {
		t.Error("ErrWriterClosed should not be nil")
	}
	if ErrChannelFull == nil {
		t.Error("ErrChannelFull should not be nil")
	}
}

func TestAuditor_ExistingTimestampPreserved(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	ts := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	auditor.Log(Event{
		Actor:     "user",
		Action:    EventCreate,
		Timestamp: ts,
		ID:        "custom-id",
		Severity:  SeverityWarning,
	})

	event := writer.events[0]
	if event.ID != "custom-id" {
		t.Errorf("expected custom ID, got %s", event.ID)
	}
	if event.Timestamp != ts {
		t.Errorf("expected custom timestamp, got %v", event.Timestamp)
	}
	if event.Severity != SeverityWarning {
		t.Errorf("expected severity WARNING, got %s", event.Severity)
	}
}

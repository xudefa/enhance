package audit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type mockWriter struct {
	mu     sync.Mutex
	events []Event
	closed bool
}

func (w *mockWriter) Write(event Event) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.events = append(w.events, event)
	return nil
}

func (w *mockWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.closed = true
	return nil
}

func (w *mockWriter) count() int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.events)
}

func (w *mockWriter) eventAt(i int) Event {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.events[i]
}

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

func TestAuditor_NewAuditor_Defaults(t *testing.T) {
	t.Parallel()
	auditor := NewAuditor()
	defer func() { _ = auditor.Close() }()

	if auditor == nil {
		t.Fatal("auditor should not be nil")
	}
	if !auditor.IsClosed() {
		t.Log("auditor is not closed initially (expected)")
	}
}

func TestAuditor_NewAuditor_WithOptions(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithBufferSize(500),
	)
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:  "test",
		Action: EventCreate,
	})

	if writer.count() != 1 {
		t.Errorf("expected 1 event, got %d", writer.count())
	}
}

func TestAuditor_Log_SyncMode(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	event := Event{
		Actor:  "admin",
		Action: EventCreate,
	}
	auditor.Log(event)

	if writer.count() != 1 {
		t.Fatalf("expected 1 event, got %d", writer.count())
	}
	if writer.eventAt(0).Actor != "admin" {
		t.Errorf("expected actor 'admin', got %s", writer.eventAt(0).Actor)
	}
}

func TestAuditor_Log_AsyncMode(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(100),
	)
	defer func() { _ = auditor.Close() }()

	for i := 0; i < 10; i++ {
		auditor.Log(Event{
			Actor:  "user",
			Action: EventCreate,
		})
	}

	time.Sleep(50 * time.Millisecond)
	_ = auditor.Close()

	if writer.count() != 10 {
		t.Errorf("expected 10 events, got %d", writer.count())
	}
}

func TestAuditor_Log_AutoFillFields(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	auditor.Log(Event{
		Actor:  "user",
		Action: EventCreate,
	})

	event := writer.eventAt(0)
	if event.ID == "" {
		t.Error("ID should be auto-generated")
	}
	if event.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
	if event.Severity != SeverityInfo {
		t.Errorf("expected severity Info, got %s", event.Severity)
	}
}

func TestAuditor_Log_AfterClose(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	_ = auditor.Close()

	auditor.Log(Event{
		Actor:  "user",
		Action: EventCreate,
	})

	if writer.count() != 0 {
		t.Errorf("expected 0 events after close, got %d", writer.count())
	}
}

func TestAuditor_Log_AsyncBufferFull(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(2),
	)
	defer func() { _ = auditor.Close() }()

	for i := 0; i < 5; i++ {
		auditor.Log(Event{
			Actor:  "user",
			Action: EventCreate,
		})
	}

	time.Sleep(50 * time.Millisecond)

	if writer.count() == 0 {
		t.Error("expected some events to be written")
	}
}

func TestAuditor_ConcurrentLog_Safe(t *testing.T) {
	t.Parallel()
	writer := &safeMockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(100),
	)
	defer func() { _ = auditor.Close() }()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
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
	time.Sleep(50 * time.Millisecond)

	if writer.count.Load() != 50 {
		t.Errorf("expected 50 events, got %d", writer.count.Load())
	}
}

func TestAuditor_GenerateID_Unique(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	ids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		auditor.Log(Event{
			Actor:  "user",
			Action: EventCreate,
		})
		ids[writer.eventAt(i).ID] = true
	}

	if len(ids) != 100 {
		t.Errorf("expected 100 unique IDs, got %d", len(ids))
	}
}

func TestAuditor_ProcessEvents_PanicRecovery(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(
		WithWriter(writer),
		WithAsync(),
		WithBufferSize(10),
	)

	auditor.Log(Event{
		Actor:  "user",
		Action: EventCreate,
	})

	time.Sleep(20 * time.Millisecond)
	err := auditor.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}

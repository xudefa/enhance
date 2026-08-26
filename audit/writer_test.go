package audit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConsoleWriter_NewConsoleWriter(t *testing.T) {
	t.Parallel()
	writer := NewConsoleWriter()
	if writer == nil {
		t.Fatal("writer should not be nil")
	}
}

func TestConsoleWriter_Write(t *testing.T) {
	t.Parallel()
	writer := &mockWriter{}
	auditor := NewAuditor(WithWriter(writer))
	defer func() { _ = auditor.Close() }()

	event := Event{
		Actor:  "test",
		Action: EventCreate,
	}
	auditor.Log(event)

	if len(writer.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(writer.events))
	}
}

func TestConsoleWriter_Close(t *testing.T) {
	t.Parallel()
	writer := NewConsoleWriter()
	err := writer.Close()
	if err != nil {
		t.Errorf("Close should not error: %v", err)
	}
}

func TestFileWriter_NewFileWriter(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_test.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}
	if writer == nil {
		t.Fatal("writer should not be nil")
	}
	_ = writer.Close()
}

func TestFileWriter_NewFileWriter_InvalidPath(t *testing.T) {
	t.Parallel()
	_, err := NewFileWriter("/invalid/path/audit.log")
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

func TestFileWriter_Write(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_write.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}

	event := Event{
		Actor:  "user",
		Action: EventCreate,
	}
	err = writer.Write(event)
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}

	_ = writer.Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected data in file")
	}
	if !strings.Contains(string(data), "user") {
		t.Error("expected actor 'user' in file")
	}
}

func TestFileWriter_WriteMultiple(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_multi.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}

	for i := 0; i < 5; i++ {
		err = writer.Write(Event{
			Actor:  "user",
			Action: EventCreate,
		})
		if err != nil {
			t.Errorf("Write %d failed: %v", i, err)
		}
	}

	_ = writer.Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 5 {
		t.Errorf("expected 5 lines, got %d", len(lines))
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
		t.Errorf("Close failed: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected data after flush")
	}
}

func TestFileWriter_ConcurrentWrite(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_concurrent.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 10; j++ {
				_ = writer.Write(Event{
					Actor:  "user",
					Action: EventCreate,
				})
			}
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	_ = writer.Close()

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 100 {
		t.Errorf("expected 100 lines, got %d", len(lines))
	}
}

func TestConsoleWriter_JSONMarshalError(t *testing.T) {
	t.Parallel()
	writer := NewConsoleWriter()

	event := Event{
		Details: map[string]any{
			"invalid": make(chan int),
		},
	}

	err := writer.Write(event)
	if err == nil {
		t.Error("expected error for unmarshalable data")
	}
}

func TestFileWriter_JSONMarshalError(t *testing.T) {
	t.Parallel()
	tmpFile := filepath.Join(t.TempDir(), "audit_error.log")
	writer, err := NewFileWriter(tmpFile)
	if err != nil {
		t.Fatalf("NewFileWriter failed: %v", err)
	}
	defer func() { _ = writer.Close() }()

	event := Event{
		Details: map[string]any{
			"invalid": make(chan int),
		},
	}

	err = writer.Write(event)
	if err == nil {
		t.Error("expected error for unmarshalable data")
	}
}

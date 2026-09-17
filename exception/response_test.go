package exception

import (
	"encoding/json"
	"testing"
	"time"
)

func TestErrorResponse_MarshalJSON(t *testing.T) {
	t.Parallel()
	resp := &ErrorResponse{
		Code:      404,
		Message:   "Not found",
		RequestID: "req-123",
		TraceID:   "trace-456",
		Details:   map[string]string{"id": "123"},
		Timestamp: time.Now().UnixMilli(),
	}

	jsonData, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	var parsed map[string]any
	if err := json.Unmarshal(jsonData, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if parsed["code"].(float64) != 404 {
		t.Errorf("Expected code 404, got %v", parsed["code"])
	}
	if parsed["message"].(string) != "Not found" {
		t.Errorf("Expected message 'Not found', got %v", parsed["message"])
	}
}

func TestNewErrorResponse(t *testing.T) {
	t.Parallel()
	resp := NewErrorResponse(500, "Internal error", WithRequestID("req-123"), WithTraceID("trace-456"))

	if resp.Code != 500 {
		t.Errorf("Expected code 500, got %d", resp.Code)
	}
	if resp.Message != "Internal error" {
		t.Errorf("Expected message 'Internal error', got %s", resp.Message)
	}
	if resp.Timestamp == 0 {
		t.Error("Expected timestamp to be set")
	}
}

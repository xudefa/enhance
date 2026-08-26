package tracing

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
)

func TestSpanHelper_SetTag(t *testing.T) {
	t.Parallel()
	span := &Span{Tags: make(map[string]string)}

	span.SetTag("http.method", "GET")
	span.SetTag("http.url", "/api")

	if span.Tags["http.method"] != "GET" {
		t.Errorf("Tags[http.method] = %s, want GET", span.Tags["http.method"])
	}
	if span.Tags["http.url"] != "/api" {
		t.Errorf("Tags[http.url] = %s, want /api", span.Tags["http.url"])
	}
}

func TestSpanHelper_SetTag_NilTags(t *testing.T) {
	t.Parallel()
	span := &Span{}
	span.SetTag("key", "value")

	if span.Tags == nil {
		t.Fatal("Tags should be initialized")
	}
	if span.Tags["key"] != "value" {
		t.Errorf("Tags[key] = %s, want value", span.Tags["key"])
	}
}

func TestSpanHelper_AddEvent(t *testing.T) {
	t.Parallel()
	span := &Span{Events: make([]SpanEvent, 0)}

	span.AddEvent("event1")
	if len(span.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(span.Events))
	}
	if span.Events[0].Name != "event1" {
		t.Errorf("event name = %s, want event1", span.Events[0].Name)
	}
}

func TestSpanHelper_AddEvent_WithAttrs(t *testing.T) {
	t.Parallel()
	span := &Span{Events: make([]SpanEvent, 0)}

	attrs := map[string]string{"key": "value"}
	span.AddEvent("event1", attrs)

	if len(span.Events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(span.Events))
	}
	if span.Events[0].Attributes["key"] != "value" {
		t.Error("event attributes should contain key=value")
	}
}

func TestSpanHelper_SetStatus(t *testing.T) {
	t.Parallel()
	span := &Span{}

	span.SetStatus(StatusOK)
	if span.Status != StatusOK {
		t.Errorf("Status = %s, want %s", span.Status, StatusOK)
	}

	span.SetStatus(StatusError)
	if span.Status != StatusError {
		t.Errorf("Status = %s, want %s", span.Status, StatusError)
	}
}

func TestSpanHelper_End(t *testing.T) {
	t.Parallel()
	span := &Span{StartTime: time.Now()}

	span.End()

	if !span.Ended {
		t.Error("Ended should be true")
	}
	if span.EndTime.IsZero() {
		t.Error("EndTime should be set")
	}
}

func TestSpanHelper_End_Idempotent(t *testing.T) {
	t.Parallel()
	span := &Span{StartTime: time.Now()}

	span.End()
	firstEnd := span.EndTime

	time.Sleep(time.Millisecond)
	span.End()

	if span.EndTime != firstEnd {
		t.Error("End should be idempotent")
	}
}

func TestSpanHelper_Duration_Ended(t *testing.T) {
	t.Parallel()
	start := time.Now()
	span := &Span{StartTime: start}

	time.Sleep(time.Millisecond)
	span.End()

	d := span.Duration()
	if d <= 0 {
		t.Errorf("Duration should be positive, got %v", d)
	}
}

func TestSpanHelper_Duration_NotEnded(t *testing.T) {
	t.Parallel()
	start := time.Now()
	span := &Span{StartTime: start}

	time.Sleep(time.Millisecond)
	d := span.Duration()

	if d <= 0 {
		t.Errorf("Duration should be positive, got %v", d)
	}
}

func TestSpanHelper_Context(t *testing.T) {
	t.Parallel()
	ctx := SpanContext{
		TraceID: "trace-1",
		SpanID:  "span-1",
		Sampled: true,
	}
	span := &Span{spanContext: ctx}

	got := span.Context()
	if got.TraceID != "trace-1" {
		t.Errorf("TraceID = %s, want trace-1", got.TraceID)
	}
	if got.SpanID != "span-1" {
		t.Errorf("SpanID = %s, want span-1", got.SpanID)
	}
	if !got.Sampled {
		t.Error("Sampled should be true")
	}
}

func TestSpanHelper_MarshalJSON(t *testing.T) {
	t.Parallel()
	span := &Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test",
		StartTime: time.Now(),
		Ended:     true,
		Tags:      map[string]string{"k": "v"},
		Events:    []SpanEvent{{Name: "e1", Timestamp: time.Now()}},
	}
	span.EndTime = span.StartTime.Add(100 * time.Millisecond)

	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if m["trace_id"] != "trace-1" {
		t.Errorf("trace_id = %v, want trace-1", m["trace_id"])
	}
	if m["name"] != "test" {
		t.Errorf("name = %v, want test", m["name"])
	}
	if _, ok := m["duration_ms"]; !ok {
		t.Error("should contain duration_ms")
	}
}

func TestSpanHelper_MarshalJSON_NotEnded(t *testing.T) {
	t.Parallel()
	span := &Span{
		TraceID:   "trace-1",
		SpanID:    "span-1",
		Name:      "test",
		StartTime: time.Now(),
		Tags:      make(map[string]string),
	}

	data, err := json.Marshal(span)
	if err != nil {
		t.Fatalf("MarshalJSON error: %v", err)
	}

	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if _, ok := m["duration_ms"]; !ok {
		t.Error("should contain duration_ms for not-ended span")
	}
}

func TestSpanHelper_ConcurrentAccess(t *testing.T) {
	t.Parallel()
	span := &Span{StartTime: time.Now()}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			span.SetTag("key", "value")
			span.AddEvent("event")
			span.SetStatus(StatusOK)
			span.Duration()
		}()
	}
	wg.Wait()
}

package tracing

import (
	"encoding/json"
	"sync"
	"time"
)

// Span 表示一个追踪单元，记录操作的开始和结束时间。
//
// Span 是分布式链路追踪的基本单元，用于记录一个操作的完整生命周期。
// 每个 Span 包含 TraceID、SpanID、操作名称、状态、标签和事件等信息。
//
// Span 是并发安全的，所有方法都通过互斥锁保护。
type Span struct {
	TraceID      TraceID           `json:"trace_id"`
	SpanID       SpanID            `json:"span_id"`
	ParentSpanID SpanID            `json:"parent_span_id,omitempty"`
	Name         string            `json:"name"`
	StartTime    time.Time         `json:"start_time"`
	EndTime      time.Time         `json:"end_time"`
	Status       SpanStatus        `json:"status"`
	Tags         map[string]string `json:"tags"`
	Events       []SpanEvent       `json:"events,omitempty"`
	Ended        bool              `json:"ended"`
	spanContext  SpanContext
	mu           sync.Mutex
}

// SetTag 设置 Span 标签。
//
// 标签用于记录额外的元数据，如 HTTP 方法、URL、数据库操作等。
func (s *Span) SetTag(key, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Tags == nil {
		s.Tags = make(map[string]string)
	}
	s.Tags[key] = value
}

// AddEvent 添加 Span 事件。
//
// 事件用于记录 Span 生命周期中的重要时间点，可附加属性。
func (s *Span) AddEvent(name string, attrs ...map[string]string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	event := SpanEvent{
		Timestamp: time.Now(),
		Name:      name,
	}

	if len(attrs) > 0 {
		event.Attributes = attrs[0]
	}

	s.Events = append(s.Events, event)
}

// SetStatus 设置 Span 状态。
func (s *Span) SetStatus(status SpanStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Status = status
}

// End 结束 Span。
//
// 记录 Span 的结束时间，计算持续时间。
// 如果 Span 已结束，此方法为幂等操作。
func (s *Span) End() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Ended {
		return
	}

	s.EndTime = time.Now()
	s.Ended = true
}

// Duration 获取 Span 持续时间。
//
// 如果 Span 未结束，返回从开始到当前的时间。
func (s *Span) Duration() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.EndTime.IsZero() {
		return time.Since(s.StartTime)
	}
	return s.EndTime.Sub(s.StartTime)
}

// Context 获取 Span 追踪上下文。
//
// 返回的 SpanContext 可用于注入到 HTTP 请求头，实现跨服务追踪。
func (s *Span) Context() SpanContext {
	return s.spanContext
}

// MarshalJSON 自定义 JSON 序列化。
//
// 添加 duration_ms 字段，表示 Span 持续时间（毫秒）。
// 该方法在锁内完成所有操作，避免递归锁定。
func (s *Span) MarshalJSON() ([]byte, error) {
	s.mu.Lock()
	endTime := s.EndTime
	startTime := s.StartTime
	ended := s.Ended
	traceID := s.TraceID
	spanID := s.SpanID
	parentSpanID := s.ParentSpanID
	name := s.Name
	status := s.Status
	tags := make(map[string]string)
	for k, v := range s.Tags {
		tags[k] = v
	}
	events := make([]SpanEvent, len(s.Events))
	copy(events, s.Events)
	s.mu.Unlock()

	var durationMs int64
	if endTime.IsZero() && !ended {
		durationMs = time.Since(startTime).Milliseconds()
	} else {
		durationMs = endTime.Sub(startTime).Milliseconds()
	}

	data := struct {
		TraceID      TraceID           `json:"trace_id"`
		SpanID       SpanID            `json:"span_id"`
		ParentSpanID SpanID            `json:"parent_span_id,omitempty"`
		Name         string            `json:"name"`
		StartTime    time.Time         `json:"start_time"`
		EndTime      time.Time         `json:"end_time"`
		Status       SpanStatus        `json:"status"`
		Tags         map[string]string `json:"tags"`
		Events       []SpanEvent       `json:"events,omitempty"`
		Ended        bool              `json:"ended"`
		DurationMs   int64             `json:"duration_ms"`
	}{
		TraceID:      traceID,
		SpanID:       spanID,
		ParentSpanID: parentSpanID,
		Name:         name,
		StartTime:    startTime,
		EndTime:      endTime,
		Status:       status,
		Tags:         tags,
		Events:       events,
		Ended:        ended,
		DurationMs:   durationMs,
	}
	return json.Marshal(&data)
}

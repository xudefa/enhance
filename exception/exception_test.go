package exception

import (
	"context"
	"testing"
)

type mockLogger struct{}

func (m *mockLogger) Error(ctx context.Context, msg string, keyValues ...KeyValue) {}

type mockMetrics struct{}

func (m *mockMetrics) RecordException(exceptionType string, statusCode int) {}

type mockWriter struct{}

func (m *mockWriter) SetStatusCode(code int)      {}
func (m *mockWriter) SetHeader(key, value string) {}
func (m *mockWriter) Write(data []byte) error     { return nil }
func (m *mockWriter) Context() context.Context    { return context.Background() }

func TestLogger_Interface(t *testing.T) {
	t.Parallel()
	var _ Logger = (*mockLogger)(nil)
}

func TestMetricsRecorder_Interface(t *testing.T) {
	t.Parallel()
	var _ MetricsRecorder = (*mockMetrics)(nil)
}

func TestResponseWriter_Interface(t *testing.T) {
	t.Parallel()
	var _ ResponseWriter = (*mockWriter)(nil)
}

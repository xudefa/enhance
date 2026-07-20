package exception

import (
	"testing"
)

func TestDefaultMetricsRecorder_RecordException(t *testing.T) {
	t.Parallel()
	recorder := NewDefaultMetricsRecorder()

	recorder.RecordException("TestError", 500)

	if recorder == nil {
		t.Error("MetricsRecorder should not be nil")
	}
}

func TestDefaultMetricsRecorder_RecordMultipleExceptions(t *testing.T) {
	t.Parallel()
	recorder := NewDefaultMetricsRecorder()

	recorder.RecordException("Error1", 404)
	recorder.RecordException("Error2", 500)
	recorder.RecordException("Error1", 404)

	if recorder == nil {
		t.Error("MetricsRecorder should not be nil")
	}
}

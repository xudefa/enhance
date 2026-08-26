package tracing

import (
	"errors"
	"testing"
)

func TestErrors_AreSentinels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrExporterNotSet", ErrExporterNotSet, "exporter not set"},
		{"ErrInvalidSamplingRate", ErrInvalidSamplingRate, "invalid sampling rate: must be between 0.0 and 1.0"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err == nil {
				t.Fatal("error should not be nil")
			}
			if tt.err.Error() != tt.msg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.msg)
			}
		})
	}
}

func TestErrors_Is(t *testing.T) {
	t.Parallel()
	wrapped := errors.New("wrapped: exporter not set")
	if errors.Is(wrapped, ErrExporterNotSet) {
		t.Error("wrapped error should not match sentinel")
	}
	if !errors.Is(ErrExporterNotSet, ErrExporterNotSet) {
		t.Error("sentinel should match itself")
	}
}

package cache

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrors_AreAccessible(t *testing.T) {
	t.Parallel()
	if ErrNotFound == nil {
		t.Error("ErrNotFound should not be nil")
	}
	if ErrCacheMiss == nil {
		t.Error("ErrCacheMiss should not be nil")
	}
}

func TestErrors_AreDistinct(t *testing.T) {
	t.Parallel()
	if errors.Is(ErrNotFound, ErrCacheMiss) {
		t.Error("ErrNotFound and ErrCacheMiss should be distinct")
	}
}

func TestErrors_HaveCorrectMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		err     error
		wantMsg string
	}{
		{"ErrNotFound", ErrNotFound, "cache: key not found"},
		{"ErrCacheMiss", ErrCacheMiss, "cache: key expired or not found"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Error() != tt.wantMsg {
				t.Errorf("error message = %q, want %q", tt.err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestErrors_CanBeWrappedAndUnwrapped(t *testing.T) {
	t.Parallel()
	wrapped := fmt.Errorf("outer: %w", ErrNotFound)
	if !errors.Is(wrapped, ErrNotFound) {
		t.Error("wrapped error should match ErrNotFound via errors.Is")
	}
}

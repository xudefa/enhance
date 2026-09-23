package resilience

import (
	"testing"
)

func TestState_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		state    State
		expected string
	}{
		{StateClosed, "closed"},
		{StateOpen, "open"},
		{StateHalfOpen, "half-open"},
		{State(999), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			stateString := tt.state.String()
			if stateString != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, stateString)
			}
		})
	}
}

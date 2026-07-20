package lifecycle

import (
	"errors"
	"testing"
)

func TestApplicationPhase_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		phase    ApplicationPhase
		expected string
	}{
		{PhaseInit, "INIT"},
		{PhaseRunning, "RUNNING"},
		{PhaseStopped, "STOPPED"},
		{ApplicationPhase(99), "UNKNOWN(99)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.phase.String(); got != tt.expected {
				t.Errorf("String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLifecycleManager_PhaseTransitions(t *testing.T) {
	t.Parallel()
	t.Run("valid forward transitions", func(t *testing.T) {
		mgr := NewLifecycleManager()

		if mgr.GetPhase() != PhaseInit {
			t.Errorf("initial phase = %s, want %s", mgr.GetPhase(), PhaseInit)
		}

		err := mgr.SetPhase(PhaseRunning)
		if err != nil {
			t.Fatalf("SetPhase(PhaseRunning) failed: %v", err)
		}

		if mgr.GetPhase() != PhaseRunning {
			t.Errorf("phase = %s, want %s", mgr.GetPhase(), PhaseRunning)
		}
	})

	t.Run("invalid backward transition returns error", func(t *testing.T) {
		mgr := NewLifecycleManager()
		_ = mgr.SetPhase(PhaseRunning)

		err := mgr.SetPhase(PhaseInit)
		if err == nil {
			t.Error("expected error for backward transition, got nil")
		}
	})

	t.Run("same phase transition returns error", func(t *testing.T) {
		mgr := NewLifecycleManager()

		_ = mgr.SetPhase(PhaseInit)
		if err := mgr.SetPhase(PhaseInit); err == nil {
			t.Error("expected error for same phase transition, got nil")
		}
	})
}

func TestLifecycleManager_Listener(t *testing.T) {
	t.Parallel()
	t.Run("listener receives phase change events", func(t *testing.T) {
		mgr := NewLifecycleManager()
		var transitions []PhaseTransition
		mgr.AddListener(&phaseListenerFunc{fn: func(old, new ApplicationPhase) error {
			transitions = append(transitions, PhaseTransition{old, new})
			return nil
		}})

		_ = mgr.SetPhase(PhaseRunning)

		if len(transitions) != 1 {
			t.Errorf("transition count = %d, want 1", len(transitions))
		}

		if transitions[0].OldPhase != PhaseInit || transitions[0].NewPhase != PhaseRunning {
			t.Errorf("unexpected first transition: %v", transitions[0])
		}
	})

	t.Run("listener error is returned", func(t *testing.T) {
		mgr := NewLifecycleManager()
		expectedErr := errors.New("listener error")

		mgr.AddListener(&errorListener{err: expectedErr})

		err := mgr.SetPhase(PhaseRunning)
		if err != expectedErr {
			t.Errorf("error = %v, want %v", err, expectedErr)
		}
	})

	t.Run("error handler is called on listener error", func(t *testing.T) {
		mgr := NewLifecycleManager()
		var handlerCalled bool
		expectedErr := errors.New("listener error")

		mgr.AddListener(&errorListener{err: expectedErr})
		mgr.SetErrorHandler(func(oldPhase, newPhase ApplicationPhase, err error) {
			handlerCalled = true
		})

		_ = mgr.SetPhase(PhaseRunning)
		if !handlerCalled {
			t.Error("expected error handler to be called")
		}
	})
}

type errorListener struct {
	err error
}

func (l *errorListener) OnPhaseChange(oldPhase, newPhase ApplicationPhase) error {
	return l.err
}

type phaseListenerFunc struct {
	fn func(old, new ApplicationPhase) error
}

func (l *phaseListenerFunc) OnPhaseChange(old, new ApplicationPhase) error {
	return l.fn(old, new)
}

func TestLifecycleBuilder(t *testing.T) {
	t.Parallel()
	t.Run("build with configuration", func(t *testing.T) {
		var transitions []PhaseTransition
		listener := &phaseListenerFunc{fn: func(old, new ApplicationPhase) error {
			transitions = append(transitions, PhaseTransition{old, new})
			return nil
		}}

		mgr := NewLifecycleBuilder().
			InitialPhase(PhaseRunning).
			Listener(listener).
			Build()

		if mgr.GetPhase() != PhaseRunning {
			t.Errorf("phase = %s, want %s", mgr.GetPhase(), PhaseRunning)
		}
	})
}

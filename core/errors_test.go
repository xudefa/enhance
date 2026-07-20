package core

import (
	"errors"
	"testing"
)

func TestErrorMessages(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"ErrContainerAlreadyInitialized", ErrContainerAlreadyInitialized, "container already initialized"},
		{"ErrContainerDestroyed", ErrContainerDestroyed, "container has been destroyed"},
		{"ErrBeanNotFound", ErrBeanNotFound, "bean not found"},
		{"ErrBeanAlreadyExists", ErrBeanAlreadyExists, "bean already exists"},
		{"ErrInvalidBeanName", ErrInvalidBeanName, "invalid bean name"},
		{"ErrCircularDependency", ErrCircularDependency, "circular dependency detected"},
		{"ErrDependencyNotFound", ErrDependencyNotFound, "dependency bean not found"},
		{"ErrInjectFailed", ErrInjectFailed, "failed to inject dependencies"},
		{"ErrNilFactory", ErrNilFactory, "factory function cannot be nil"},
		{"ErrInitFailed", ErrInitFailed, "bean initialization failed"},
		{"ErrDestroyFailed", ErrDestroyFailed, "bean destruction failed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.msg {
				t.Errorf("Expected %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestErrorSentinels(t *testing.T) {
	t.Parallel()
	if errors.Is(ErrContainerAlreadyInitialized, ErrContainerDestroyed) {
		t.Error("Errors should be distinct")
	}

	if errors.Is(ErrBeanNotFound, ErrBeanAlreadyExists) {
		t.Error("Errors should be distinct")
	}

	if errors.Is(ErrCircularDependency, ErrDependencyNotFound) {
		t.Error("Errors should be distinct")
	}

	if errors.Is(ErrInitFailed, ErrDestroyFailed) {
		t.Error("Errors should be distinct")
	}
}

func TestErrorComparison(t *testing.T) {
	t.Parallel()
	err := ErrContainerAlreadyInitialized
	if err != ErrContainerAlreadyInitialized {
		t.Error("Error comparison failed")
	}

	err2 := ErrBeanNotFound
	if err == err2 {
		t.Error("Different errors should not be equal")
	}
}

func TestWrappedError(t *testing.T) {
	t.Parallel()
	wrapped := errors.Join(ErrContainerAlreadyInitialized, ErrBeanNotFound)

	if !errors.Is(wrapped, ErrContainerAlreadyInitialized) {
		t.Error("Should detect wrapped ErrContainerAlreadyInitialized")
	}

	if !errors.Is(wrapped, ErrBeanNotFound) {
		t.Error("Should detect wrapped ErrBeanNotFound")
	}
}

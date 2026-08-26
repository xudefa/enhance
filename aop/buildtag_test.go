//go:build !goaop

package aop

import "testing"

func TestBuildTagChecker_DefaultBuildHasNoGoaopTag(t *testing.T) {
	t.Parallel()
	c := NewBuildTagChecker()
	if c.HasTag("goaop") {
		t.Error("expected HasTag(goaop) = false in default build")
	}
	if c.IsGeneratedMode() {
		t.Error("expected default build to be runtime mode")
	}
	if c.GetOptimalMode() != AopModeRuntime {
		t.Errorf("expected runtime mode, got %s", c.GetOptimalMode())
	}
}

func TestBuildTagChecker_OtherTagsAlwaysFalse(t *testing.T) {
	t.Parallel()
	c := NewBuildTagChecker()
	if c.HasTag("unknown") {
		t.Error("expected HasTag(unknown) = false")
	}
	if c.HasTag("custom") {
		t.Error("expected HasTag(custom) = false")
	}
}

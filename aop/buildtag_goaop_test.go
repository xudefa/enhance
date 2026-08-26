//go:build goaop

package aop

import "testing"

func TestBuildTagChecker_GoaopTaggedBuild(t *testing.T) {
	t.Parallel()
	c := NewBuildTagChecker()
	if !c.HasTag("goaop") {
		t.Error("expected HasTag(goaop) = true in goaop-tagged build")
	}
	if !c.IsGeneratedMode() {
		t.Error("expected generated mode in goaop-tagged build")
	}
	if c.GetOptimalMode() != AopModeGenerated {
		t.Errorf("expected generated mode, got %s", c.GetOptimalMode())
	}
}

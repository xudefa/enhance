package validation

import (
	"testing"
)

func TestCompileRegex_Cached(t *testing.T) {
	t.Parallel()
	re1 := compileRegex(`^[a-z]+$`)
	if re1 == nil {
		t.Fatal("expected non-nil regex")
	}

	re2 := compileRegex(`^[a-z]+$`)
	if re2 == nil {
		t.Fatal("expected non-nil cached regex")
	}
}

func TestCompileRegex_Invalid(t *testing.T) {
	t.Parallel()
	re := compileRegex(`[invalid`)
	if re != nil {
		t.Error("expected nil for invalid regex")
	}
}

func TestAcquireValidationErrors(t *testing.T) {
	t.Parallel()
	p := acquireValidationErrors()
	if p == nil {
		t.Fatal("expected non-nil pool")
	}
	if len(*p) != 0 {
		t.Errorf("expected empty slice, got length %d", len(*p))
	}
	releaseValidationErrors(p)
}

func TestAcquireAndReleaseValidationErrors(t *testing.T) {
	t.Parallel()
	p1 := acquireValidationErrors()
	*p1 = append(*p1, ValidationError{Field: "name", Message: "required"})
	releaseValidationErrors(p1)

	p2 := acquireValidationErrors()
	if len(*p2) != 0 {
		t.Errorf("expected empty slice after release, got length %d", len(*p2))
	}
	releaseValidationErrors(p2)
}

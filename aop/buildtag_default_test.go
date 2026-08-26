//go:build !goaop

package aop

import "testing"

func TestHasGoAopBuildTag_IsFalse(t *testing.T) {
	t.Parallel()
	if hasGoAopBuildTag {
		t.Error("expected hasGoAopBuildTag to be false in default build")
	}
}

package aop

import (
	"context"
	"testing"
)

func TestNewAopIntegration_WithCustomConfig(t *testing.T) {
	t.Parallel()

	config := &AopConfig{
		Mode:        AopModeRuntime,
		EnableCache: false,
	}
	integration := NewAopIntegration(config)

	if integration.config != config {
		t.Error("should use provided config")
	}
	if integration.manager == nil {
		t.Error("manager should be initialized")
	}
	if integration.proxyFactory == nil {
		t.Error("proxyFactory should be initialized")
	}
	if integration.metadataExtractor == nil {
		t.Error("metadataExtractor should be initialized")
	}
}

func TestNewAopIntegration_NilConfig(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(nil)
	if integration.config == nil {
		t.Error("should use default config when nil passed")
	}
	if integration.config.Mode != AopModeMixed {
		t.Errorf("default mode = %v, want %v", integration.config.Mode, AopModeMixed)
	}
}

func TestAopIntegration_GetManager(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	manager := integration.GetManager()
	if manager == nil {
		t.Fatal("GetManager() returned nil")
	}
}

func TestAopIntegration_CreateProxy_RuntimeMode(t *testing.T) {
	t.Parallel()

	config := &AopConfig{
		Mode:        AopModeRuntime,
		Weaver:      NewWeaver(),
		EnableCache: true,
	}
	integration := NewAopIntegration(config)

	type TestBean struct{ Name string }
	bean := &TestBean{Name: "test"}
	proxy := integration.CreateProxy("testBean", bean)
	if proxy == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestAopIntegration_CreateProxy_NilConfig(t *testing.T) {
	t.Parallel()

	integration := &AopIntegration{config: nil}

	type TestBean struct{ Name string }
	bean := &TestBean{Name: "test"}
	result := integration.CreateProxy("testBean", bean)
	if result != bean {
		t.Error("should return original target when config is nil")
	}
}

func TestAopIntegration_RegisterAndGetAspects(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	aspect := &AspectMeta{Order: 1}
	integration.RegisterAspect(aspect)

	aspects := integration.GetAspects()
	found := false
	for _, a := range aspects {
		if a == aspect {
			found = true
			break
		}
	}
	if !found {
		t.Error("registered aspect not found")
	}
}

func TestAopIntegration_RegisterAspects(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	a1 := &AspectMeta{Order: 1}
	a2 := &AspectMeta{Order: 2}
	integration.RegisterAspects(a1, a2)

	aspects := integration.GetAspects()
	count := 0
	for _, a := range aspects {
		if a == a1 || a == a2 {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 registered aspects, found %d", count)
	}
}

func TestAopIntegration_RegisterAndGetProxyType(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())

	integration.RegisterProxyType("MyProxy", "/path/to/proxy.go")
	path, ok := integration.GetScannedProxy("MyProxy")
	if !ok {
		t.Error("expected to find registered proxy type")
	}
	if path != "/path/to/proxy.go" {
		t.Errorf("path = %q, want %q", path, "/path/to/proxy.go")
	}
}

func TestAopIntegration_GetScannedProxy_NotFound(t *testing.T) {
	t.Parallel()

	integration := NewAopIntegration(DefaultAopConfig())
	_, ok := integration.GetScannedProxy("NonExistent")
	if ok {
		t.Error("should return false for non-existent proxy type")
	}
}

func TestBuildTagChecker_HasTag(t *testing.T) {
	t.Parallel()

	checker := NewBuildTagChecker()

	// Without goaop build tag, HasTag("goaop") should return false
	if checker.HasTag("goaop") {
		t.Log("goaop build tag present (unexpected in test)")
	}

	// Unknown tag should always return false
	if checker.HasTag("unknown") {
		t.Error("HasTag('unknown') should return false")
	}
}

func TestBuildTagChecker_GetOptimalMode(t *testing.T) {
	t.Parallel()

	checker := NewBuildTagChecker()
	mode := checker.GetOptimalMode()
	if mode != AopModeRuntime && mode != AopModeGenerated {
		t.Errorf("unexpected mode: %v", mode)
	}
}

func TestGetGlobalAopIntegration(t *testing.T) {
	t.Parallel()

	global := GetGlobalAopIntegration()
	if global == nil {
		t.Fatal("GetGlobalAopIntegration() returned nil")
	}
}

func TestSetGlobalAopIntegration(t *testing.T) {
	t.Parallel()

	original := GetGlobalAopIntegration()
	defer SetGlobalAopIntegration(original)

	custom := NewAopIntegration(&AopConfig{Mode: AopModeGenerated})
	SetGlobalAopIntegration(custom)
	if GetGlobalAopIntegration() != custom {
		t.Error("SetGlobalAopIntegration did not update global")
	}
}

func TestCreateProxy_GlobalFunc(t *testing.T) {
	t.Parallel()

	type TestBean struct{ Name string }
	bean := &TestBean{Name: "test"}

	result := CreateProxy("testBean", bean)
	if result == nil {
		t.Error("expected non-nil result from global CreateProxy")
	}
}

func TestGetGlobalAspects(t *testing.T) {
	t.Parallel()

	aspects := GetGlobalAspects()
	if aspects == nil {
		t.Error("GetGlobalAspects() should not return nil")
	}
}

func TestAutoRegister_NonExistentBean(t *testing.T) {
	t.Parallel()

	err := AutoRegister("NonExistentBean_12345")
	if err == nil {
		t.Error("expected error for non-existent bean")
	}
}

func TestAutoRegisterAll_NoPanic(t *testing.T) {
	t.Parallel()

	// AutoRegisterAll should not panic even when no beans are registered
	_ = AutoRegisterAll()
}

func TestGlobalBuildTagChecker_Exists(t *testing.T) {
	t.Parallel()

	if GlobalBuildTagChecker == nil {
		t.Error("GlobalBuildTagChecker should not be nil")
	}
}

func TestAopMode_Values(t *testing.T) {
	t.Parallel()

	tests := []struct {
		mode AopMode
		want string
	}{
		{AopModeRuntime, "runtime"},
		{AopModeGenerated, "generated"},
		{AopModeMixed, "mixed"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(string(tt.mode), func(t *testing.T) {
			t.Parallel()
			if string(tt.mode) != tt.want {
				t.Errorf("mode string = %q, want %q", string(tt.mode), tt.want)
			}
		})
	}
}

func TestAopIntegration_CreateProxy_MixedMode_WithAspect(t *testing.T) {
	t.Parallel()

	config := &AopConfig{
		Mode:        AopModeMixed,
		Weaver:      NewWeaver(),
		EnableCache: true,
	}
	integration := NewAopIntegration(config)

	type MixedBean struct{ Val int }
	bean := &MixedBean{Val: 42}

	integration.RegisterAspect(&AspectMeta{
		PointCut: MatchByClassName("MixedBean"),
		Advice:   Before(func(jp JoinPoint) {}),
		Order:    1,
	})

	ctx := context.Background()
	_ = ctx

	proxy := integration.CreateProxy("mixedBean", bean)
	if proxy == nil {
		t.Fatal("expected non-nil proxy in mixed mode")
	}
}

func TestAopIntegration_CreateProxy_GeneratedMode(t *testing.T) {
	t.Parallel()

	config := &AopConfig{
		Mode:        AopModeGenerated,
		Weaver:      NewWeaver(),
		EnableCache: true,
	}
	integration := NewAopIntegration(config)

	type GenBean struct{}
	bean := &GenBean{}

	// No generated proxy registered, should fallback
	proxy := integration.CreateProxy("genBean", bean)
	if proxy == nil {
		t.Fatal("expected non-nil result")
	}
}



package aop

import (
	"reflect"
	"testing"
)

func TestDefaultAopConfig(t *testing.T) {
	t.Parallel()

	config := DefaultAopConfig()
	if config == nil {
		t.Fatal("DefaultAopConfig() returned nil")
	}
	if config.Mode != AopModeMixed {
		t.Errorf("Mode = %v, want %v", config.Mode, AopModeMixed)
	}
	if config.Weaver == nil {
		t.Error("Weaver should not be nil")
	}
	if !config.EnableCache {
		t.Error("EnableCache should be true by default")
	}
}

func TestAopManager_GetConfig_NilConfig(t *testing.T) {
	t.Parallel()

	manager := &AopManager{aspects: make([]*AspectMeta, 0)}
	config := manager.GetConfig()
	if config != nil {
		t.Errorf("GetConfig() should return nil for nil config, got %v", config)
	}
}

func TestAopManager_GetConfig_ReturnsCopy(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}
	config1 := manager.GetConfig()
	config2 := manager.GetConfig()

	config1.EnableCache = false
	if config2.EnableCache != true {
		t.Error("GetConfig() should return independent copies")
	}
}

func TestAopManager_MatchAspectsForType_MultipleAspects(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}

	pc1 := MatchByClassName("TestUserService")
	pc2 := MatchByClassName("PointCutTestService")
	pc3 := MatchByClassName("NonExistent")

	manager.RegisterAspects(
		&AspectMeta{PointCut: pc1, Order: 1},
		&AspectMeta{PointCut: pc2, Order: 2},
		&AspectMeta{PointCut: pc3, Order: 3},
	)

	matched := manager.MatchAspectsForType(reflect.TypeOf(TestUserService{}))
	if len(matched) != 1 {
		t.Errorf("expected 1 matched aspect, got %d", len(matched))
	}
}

func TestAopManager_GetAspects_ReturnsCopy(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}
	manager.RegisterAspect(&AspectMeta{Order: 1})

	aspects := manager.GetAspects()
	if len(aspects) != 1 {
		t.Fatalf("expected 1 aspect, got %d", len(aspects))
	}

	// Modifying returned slice should not affect manager
	aspects[0] = &AspectMeta{Order: 999}
	if manager.GetAspects()[0].Order != 1 {
		t.Error("GetAspects() should return independent copy")
	}
}

func TestAopManager_GetAspects_Empty(t *testing.T) {
	t.Parallel()

	manager := &AopManager{
		config:  DefaultAopConfig(),
		aspects: make([]*AspectMeta, 0),
	}
	aspects := manager.GetAspects()
	if len(aspects) != 0 {
		t.Errorf("expected 0 aspects, got %d", len(aspects))
	}
}

func TestGlobalAopManager_Exists(t *testing.T) {
	t.Parallel()

	if GlobalAopManager == nil {
		t.Error("GlobalAopManager should not be nil")
	}
}

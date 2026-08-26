package boot

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/condition"
	"github.com/xudefa/enhance/lifecycle"
)

// ==================== SimpleFailureAnalyzer 测试 ====================

func TestNewSimpleFailureAnalyzer(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		if errors.Is(err, errors.New("test")) {
			return &FailureReport{
				Headline:    "Test Error",
				Description: "A test error occurred",
				Action:      "Do nothing",
			}
		}
		return nil
	})

	if analyzer == nil {
		t.Fatal("expected non-nil analyzer")
	}
}

func TestSimpleFailureAnalyzer_CanAnalyze_NoCheckFn(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		return &FailureReport{Description: "analyzed"}
	})

	if !analyzer.CanAnalyze(errors.New("any")) {
		t.Error("expected CanAnalyze to return true when analyzeFn returns non-nil")
	}
}

func TestSimpleFailureAnalyzer_CanAnalyze_NoCheckFn_NilReport(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		return nil
	})

	if analyzer.CanAnalyze(errors.New("any")) {
		t.Error("expected CanAnalyze to return false when analyzeFn returns nil")
	}
}

func TestSimpleFailureAnalyzer_Analyze(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		return &FailureReport{
			Headline:    "Test Failure",
			Description: err.Error(),
			Action:      "Fix it",
		}
	})

	report := analyzer.Analyze(errors.New("something broke"))
	if report == nil {
		t.Fatal("expected non-nil report")
	}
	if report.Description != "something broke" {
		t.Errorf("expected 'something broke', got '%s'", report.Description)
	}
}

func TestNewSimpleFailureAnalyzerWithCheck(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzerWithCheck(
		func(err error) bool {
			return err.Error() == "match"
		},
		func(err error) *FailureReport {
			return &FailureReport{Description: "matched"}
		},
	)

	if !analyzer.CanAnalyze(errors.New("match")) {
		t.Error("expected CanAnalyze to return true for matching error")
	}
	if analyzer.CanAnalyze(errors.New("no match")) {
		t.Error("expected CanAnalyze to return false for non-matching error")
	}
}

func TestSimpleFailureAnalyzerWithCheck_Analyze(t *testing.T) {
	t.Parallel()
	analyzer := NewSimpleFailureAnalyzerWithCheck(
		func(err error) bool { return true },
		func(err error) *FailureReport {
			return &FailureReport{Description: "analyzed: " + err.Error()}
		},
	)

	report := analyzer.Analyze(errors.New("test"))
	if report == nil || report.Description != "analyzed: test" {
		t.Errorf("expected 'analyzed: test', got %v", report)
	}
}

// ==================== BootConfig Options 测试 ====================

func TestWithExclude(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithExclude("DatabaseAutoConfig", "CacheAutoConfig")
	opt(cfg)

	if len(cfg.ExcludedAutoConfigs) != 2 {
		t.Fatalf("expected 2 excluded configs, got %d", len(cfg.ExcludedAutoConfigs))
	}
	if cfg.ExcludedAutoConfigs[0] != "DatabaseAutoConfig" {
		t.Errorf("expected 'DatabaseAutoConfig', got '%s'", cfg.ExcludedAutoConfigs[0])
	}
}

func TestWithProperty(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithProperty("app.debug", "true")
	opt(cfg)

	if len(cfg.CustomPropertySources) != 1 {
		t.Fatalf("expected 1 property source, got %d", len(cfg.CustomPropertySources))
	}
}

func TestWithProperties(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithProperties("key1", "val1", "key2", 42)
	opt(cfg)

	if len(cfg.CustomPropertySources) != 1 {
		t.Fatalf("expected 1 property source, got %d", len(cfg.CustomPropertySources))
	}
}

func TestWithProperties_OddArgs_Panics(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for odd number of args")
		}
	}()

	opt := WithProperties("key1")
	opt(cfg)
}

func TestWithProperties_NonStringKey_Panics(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for non-string key")
		}
	}()

	opt := WithProperties(123, "val")
	opt(cfg)
}

func TestWithConfigCenter(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithConfigCenter("nacos", []string{"127.0.0.1:8848"})
	opt(cfg)

	if !cfg.ConfigCenterEnabled {
		t.Error("expected ConfigCenterEnabled to be true")
	}
	if cfg.ConfigCenterType != "nacos" {
		t.Errorf("expected 'nacos', got '%s'", cfg.ConfigCenterType)
	}
	if cfg.ConfigCenterTimeout != 5*time.Second {
		t.Errorf("expected 5s timeout, got %v", cfg.ConfigCenterTimeout)
	}
}

func TestWithConfigCenter_WithOptions(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	opt := WithConfigCenter("etcd", []string{"localhost:2379"},
		WithConfigCenterDataID("app-config"),
		WithConfigCenterGroup("DEFAULT_GROUP"),
		WithConfigCenterPrefix("/config/"),
		WithConfigCenterTimeout(10*time.Second),
	)
	opt(cfg)

	if cfg.ConfigCenterDataID != "app-config" {
		t.Errorf("expected DataID 'app-config', got '%s'", cfg.ConfigCenterDataID)
	}
	if cfg.ConfigCenterGroup != "DEFAULT_GROUP" {
		t.Errorf("expected Group 'DEFAULT_GROUP', got '%s'", cfg.ConfigCenterGroup)
	}
	if cfg.ConfigCenterPrefix != "/config/" {
		t.Errorf("expected Prefix '/config/', got '%s'", cfg.ConfigCenterPrefix)
	}
	if cfg.ConfigCenterTimeout != 10*time.Second {
		t.Errorf("expected 10s timeout, got %v", cfg.ConfigCenterTimeout)
	}
}

func TestWithHook(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	hook := lifecycle.NewHookFunc(
		func(ctx context.Context) error { return nil },
		nil,
		nil,
	)
	opt := WithHook(hook)
	opt(cfg)

	if len(cfg.Hooks) != 1 {
		t.Fatalf("expected 1 hook, got %d", len(cfg.Hooks))
	}
}

// ==================== Report Print 测试 ====================

func TestConditionEvaluationReport_Print(t *testing.T) {
	t.Parallel()
	report := NewConditionEvaluationReport()
	report.RecordPositiveMatch("TestConfig", []ConditionResult{
		{Condition: "OnProperty", Matched: true, Message: "test matched"},
	})

	// Print 不 panic 即可
	report.Print()
}

// ==================== AutoConfigReport 全局报告实例获取 ====================

func TestGetAutoConfigReport_NonNil(t *testing.T) {
	t.Parallel()
	report := GetAutoConfigReport()
	if report == nil {
		t.Fatal("GetAutoConfigReport should return non-nil")
	}
}

// ==================== RegisterAutoConfig（全局注册表） ====================

func TestRegisterAutoConfig_Global(t *testing.T) {
	before := len(GlobalRegistry().GetAll())

	RegisterAutoConfig(&mockAutoConfig{},
		condition.OnProperty("test.register", "true"),
	)

	after := len(GlobalRegistry().GetAll())
	if after != before+1 {
		t.Fatalf("expected %d entries, got %d", before+1, after)
	}
}

// ==================== FailureAnalyzerRegistry 顺序测试 ====================

func TestFailureAnalyzerRegistry_FirstMatchWins(t *testing.T) {
	t.Parallel()
	registry := NewFailureAnalyzerRegistry()

	registry.Register(NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		return &FailureReport{Description: "first"}
	}))
	registry.Register(NewSimpleFailureAnalyzer(func(err error) *FailureReport {
		return &FailureReport{Description: "second"}
	}))

	report := registry.Analyze(errors.New("test"))
	if report == nil || report.Description != "first" {
		t.Errorf("expected first analyzer to win, got %v", report)
	}
}

// ==================== NewAutoConfigRegistry 测试 ====================

func TestNewAutoConfigRegistry_Empty(t *testing.T) {
	t.Parallel()
	registry := NewAutoConfigRegistry()
	if len(registry.GetAll()) != 0 {
		t.Error("expected empty registry")
	}
}

// ==================== ConfigCenterOption 独立使用测试 ====================

func TestWithConfigCenterDataID(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()
	opt := WithConfigCenterDataID("my-data-id")
	opt(cfg)
	if cfg.ConfigCenterDataID != "my-data-id" {
		t.Errorf("expected 'my-data-id', got '%s'", cfg.ConfigCenterDataID)
	}
}

func TestWithConfigCenterGroup(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()
	opt := WithConfigCenterGroup("PROD_GROUP")
	opt(cfg)
	if cfg.ConfigCenterGroup != "PROD_GROUP" {
		t.Errorf("expected 'PROD_GROUP', got '%s'", cfg.ConfigCenterGroup)
	}
}

func TestWithConfigCenterPrefix(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()
	opt := WithConfigCenterPrefix("/myapp/config/")
	opt(cfg)
	if cfg.ConfigCenterPrefix != "/myapp/config/" {
		t.Errorf("expected '/myapp/config/', got '%s'", cfg.ConfigCenterPrefix)
	}
}

func TestWithConfigCenterTimeout(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()
	opt := WithConfigCenterTimeout(30 * time.Second)
	opt(cfg)
	if cfg.ConfigCenterTimeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", cfg.ConfigCenterTimeout)
	}
}

// ==================== PropertySource 测试 ====================

func TestWithPropertySource(t *testing.T) {
	t.Parallel()
	cfg := defaultBootConfig()

	src := environment.NewMapPropertySource("custom", environment.PriorityNormal, map[string]any{"key": "val"})
	opt := WithPropertySource(src)
	opt(cfg)

	if len(cfg.CustomPropertySources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(cfg.CustomPropertySources))
	}
}

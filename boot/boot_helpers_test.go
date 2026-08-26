package boot

import (
	"testing"

	"github.com/xudefa/enhance/condition"
)

func TestDeduplicateStarters(t *testing.T) {
	t.Parallel()

	s1 := newMockStarter("s1")
	s2 := newMockStarter("s2")
	s3 := newMockStarter("s1") // 重复名称

	result := deduplicateStarters([]Starter{s1, s2, s3})

	if len(result) != 2 {
		t.Errorf("expected 2 starters after dedup, got %d", len(result))
	}
	// 应保留第一个 s1，而不是 s3
	if result[0].Name() != "s1" {
		t.Errorf("expected first starter name 's1', got %v", result[0].Name())
	}
	if result[1].Name() != "s2" {
		t.Errorf("expected second starter name 's2', got %v", result[1].Name())
	}
}

func TestDeduplicateStarters_Empty(t *testing.T) {
	t.Parallel()

	result := deduplicateStarters([]Starter{})
	if len(result) != 0 {
		t.Errorf("expected 0 starters, got %d", len(result))
	}
}

func TestDeduplicateStarters_AllSame(t *testing.T) {
	t.Parallel()

	s1 := newMockStarter("same")
	s2 := newMockStarter("same")

	result := deduplicateStarters([]Starter{s1, s2})
	if len(result) != 1 {
		t.Errorf("expected 1 starter, got %d", len(result))
	}
}

func TestStarterMatches_NoCondition(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	s := newMockStarter("test")
	if !boot.starterMatches(s) {
		t.Error("starter with no condition should match")
	}
}

func TestStarterMatches_WithCondition(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 使用 nil 条件，应该匹配
	s := newMockStarter("test")
	if !boot.starterMatches(s) {
		t.Error("starter with no condition should match")
	}
}

func TestStarterMatches_ConditionNotMatch(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 条件要求 app.name = "other"，但实际是 "test"，不应匹配
	s := newMockStarterWithCondition("test", condition.OnProperty("app.name", "other"))
	if boot.starterMatches(s) {
		t.Error("starter with non-matching condition should not match")
	}
}

func TestModuleMatches_NoConditions(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	mod := Module{moduleName: "test"}
	if !boot.moduleMatches(mod) {
		t.Error("module with no conditions should match")
	}
}

func TestModuleMatches_AllConditionsMatch(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	// 使用空条件列表，应该匹配
	mod := Module{
		moduleName: "test",
		conditions: []condition.Condition{},
	}
	if !boot.moduleMatches(mod) {
		t.Error("module with empty conditions should match")
	}
}

func TestModuleMatches_ConditionNotMatch(t *testing.T) {
	t.Parallel()

	boot, err := NewApplication(WithAppName("test"))
	if err != nil {
		t.Fatalf("NewApplication() error = %v", err)
	}

	mod := Module{
		moduleName: "test",
		conditions: []condition.Condition{
			condition.OnProperty("nonexistent.key", "value"),
		},
	}
	if boot.moduleMatches(mod) {
		t.Error("module with non-matching condition should not match")
	}
}

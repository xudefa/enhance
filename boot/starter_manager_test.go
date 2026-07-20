package boot

import (
	"testing"
)

// TestCompositeStarter 验证 CompositeStarter 的组合功能:
//  1. 多个 Starter 组合
//  2. 依赖合并
//  3. 可选标记
func TestCompositeStarter(t *testing.T) {
	t.Parallel()
	httpStarter := StarterModule{
		Name:         "http",
		Dependencies: []string{"config"},
		Optional:     false,
	}

	validationStarter := StarterModule{
		Name:         "validation",
		Dependencies: []string{"http"},
		Optional:     true,
	}

	exceptionStarter := StarterModule{
		Name:         "exception",
		Dependencies: []string{},
		Optional:     false,
	}

	composite := CompositeStarter(httpStarter, validationStarter, exceptionStarter)

	if composite.Name != "http" {
		t.Errorf("expected name 'http', got '%s'", composite.Name)
	}

	// Dependencies: http的["config"] + validation的["http"] + exception的[]
	if len(composite.Dependencies) != 2 {
		t.Errorf("expected 2 dependencies, got %d", len(composite.Dependencies))
	}

	if !composite.Optional {
		t.Error("expected composite to be optional (one of them is optional)")
	}
}

// TestDetectConflicts 验证冲突检测功能:
//  1. 重复名称检测
//  2. 缺失依赖检测
//  3. 循环依赖检测
func TestDetectConflicts(t *testing.T) {
	t.Parallel()
	// 测试重复名称
	starters1 := []StarterModule{
		{Name: "http", Dependencies: []string{}},
		{Name: "http", Dependencies: []string{}},
	}
	conflicts1 := DetectConflicts(starters1)
	if len(conflicts1) == 0 {
		t.Error("expected conflict for duplicate names")
	}

	// 测试缺失依赖
	starters2 := []StarterModule{
		{Name: "http", Dependencies: []string{"nonexistent"}},
	}
	conflicts2 := DetectConflicts(starters2)
	if len(conflicts2) == 0 {
		t.Error("expected conflict for missing dependency")
	}

	// 测试循环依赖
	starters3 := []StarterModule{
		{Name: "a", Dependencies: []string{"b"}},
		{Name: "b", Dependencies: []string{"c"}},
		{Name: "c", Dependencies: []string{"a"}},
	}
	conflicts3 := DetectConflicts(starters3)
	if len(conflicts3) == 0 {
		t.Error("expected conflict for circular dependency")
	}

	// 测试无冲突
	starters4 := []StarterModule{
		{Name: "config", Dependencies: []string{}},
		{Name: "http", Dependencies: []string{"config"}},
		{Name: "web", Dependencies: []string{"http"}},
	}
	conflicts4 := DetectConflicts(starters4)
	if len(conflicts4) != 0 {
		t.Errorf("expected no conflicts, got %d", len(conflicts4))
	}
}

// TestValidateStarters 验证 Starter 验证功能:
//  1. 有效 Starter 列表
//  2. 无效 Starter 列表
func TestValidateStarters(t *testing.T) {
	t.Parallel()
	// 有效的 Starter 列表
	validStarters := []StarterModule{
		{Name: "config", Dependencies: []string{}},
		{Name: "http", Dependencies: []string{"config"}},
	}
	err := ValidateStarters(validStarters)
	if err != nil {
		t.Errorf("expected no error for valid starters, got: %v", err)
	}

	// 无效的 Starter 列表
	invalidStarters := []StarterModule{
		{Name: "http", Dependencies: []string{"nonexistent"}},
	}
	err = ValidateStarters(invalidStarters)
	if err == nil {
		t.Error("expected error for invalid starters")
	}
}

// TestResolveDependencies 验证依赖解析功能:
//  1. 正确的拓扑排序
//  2. 循环依赖错误
func TestResolveDependencies(t *testing.T) {
	t.Parallel()
	starters := []StarterModule{
		{Name: "config", Dependencies: []string{}},
		{Name: "http", Dependencies: []string{"config"}},
		{Name: "web", Dependencies: []string{"http"}},
	}

	resolved, err := ResolveDependencies(starters)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(resolved) != 3 {
		t.Fatalf("expected 3 starters, got %d", len(resolved))
	}

	// 验证顺序：config -> http -> web
	order := make(map[string]int)
	for i, s := range resolved {
		order[s.Name] = i
	}

	if order["config"] >= order["http"] {
		t.Error("config should come before http")
	}
	if order["http"] >= order["web"] {
		t.Error("http should come before web")
	}

	// 测试循环依赖
	cyclicStarters := []StarterModule{
		{Name: "a", Dependencies: []string{"b"}},
		{Name: "b", Dependencies: []string{"a"}},
	}
	_, err = ResolveDependencies(cyclicStarters)
	if err == nil {
		t.Error("expected error for cyclic dependencies")
	}
}

// TestConflictString 验证冲突信息的可读性
func TestConflictString(t *testing.T) {
	t.Parallel()
	conflict := Conflict{
		StarterA: "http",
		StarterB: "tcp",
		Reason:   "端口冲突",
	}

	expected := "冲突: http <-> tcp (端口冲突)"
	if conflict.String() != expected {
		t.Errorf("expected '%s', got '%s'", expected, conflict.String())
	}
}

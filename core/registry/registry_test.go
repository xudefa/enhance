package registry

import (
	"reflect"
	"testing"
)

type TestBean struct {
	Value string
}

func TestRegisterAndGetDefinition(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def := BeanDef{
		Type:    typ,
		Name:    "testBean",
		Scope:   Singleton,
		Factory: func(c ...any) (any, error) { return &TestBean{Value: "test"}, nil },
	}

	err := reg.Register(def, "testBean")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	retrieved, ok := reg.GetDefinition("testBean")
	if !ok {
		t.Fatal("GetDefinition failed")
	}

	if retrieved.Type != typ {
		t.Errorf("Expected type %v, got %v", typ, retrieved.Type)
	}

	if retrieved.Scope != Singleton {
		t.Errorf("Expected scope %v, got %v", Singleton, retrieved.Scope)
	}
}

func TestRegisterNilType(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	def := BeanDef{
		Type: nil,
	}

	err := reg.Register(def, "nilBean")
	if err == nil {
		t.Error("Expected error for nil type")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def := BeanDef{
		Type:  typ,
		Name:  "duplicate",
		Scope: Singleton,
	}

	err := reg.Register(def, "duplicate")
	if err != nil {
		t.Fatalf("First register failed: %v", err)
	}

	// Duplicate registration should not error
	err = reg.Register(def, "duplicate")
	if err != nil {
		t.Errorf("Expected no error for duplicate, got: %v", err)
	}
}

func TestRegisterWithPrimary(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def1 := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:    typ,
		Name:    "bean2",
		Scope:   Singleton,
		Primary: true,
	}

	err := reg.Register(def1, "bean1")
	if err != nil {
		t.Fatalf("Register bean1 failed: %v", err)
	}

	err = reg.Register(def2, "bean2")
	if err != nil {
		t.Fatalf("Register bean2 failed: %v", err)
	}

	// bean2 should be primary
	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("GetPrimaryByType failed")
	}

	if primaryID != "bean2" {
		t.Errorf("Expected primary bean2, got %s", primaryID)
	}
}

func TestGetInstance(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def := BeanDef{
		Type:  typ,
		Name:  "testBean",
		Scope: Singleton,
	}

	err := reg.Register(def, "testBean")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Instance should not exist yet
	_, ok := reg.GetInstance("testBean")
	if ok {
		t.Error("Expected instance not to exist")
	}

	// Set instance
	instance := &TestBean{Value: "test"}
	reg.SetInstance("testBean", instance)

	// Now it should exist
	retrieved, ok := reg.GetInstance("testBean")
	if !ok {
		t.Fatal("GetInstance failed")
	}

	bean := retrieved.(*TestBean)
	if bean.Value != "test" {
		t.Errorf("Expected value 'test', got '%s'", bean.Value)
	}
}

func TestGetByType(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def1 := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:  typ,
		Name:  "bean2",
		Scope: Singleton,
	}

	err := reg.Register(def1, "bean1")
	if err != nil {
		t.Fatalf("Register bean1 failed: %v", err)
	}

	err = reg.Register(def2, "bean2")
	if err != nil {
		t.Fatalf("Register bean2 failed: %v", err)
	}

	ids := reg.GetByType(typ)
	if len(ids) != 2 {
		t.Errorf("Expected 2 bean IDs, got %d", len(ids))
	}
}

func TestGetDefinitionByCustomName(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	fullID := "github.com/example.TestBean#custom"
	def := BeanDef{
		Type:  typ,
		Name:  fullID,
		Scope: Singleton,
	}

	err := reg.Register(def, fullID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Lookup by custom name should work
	retrieved, ok := reg.GetDefinition("custom")
	if !ok {
		t.Fatal("GetDefinition by custom name failed")
	}

	if retrieved.Name != fullID {
		t.Errorf("Expected name %s, got %s", fullID, retrieved.Name)
	}
}

// ==================== 补充单测：提高覆盖率 ====================

func TestGetByTypeEmpty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	ids := reg.GetByType(typ)
	if ids != nil {
		t.Errorf("Expected nil for empty type, got %v", ids)
	}
}

func TestGetPrimaryByTypeNotFound(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	_, ok := reg.GetPrimaryByType(typ)
	if ok {
		t.Error("Expected no primary for non-existent type")
	}
}

func TestRegisterWithStandardFormatName(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	fullID := "github.com/registry.TestBean#myBean"
	def := BeanDef{
		Type:  typ,
		Name:  fullID,
		Scope: Singleton,
	}

	err := reg.Register(def, fullID)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// 应该能通过完整 ID 查找
	_, ok := reg.GetDefinition(fullID)
	if !ok {
		t.Error("GetDefinition by full ID failed")
	}

	// 应该能通过自定义名称查找
	_, ok = reg.GetDefinition("myBean")
	if !ok {
		t.Error("GetDefinition by custom name failed")
	}

	// HasBean 应该对两者都返回 true
	if !reg.HasBean(fullID) {
		t.Error("HasBean should return true for full ID")
	}

	if !reg.HasBean("myBean") {
		t.Error("HasBean should return true for custom name")
	}
}

func TestGetDefinitionNotFound(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	_, ok := reg.GetDefinition("nonexistent")
	if ok {
		t.Error("Expected GetDefinition to return false for non-existent bean")
	}
}

func TestGetInstanceNotFound(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	_, ok := reg.GetInstance("nonexistent")
	if ok {
		t.Error("Expected GetInstance to return false for non-existent bean")
	}
}

func TestRegisterDuplicateWithPrimary(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def1 := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:    typ,
		Name:    "bean1",
		Scope:   Singleton,
		Primary: true,
	}

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "bean1")

	// bean1 应该成为 primary
	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("GetPrimaryByType failed")
	}

	if primaryID != "bean1" {
		t.Errorf("Expected primary bean1, got %s", primaryID)
	}
}

func TestGetByTypeReturnsCorrectList(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	otherType := reflect.TypeOf((*string)(nil))

	def1 := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:  typ,
		Name:  "bean2",
		Scope: Singleton,
	}

	def3 := BeanDef{
		Type:  otherType,
		Name:  "strBean",
		Scope: Singleton,
	}

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "bean2")
	_ = reg.Register(def3, "strBean")

	ids := reg.GetByType(typ)
	if len(ids) != 2 {
		t.Fatalf("Expected 2 bean IDs, got %d", len(ids))
	}

	// 验证返回的 ID 正确
	found := false
	for _, id := range ids {
		if id == "bean1" || id == "bean2" {
			found = true
		}
	}
	if !found {
		t.Errorf("Expected to find bean1 or bean2, got %v", ids)
	}
}

// ==================== 边界情况和 Bug 修复测试 ====================

func TestGetDefinitionWithMultipleCustomNames(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 注册多个相同自定义名称的 bean（不同包）
	def1 := BeanDef{
		Type:  typ,
		Name:  "github.com/pkg1.TestBean#myBean",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:  typ,
		Name:  "github.com/pkg2.TestBean#myBean",
		Scope: Singleton,
	}

	_ = reg.Register(def1, "github.com/pkg1.TestBean#myBean")
	_ = reg.Register(def2, "github.com/pkg2.TestBean#myBean")

	// 应该能找到其中一个（取决于遍历顺序）
	_, ok := reg.GetDefinition("myBean")
	if !ok {
		t.Error("Expected to find bean by custom name")
	}
}

func TestRegisterWithPrimaryOnDuplicate(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 首次注册，非 Primary
	def1 := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}
	_ = reg.Register(def1, "bean1")

	// 重复注册，标记为 Primary
	def2 := BeanDef{
		Type:    typ,
		Name:    "bean1",
		Scope:   Singleton,
		Primary: true,
	}
	_ = reg.Register(def2, "bean1")

	// 应该成为 primary
	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("Expected primary to exist")
	}

	if primaryID != "bean1" {
		t.Errorf("Expected primary 'bean1', got '%s'", primaryID)
	}
}

func TestRegisterWithoutPrimaryWhenNoPrimaryExists(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 注册第一个 bean，非 Primary
	def := BeanDef{
		Type:  typ,
		Name:  "bean1",
		Scope: Singleton,
	}
	_ = reg.Register(def, "bean1")

	// 第一个 bean 应该自动成为 primary
	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("Expected primary to exist")
	}

	if primaryID != "bean1" {
		t.Errorf("Expected primary 'bean1', got '%s'", primaryID)
	}
}

func TestGetDefinitionWithEmptyCustomName(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	def := BeanDef{
		Type:  typ,
		Name:  "github.com/example.TestBean",
		Scope: Singleton,
	}

	_ = reg.Register(def, "github.com/example.TestBean")

	_, ok := reg.GetDefinition("github.com/example.TestBean")
	if !ok {
		t.Error("Expected to find bean by full ID")
	}

	_, ok = reg.GetDefinition("")
	if ok {
		t.Error("Expected not to find bean with empty string")
	}
}

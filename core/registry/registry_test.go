package registry

import (
	"reflect"
	"sync"
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

func TestHasBean(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def := BeanDef{
		Type:  typ,
		Name:  "github.com/example.TestBean#myBean",
		Scope: Singleton,
	}

	err := reg.Register(def, "github.com/example.TestBean#myBean")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Direct lookup
	if !reg.HasBean("github.com/example.TestBean#myBean") {
		t.Error("Expected HasBean to return true for full ID")
	}

	// Custom name lookup
	if !reg.HasBean("myBean") {
		t.Error("Expected HasBean to return true for custom name")
	}

	// Non-existent
	if reg.HasBean("nonexistent") {
		t.Error("Expected HasBean to return false for non-existent bean")
	}
}

func TestHasType(t *testing.T) {
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

	if !reg.HasType(typ) {
		t.Error("Expected HasType to return true")
	}

	otherType := reflect.TypeOf((*string)(nil))
	if reg.HasType(otherType) {
		t.Error("Expected HasType to return false for other type")
	}
}

func TestCount(t *testing.T) {
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

	if reg.Count() != 2 {
		t.Errorf("Expected count 2, got %d", reg.Count())
	}
}

func TestCountByType(t *testing.T) {
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

	reg.Register(def1, "bean1")
	reg.Register(def2, "bean2")
	reg.Register(def3, "strBean")

	if reg.CountByType(typ) != 2 {
		t.Errorf("Expected 2 TestBean, got %d", reg.CountByType(typ))
	}

	if reg.CountByType(otherType) != 1 {
		t.Errorf("Expected 1 string, got %d", reg.CountByType(otherType))
	}
}

func TestTypes(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ1 := reflect.TypeOf((*TestBean)(nil))
	typ2 := reflect.TypeOf((*string)(nil))

	def1 := BeanDef{
		Type:  typ1,
		Name:  "bean1",
		Scope: Singleton,
	}

	def2 := BeanDef{
		Type:  typ2,
		Name:  "strBean",
		Scope: Singleton,
	}

	reg.Register(def1, "bean1")
	reg.Register(def2, "strBean")

	types := reg.Types()
	if len(types) != 2 {
		t.Fatalf("Expected 2 types, got %d", len(types))
	}
}

func TestBeanIDs(t *testing.T) {
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

	reg.Register(def1, "bean1")
	reg.Register(def2, "bean2")

	ids := reg.BeanIDs()
	if len(ids) != 2 {
		t.Fatalf("Expected 2 bean IDs, got %d", len(ids))
	}
}

func TestClear(t *testing.T) {
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

	reg.SetInstance("testBean", &TestBean{Value: "test"})

	reg.Clear()

	if reg.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", reg.Count())
	}

	_, ok := reg.GetInstance("testBean")
	if ok {
		t.Error("Expected instance to be cleared")
	}
}

func TestConcurrentRegister(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			beanID := "bean" + string(rune('0'+i%10))
			def := BeanDef{
				Type:  typ,
				Name:  beanID,
				Scope: Singleton,
			}
			reg.Register(def, beanID)
		}(i)
	}
	wg.Wait()

	// Should have 10 unique beans (0-9)
	if reg.Count() != 10 {
		t.Errorf("Expected 10 beans, got %d", reg.Count())
	}
}

func TestConcurrentGetAndSet(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	def := BeanDef{
		Type:  typ,
		Name:  "testBean",
		Scope: Singleton,
	}

	reg.Register(def, "testBean")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			instance := &TestBean{Value: "test"}
			reg.SetInstance("testBean", instance)

			_, ok := reg.GetInstance("testBean")
			if !ok {
				t.Error("Expected instance to exist")
			}
		}(i)
	}
	wg.Wait()
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

func TestCountByTypeEmpty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	count := reg.CountByType(typ)
	if count != 0 {
		t.Errorf("Expected 0 for empty type, got %d", count)
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

func TestHasBeanWithStandardFormat(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	fullID := "github.com/example.TestBean#custom"
	def := BeanDef{
		Type:  typ,
		Name:  fullID,
		Scope: Singleton,
	}

	reg.Register(def, fullID)

	// 测试部分匹配
	if !reg.HasBean("custom") {
		t.Error("Expected HasBean to return true for custom name")
	}
}

func TestClearWithMultipleBeans(t *testing.T) {
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

	reg.Register(def1, "bean1")
	reg.Register(def2, "bean2")

	reg.SetInstance("bean1", &TestBean{Value: "1"})
	reg.SetInstance("bean2", &TestBean{Value: "2"})

	reg.Clear()

	// 验证所有数据都被清空
	if reg.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", reg.Count())
	}

	_, ok := reg.GetInstance("bean1")
	if ok {
		t.Error("Expected bean1 instance to be cleared")
	}

	_, ok = reg.GetInstance("bean2")
	if ok {
		t.Error("Expected bean2 instance to be cleared")
	}

	_, ok = reg.GetDefinition("bean1")
	if ok {
		t.Error("Expected bean1 definition to be cleared")
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

	reg.Register(def1, "bean1")
	reg.Register(def2, "bean1")

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

	reg.Register(def1, "bean1")
	reg.Register(def2, "bean2")
	reg.Register(def3, "strBean")

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

// ==================== BeanIDGenerator 测试 ====================

func TestGenerateBeanID(t *testing.T) {
	t.Parallel()
	gen := &defaultBeanIDGenerator{}

	// 测试无自定义名称
	id := gen.Generate("github.com/example", "TestBean", "")
	expected := "github.com/example.TestBean"
	if id != expected {
		t.Errorf("Expected %s, got %s", expected, id)
	}

	// 测试有自定义名称
	id = gen.Generate("github.com/example", "TestBean", "myBean")
	expected = "github.com/example.TestBean#myBean"
	if id != expected {
		t.Errorf("Expected %s, got %s", expected, id)
	}
}

func TestParseBeanID(t *testing.T) {
	t.Parallel()
	gen := &defaultBeanIDGenerator{}

	// 测试完整格式
	pkg, typ, name := gen.Parse("github.com/example.TestBean#myBean")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "TestBean" {
		t.Errorf("Expected type 'TestBean', got '%s'", typ)
	}
	if name != "myBean" {
		t.Errorf("Expected name 'myBean', got '%s'", name)
	}

	// 测试无自定义名称
	pkg, typ, name = gen.Parse("github.com/example.TestBean")
	if pkg != "github.com/example" {
		t.Errorf("Expected pkg 'github.com/example', got '%s'", pkg)
	}
	if typ != "TestBean" {
		t.Errorf("Expected type 'TestBean', got '%s'", typ)
	}
	if name != "" {
		t.Errorf("Expected empty name, got '%s'", name)
	}

	// 测试无包路径
	pkg, typ, name = gen.Parse("TestBean")
	if pkg != "" {
		t.Errorf("Expected empty pkg, got '%s'", pkg)
	}
	if typ != "TestBean" {
		t.Errorf("Expected type 'TestBean', got '%s'", typ)
	}
}

func TestStringBeanID(t *testing.T) {
	t.Parallel()
	gen := &defaultBeanIDGenerator{}

	// 测试有自定义名称
	result := gen.String("github.com/example.TestBean#myBean")
	expected := "github.com/example.TestBean#myBean"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}

	// 测试无自定义名称
	result = gen.String("github.com/example.TestBean")
	expected = "github.com/example.TestBean"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
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

	reg.Register(def1, "github.com/pkg1.TestBean#myBean")
	reg.Register(def2, "github.com/pkg2.TestBean#myBean")

	// 应该能找到其中一个（取决于遍历顺序）
	_, ok := reg.GetDefinition("myBean")
	if !ok {
		t.Error("Expected to find bean by custom name")
	}
}

func TestHasBeanWithMultipleMatches(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 注册多个相同自定义名称的 bean
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

	reg.Register(def1, "github.com/pkg1.TestBean#myBean")
	reg.Register(def2, "github.com/pkg2.TestBean#myBean")

	// HasBean 应该返回 true
	if !reg.HasBean("myBean") {
		t.Error("Expected HasBean to return true")
	}
}

func TestBeanIDsOrder(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 按特定顺序注册
	reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")
	reg.Register(BeanDef{Type: typ, Name: "bean2", Scope: Singleton}, "bean2")
	reg.Register(BeanDef{Type: typ, Name: "bean3", Scope: Singleton}, "bean3")

	ids := reg.BeanIDs()

	// 验证顺序
	if len(ids) != 3 {
		t.Fatalf("Expected 3 bean IDs, got %d", len(ids))
	}

	if ids[0] != "bean1" || ids[1] != "bean2" || ids[2] != "bean3" {
		t.Errorf("Expected order [bean1, bean2, bean3], got %v", ids)
	}
}

func TestBeanIDsReturnsCopy(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")

	ids1 := reg.BeanIDs()
	ids1[0] = "modified"

	ids2 := reg.BeanIDs()
	if ids2[0] == "modified" {
		t.Error("BeanIDs should return a copy, not the original slice")
	}
}

func TestClearAfterRegister(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")
	reg.SetInstance("bean1", &TestBean{Value: "test"})

	// 验证注册成功
	if reg.Count() != 1 {
		t.Fatalf("Expected count 1, got %d", reg.Count())
	}

	// 清空
	reg.Clear()

	// 验证清空后状态
	if reg.Count() != 0 {
		t.Errorf("Expected count 0 after clear, got %d", reg.Count())
	}

	if reg.HasBean("bean1") {
		t.Error("Expected HasBean to return false after clear")
	}

	if reg.HasType(typ) {
		t.Error("Expected HasType to return false after clear")
	}

	_, ok := reg.GetPrimaryByType(typ)
	if ok {
		t.Error("Expected GetPrimaryByType to return false after clear")
	}

	ids := reg.BeanIDs()
	if len(ids) != 0 {
		t.Errorf("Expected 0 bean IDs after clear, got %d", len(ids))
	}

	types := reg.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types after clear, got %d", len(types))
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
	reg.Register(def1, "bean1")

	// 重复注册，标记为 Primary
	def2 := BeanDef{
		Type:    typ,
		Name:    "bean1",
		Scope:   Singleton,
		Primary: true,
	}
	reg.Register(def2, "bean1")

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
	reg.Register(def, "bean1")

	// 第一个 bean 应该自动成为 primary
	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("Expected primary to exist")
	}

	if primaryID != "bean1" {
		t.Errorf("Expected primary 'bean1', got '%s'", primaryID)
	}
}

func TestConcurrentBeanIDs(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			beanID := "bean" + string(rune('0'+i%10))
			def := BeanDef{
				Type:  typ,
				Name:  beanID,
				Scope: Singleton,
			}
			reg.Register(def, beanID)
		}(i)
	}
	wg.Wait()

	// BeanIDs 应该能正常返回
	ids := reg.BeanIDs()
	if len(ids) != 10 {
		t.Errorf("Expected 10 bean IDs, got %d", len(ids))
	}
}

func TestTypesAfterClear(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")

	// 验证 types 存在
	types := reg.Types()
	if len(types) != 1 {
		t.Fatalf("Expected 1 type, got %d", len(types))
	}

	// 清空
	reg.Clear()

	// 验证 types 被清空
	types = reg.Types()
	if len(types) != 0 {
		t.Errorf("Expected 0 types after clear, got %d", len(types))
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

	reg.Register(def, "github.com/example.TestBean")

	_, ok := reg.GetDefinition("github.com/example.TestBean")
	if !ok {
		t.Error("Expected to find bean by full ID")
	}

	_, ok = reg.GetDefinition("")
	if ok {
		t.Error("Expected not to find bean with empty string")
	}
}

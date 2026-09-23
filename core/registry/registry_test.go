package registry

import (
	"errors"
	"reflect"
	"sync"
	"testing"
)

type TestBean struct {
	Value string
}

// ==================== 注册和获取测试 ====================

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
	def := BeanDef{
		Type:    typ,
		Name:    "primaryBean",
		Scope:   Singleton,
		Primary: true,
	}

	err := reg.Register(def, "primaryBean")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("Expected primary to exist")
	}

	if primaryID != "primaryBean" {
		t.Errorf("Expected primary 'primaryBean', got %q", primaryID)
	}
}

func TestRegister_DifferentDefinition_ErrorContainsBeanID(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.Register(BeanDef{Type: typ, Name: "a", Scope: Singleton}, "svc")
	err := reg.Register(BeanDef{Type: typ, Name: "b", Scope: Prototype}, "svc")
	if err == nil {
		t.Fatal("expected error for different definition")
	}
	if !errors.Is(err, ErrBeanAlreadyExists) {
		t.Errorf("expected ErrBeanAlreadyExists, got %v", err)
	}
}

func TestRegisterInstance_DuplicateID(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.RegisterInstance(&TestBean{Value: "first"}, typ, "myBean")
	err := reg.RegisterInstance(&TestBean{Value: "second"}, typ, "myBean")
	if err == nil {
		t.Error("expected error for duplicate instance registration")
	}
}

func TestRegisterInstance_SetsPrimary(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.RegisterInstance(&TestBean{Value: "inst"}, typ, "myBean")

	primaryID, ok := reg.GetPrimaryByType(typ)
	if !ok {
		t.Fatal("expected primary to exist")
	}
	if primaryID != "myBean" {
		t.Errorf("expected primary 'myBean', got %q", primaryID)
	}

	def, ok := reg.GetDefinition("myBean")
	if !ok {
		t.Fatal("expected definition to exist")
	}
	if def.Factory == nil {
		t.Error("expected Factory to be set")
	}
	if def.Scope != Singleton {
		t.Errorf("Scope = %q, want %q", def.Scope, Singleton)
	}
}

func TestRegisterInstance_FactoryReturnsInstance(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	original := &TestBean{Value: "the-instance"}

	_ = reg.RegisterInstance(original, typ, "myBean")

	def, _ := reg.GetDefinition("myBean")
	instance, err := def.Factory()
	if err != nil {
		t.Fatalf("Factory: %v", err)
	}
	if instance != original {
		t.Error("Factory should return the original instance")
	}
}

func TestRegisterInstance_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(reg BeanRegistry)
		instance  any
		typ       reflect.Type
		wantErr   bool
		checkFunc func(t *testing.T, reg BeanRegistry)
	}{
		{
			name:     "success registers primary singleton",
			instance: &TestBean{Value: "inst"},
			typ:      reflect.TypeOf((*TestBean)(nil)),
			checkFunc: func(t *testing.T, reg BeanRegistry) {
				checkRegisterInstanceSuccess(t, reg)
			},
		},
		{
			name: "duplicate instance id returns error",
			setup: func(reg BeanRegistry) {
				if err := reg.RegisterInstance(&TestBean{Value: "first"}, reflect.TypeOf((*TestBean)(nil)), "id-inst"); err != nil {
					t.Errorf("setup register failed: %v", err)
				}
			},
			instance: &TestBean{Value: "second"},
			typ:      reflect.TypeOf((*TestBean)(nil)),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg := NewBeanRegistry()
			if tt.setup != nil {
				tt.setup(reg)
			}
			err := reg.RegisterInstance(tt.instance, tt.typ, "id-inst")
			if (err != nil) != tt.wantErr {
				t.Fatalf("RegisterInstance error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, reg)
			}
		})
	}
}

func checkRegisterInstanceSuccess(t *testing.T, reg BeanRegistry) {
	t.Helper()
	def, ok := reg.GetDefinition("id-inst")
	if !ok {
		t.Fatal("expected definition to exist")
	}
	if !def.Primary {
		t.Error("expected instance bean to be primary")
	}
	if def.Scope != Singleton {
		t.Errorf("expected singleton scope, got %v", def.Scope)
	}
	bean, err := def.Factory()
	if err != nil {
		t.Fatalf("factory failed: %v", err)
	}
	got, ok := bean.(*TestBean)
	if !ok || got.Value != "inst" {
		t.Errorf("unexpected factory result: %v", bean)
	}
	primaryID, ok := reg.GetPrimaryByType(reflect.TypeOf((*TestBean)(nil)))
	if !ok || primaryID != "id-inst" {
		t.Errorf("expected id-inst as primary, got %q", primaryID)
	}
}

// ==================== Has 测试 ====================

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

	_ = reg.Register(def, fullID)

	// 测试部分匹配
	if !reg.HasBean("custom") {
		t.Error("Expected HasBean to return true for custom name")
	}
}

// ==================== 计数和列表测试 ====================

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

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "bean2")
	_ = reg.Register(def3, "strBean")

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

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "strBean")

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

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "bean2")

	ids := reg.BeanIDs()
	if len(ids) != 2 {
		t.Fatalf("Expected 2 bean IDs, got %d", len(ids))
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

func TestBeanIDsOrder(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	// 按特定顺序注册
	_ = reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")
	_ = reg.Register(BeanDef{Type: typ, Name: "bean2", Scope: Singleton}, "bean2")
	_ = reg.Register(BeanDef{Type: typ, Name: "bean3", Scope: Singleton}, "bean3")

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

	_ = reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")

	ids1 := reg.BeanIDs()
	ids1[0] = "modified"

	ids2 := reg.BeanIDs()
	if ids2[0] == "modified" {
		t.Error("BeanIDs should return a copy, not the original slice")
	}
}

func TestTypesAfterClear(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")

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

// ==================== 清除测试 ====================

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

	_ = reg.Register(def1, "bean1")
	_ = reg.Register(def2, "bean2")

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

// ==================== 列表和实例测试 ====================

func TestListBeansAndInstances_Coverage(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	_ = reg.Register(BeanDef{Type: typ, Name: "b1", Scope: Singleton}, "b1")
	_ = reg.Register(BeanDef{Type: typ, Name: "b2", Scope: Prototype}, "b2")
	reg.SetInstance("b1", &TestBean{Value: "i1"})
	reg.SetInstance("b2", &TestBean{Value: "i2"})

	beans := reg.ListBeans()
	if len(beans) != 2 {
		t.Fatalf("expected 2 beans, got %d", len(beans))
	}
	if beans["b1"].Name != "b1" || beans["b2"].Scope != Prototype {
		t.Errorf("unexpected bean definitions: %+v", beans)
	}

	instances := reg.ListInstances()
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if v, ok := instances["b1"].(*TestBean); !ok || v.Value != "i1" {
		t.Errorf("unexpected instance for b1: %v", instances["b1"])
	}

	beans["b1"].Name = "mutated"
	delete(beans, "b2")

	if again := reg.ListBeans(); again["b1"].Name != "b1" || len(again) != 2 {
		t.Error("ListBeans must return copies, mutations must not leak")
	}

	instances["b1"] = nil
	delete(instances, "b2")

	if again := reg.ListInstances(); len(again) != 2 {
		t.Error("ListInstances must return a snapshot copy")
	}
}

func TestRegisterDifferentDefinitionReturnsErrAlreadyExists_Coverage(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	first := BeanDef{Type: typ, Name: "svc", Scope: Singleton}
	second := BeanDef{Type: typ, Name: "svc", Scope: Prototype}

	if err := reg.Register(first, "svc"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	err := reg.Register(second, "svc")
	if !errors.Is(err, ErrBeanAlreadyExists) {
		t.Fatalf("expected ErrBeanAlreadyExists, got %v", err)
	}
}

func TestRegisterCustomNameConflict_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		firstID    string
		firstName  string
		secondID   string
		secondName string
		wantErr    bool
	}{
		{"plain name conflict", "id-1", "alpha", "id-2", "alpha", true},
		{"suffix conflict via #name", "id-1", "pkg1.Svc#shared", "id-2", "pkg2.Svc#shared", true},
		{"distinct names ok", "id-1", "pkg1.Svc#a", "id-2", "pkg2.Svc#b", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg := NewBeanRegistry()
			typ := reflect.TypeOf((*TestBean)(nil))

			if err := reg.Register(BeanDef{Type: typ, Name: tt.firstName}, tt.firstID); err != nil {
				t.Fatalf("first register failed: %v", err)
			}
			err := reg.Register(BeanDef{Type: typ, Name: tt.secondName}, tt.secondID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFuncPtr_Coverage(t *testing.T) {
	t.Parallel()

	var nilFn func(bean any) error
	fn := func(bean any) error { return nil }

	tests := []struct {
		name string
		in   any
		want uintptr
	}{
		{"nil interface", nil, 0},
		{"non-func kind", "not-a-func", 0},
		{"nil func value", nilFn, 0},
		{"valid func", fn, reflect.ValueOf(fn).Pointer()},
		{"same func twice equal", fn, reflect.ValueOf(fn).Pointer()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := funcPtr(tt.in); got != tt.want {
				t.Errorf("funcPtr() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeScope_Coverage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Scope
		want Scope
	}{
		{"empty maps to singleton", "", Singleton},
		{"singleton kept", Singleton, Singleton},
		{"prototype kept", Prototype, Prototype},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeScope(tt.in); got != tt.want {
				t.Errorf("normalizeScope(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSameBeanDefinitionIgnoresEmptyScopeAndFactory_Coverage(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	if err := reg.Register(BeanDef{Type: typ}, "svc"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := reg.Register(BeanDef{Type: typ, Scope: Singleton}, "svc"); err != nil {
		t.Errorf("empty scope should equal singleton, got %v", err)
	}
}

// ==================== 并发测试 ====================

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
			_ = reg.Register(def, beanID)
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

	_ = reg.Register(def, "testBean")

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
			_ = reg.Register(def, beanID)
		}(i)
	}
	wg.Wait()

	// BeanIDs 应该能正常返回
	ids := reg.BeanIDs()
	if len(ids) != 10 {
		t.Errorf("Expected 10 bean IDs, got %d", len(ids))
	}
}

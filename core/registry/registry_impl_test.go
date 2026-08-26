package registry

import (
	"errors"
	"reflect"
	"testing"
)

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
	result, err := def.Factory()
	if err != nil {
		t.Fatalf("Factory: %v", err)
	}
	if result != original {
		t.Error("Factory should return the original instance")
	}
}

func TestListBeans_Empty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	beans := reg.ListBeans()
	if len(beans) != 0 {
		t.Errorf("expected empty map, got %d", len(beans))
	}
}

func TestListInstances_Empty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	instances := reg.ListInstances()
	if len(instances) != 0 {
		t.Errorf("expected empty map, got %d", len(instances))
	}
}

func TestTypes_Empty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	types := reg.Types()
	if len(types) != 0 {
		t.Errorf("expected empty, got %d", len(types))
	}
}

func TestBeanIDs_Empty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	ids := reg.BeanIDs()
	if len(ids) != 0 {
		t.Errorf("expected empty, got %d", len(ids))
	}
}

func TestCount_Empty(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	if reg.Count() != 0 {
		t.Errorf("expected 0, got %d", reg.Count())
	}
}

func TestHasBean_Nonexistent(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	if reg.HasBean("anything") {
		t.Error("expected false for nonexistent bean")
	}
}

func TestHasType_Nonexistent(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	if reg.HasType(reflect.TypeOf((*TestBean)(nil))) {
		t.Error("expected false for nonexistent type")
	}
}

func TestGetDefinition_Nonexistent(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	_, ok := reg.GetDefinition("nope")
	if ok {
		t.Error("expected false for nonexistent bean")
	}
}

func TestGetDefinition_ByCustomName(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	fullID := "github.com/test.TestBean#myBean"
	_ = reg.Register(BeanDef{Type: typ, Name: fullID, Scope: Singleton}, fullID)

	def, ok := reg.GetDefinition("myBean")
	if !ok {
		t.Fatal("expected to find by custom name")
	}
	if def.Name != fullID {
		t.Errorf("Name = %q, want %q", def.Name, fullID)
	}
}

func TestHasBean_ByCustomName(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	fullID := "github.com/test.TestBean#svc"
	_ = reg.Register(BeanDef{Type: typ, Name: fullID, Scope: Singleton}, fullID)

	if !reg.HasBean("svc") {
		t.Error("expected HasBean true for custom name suffix")
	}
}

func TestGetByType_ReturnsCopy(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	_ = reg.Register(BeanDef{Type: typ, Name: "b1", Scope: Singleton}, "b1")
	_ = reg.Register(BeanDef{Type: typ, Name: "b2", Scope: Singleton}, "b2")

	ids := reg.GetByType(typ)
	ids[0] = "mutated"

	ids2 := reg.GetByType(typ)
	for _, id := range ids2 {
		if id == "mutated" {
			t.Error("GetByType should return a copy")
		}
	}
}

func TestBeanIDs_ReturnsInsertOrder(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	_ = reg.Register(BeanDef{Type: typ, Name: "c"}, "c")
	_ = reg.Register(BeanDef{Type: typ, Name: "a"}, "a")
	_ = reg.Register(BeanDef{Type: typ, Name: "b"}, "b")

	ids := reg.BeanIDs()
	if len(ids) != 3 || ids[0] != "c" || ids[1] != "a" || ids[2] != "b" {
		t.Errorf("unexpected order: %v", ids)
	}
}

func TestClear_ResetsAllState(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	_ = reg.Register(BeanDef{Type: typ, Name: "b1", Scope: Singleton}, "b1")
	reg.SetInstance("b1", &TestBean{Value: "x"})

	reg.Clear()

	if reg.Count() != 0 {
		t.Errorf("Count = %d after Clear", reg.Count())
	}
	if reg.HasBean("b1") {
		t.Error("HasBean should be false after Clear")
	}
	if reg.HasType(typ) {
		t.Error("HasType should be false after Clear")
	}
	_, ok := reg.GetDefinition("b1")
	if ok {
		t.Error("GetDefinition should return false after Clear")
	}
	_, ok = reg.GetInstance("b1")
	if ok {
		t.Error("GetInstance should return false after Clear")
	}
	if len(reg.Types()) != 0 {
		t.Error("Types should be empty after Clear")
	}
	if len(reg.BeanIDs()) != 0 {
		t.Error("BeanIDs should be empty after Clear")
	}
}

func TestRegister_TypeNil_ReturnsError(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	err := reg.Register(BeanDef{Type: nil}, "nil")
	if err == nil {
		t.Error("expected error for nil type")
	}
}

func TestSameBeanDefinition_Idempotent(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))

	def := BeanDef{Type: typ, Name: "same", Scope: Singleton}
	if err := reg.Register(def, "same"); err != nil {
		t.Fatalf("first: %v", err)
	}
	// Same definition → idempotent, no error
	if err := reg.Register(def, "same"); err != nil {
		t.Errorf("idempotent register: %v", err)
	}
}

func TestRegister_DifferentScope_DifferentDefinition(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.Register(BeanDef{Type: typ, Name: "x", Scope: Singleton}, "x")
	err := reg.Register(BeanDef{Type: typ, Name: "x", Scope: Prototype}, "x")
	if !errors.Is(err, ErrBeanAlreadyExists) {
		t.Errorf("expected ErrBeanAlreadyExists, got %v", err)
	}
}

func TestCountByType_MultipleBeans(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()
	typ := reflect.TypeOf((*TestBean)(nil))
	strType := reflect.TypeOf((*string)(nil))

	_ = reg.Register(BeanDef{Type: typ, Name: "b1", Scope: Singleton}, "b1")
	_ = reg.Register(BeanDef{Type: typ, Name: "b2", Scope: Singleton}, "b2")
	_ = reg.Register(BeanDef{Type: strType, Name: "s1", Scope: Singleton}, "s1")

	if reg.CountByType(typ) != 2 {
		t.Errorf("CountByType(TestBean) = %d, want 2", reg.CountByType(typ))
	}
	if reg.CountByType(strType) != 1 {
		t.Errorf("CountByType(string) = %d, want 1", reg.CountByType(strType))
	}
}

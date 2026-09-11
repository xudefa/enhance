package registry

import (
	"reflect"
	"testing"
)

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

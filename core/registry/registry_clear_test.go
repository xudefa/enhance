package registry

import (
	"reflect"
	"testing"
)

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

func TestClearAfterRegister(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	_ = reg.Register(BeanDef{Type: typ, Name: "bean1", Scope: Singleton}, "bean1")
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

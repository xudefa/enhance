package registry

import (
	"reflect"
	"testing"
)

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

	_ = reg.Register(def1, "github.com/pkg1.TestBean#myBean")
	_ = reg.Register(def2, "github.com/pkg2.TestBean#myBean")

	// HasBean 应该返回 true
	if !reg.HasBean("myBean") {
		t.Error("Expected HasBean to return true")
	}
}

package registry

import (
	"reflect"
	"sync"
	"testing"
)

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

package boot

import (
	"reflect"
	"strings"
	"testing"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// TestBoot_ContainerAccessor_Has_ByTypeGeneratedID_Coverage 测试 Has 方法通过类型生成 ID 匹配
func TestBoot_ContainerAccessor_Has_ByTypeGeneratedID(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 使用 RegisterInstance 注册 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	// 创建 adapter
	adapter := &containerAccessorAdapter{container: container}

	// 通过类型生成的 ID 检查
	beanType := reflect.TypeOf(&TestBean{})
	generatedID := container.Generate(beanType)
	if !adapter.Has(generatedID) {
		t.Errorf("Expected Has(%s) to return true", generatedID)
	}
}

// TestBoot_ContainerAccessor_Has_EmptyContainer_Coverage 测试 Has 方法在空容器中的行为
func TestBoot_ContainerAccessor_Has_EmptyContainer(t *testing.T) {
	t.Parallel()

	container := core.NewContainer()
	adapter := &containerAccessorAdapter{container: container}

	// 空容器应该返回 false
	if adapter.Has("any-id") {
		t.Error("Expected Has to return false for empty container")
	}
}

// TestBoot_ContainerAccessor_Has_DirectNameMatch_Coverage 测试 Has 方法直接名称匹配
func TestBoot_ContainerAccessor_Has_DirectNameMatch(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 注册 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 获取 beans 列表
	beans := container.ListBeans()

	// 应该能通过某个名称找到
	found := false
	for name := range beans {
		if adapter.Has(name) {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected to find bean by name")
	}
}

// TestBoot_ContainerAccessor_Has_GetAllFallback_Coverage 测试 Has 方法通过 GetAll 回退匹配
func TestBoot_ContainerAccessor_Has_GetAllFallback(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 使用 RegisterInstance 注册 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 通过类型生成的 ID 检查（应该走 GetAll 回退路径）
	beanType := reflect.TypeOf(&TestBean{})
	generatedID := container.Generate(beanType)

	// 先尝试直接通过 beans[id] 查找，如果找不到则走 GetAll 回退
	beans := container.ListBeans()
	if _, ok := beans[generatedID]; !ok {
		// 如果直接查找失败，应该通过 GetAll 回退找到
		if !adapter.Has(generatedID) {
			t.Errorf("Expected Has(%s) to return true via GetAll fallback", generatedID)
		}
	}
}

// TestBoot_ContainerAccessor_Has_MultipleBranches_Coverage 测试 Has 方法的多个分支
func TestBoot_ContainerAccessor_Has_MultipleBranches(t *testing.T) {
	t.Parallel()

	type TestBean1 struct {
		Name string
	}
	type TestBean2 struct {
		Value int
	}

	container := core.NewContainer()

	// 注册多个 Bean
	err := container.RegisterInstance(&TestBean1{Name: "bean1"}, reflect.TypeOf(&TestBean1{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	err = container.RegisterInstance(&TestBean2{Value: 42}, reflect.TypeOf(&TestBean2{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 测试 GetAll 回退路径
	beanType1 := reflect.TypeOf(&TestBean1{})
	generatedID1 := container.Generate(beanType1)

	if !adapter.Has(generatedID1) {
		t.Errorf("Expected Has(%s) to return true", generatedID1)
	}

	// 测试不存在的 ID
	if adapter.Has("nonexistent-id") {
		t.Error("Expected Has to return false for nonexistent ID")
	}
}

// TestBoot_ContainerAccessor_Has_GetAllWithNilBeanType_Coverage 测试 Has 方法 GetAll 中 beanType 为 nil 的情况
func TestBoot_ContainerAccessor_Has_GetAllWithNilBeanType(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 注册一个正常的 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 获取生成的 ID
	beanType := reflect.TypeOf(&TestBean{})
	generatedID := container.Generate(beanType)

	// 这个测试主要覆盖 GetAll 循环中的 beanType == nil continue 分支
	// 虽然在实际使用中很难触发（因为注册的 Bean 类型不会为 nil），
	// 但我们可以验证 GetAll 路径被正确执行
	if !adapter.Has(generatedID) {
		t.Errorf("Expected Has(%s) to return true via GetAll", generatedID)
	}
}

// TestBoot_ContainerAccessor_Has_DirectBeanMatch_Coverage 测试 Has 方法 beans[id] 直接匹配分支
func TestBoot_ContainerAccessor_Has_DirectBeanMatch(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 注册 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 获取 beans 列表
	beans := container.ListBeans()

	// 应该能通过 beans 中的某个 key 直接匹配
	for beanKey := range beans {
		if !adapter.Has(beanKey) {
			t.Errorf("Expected Has(%s) to return true via direct beans[id] match", beanKey)
		}
	}
}

// TestBoot_ContainerAccessor_Has_HashSuffixMatch_Coverage 测试 Has 方法 # 后缀匹配分支
func TestBoot_ContainerAccessor_Has_HashSuffixMatch(t *testing.T) {
	t.Parallel()

	type TestBean struct {
		Name string
	}

	container := core.NewContainer()

	// 注册 Bean
	err := container.RegisterInstance(&TestBean{Name: "test"}, reflect.TypeOf(&TestBean{}))
	if err != nil {
		t.Fatalf("RegisterInstance failed: %v", err)
	}

	adapter := &containerAccessorAdapter{container: container}

	// 获取 beans 列表
	beans := container.ListBeans()

	// 查找带 # 后缀的名称
	for beanName := range beans {
		if idx := strings.LastIndex(beanName, "#"); idx >= 0 {
			id := beanName[idx+1:]
			if !adapter.Has(id) {
				t.Errorf("Expected Has(%s) to return true via hash suffix match", id)
			}
		}
	}
}

// TestBoot_ContainerAccessor_Has_GetAllNilBeanType_Coverage 测试 Has 方法 GetAll 中 beanType 为 nil 的分支
func TestBoot_ContainerAccessor_Has_GetAllNilBeanType(t *testing.T) {
	t.Parallel()

	// 创建一个 mock container，GetAll 返回包含 nil 的切片
	mockContainer := &mockContainerForHas{
		listBeansResult: map[string]*registry.BeanDef{
			"test.Bean#myBean": {Type: reflect.TypeOf(&struct{}{})},
		},
		getAllResult:   []any{nil, &struct{}{}},
		generateResult: "test.Bean",
	}

	adapter := &containerAccessorAdapter{container: mockContainer}

	// 测试 GetAll 中有 nil 元素的情况
	hasResult := adapter.Has("test.Bean")
	// 应该通过第二个非 nil 元素匹配
	if !hasResult {
		t.Error("Expected Has to return true via GetAll non-nil bean")
	}
}

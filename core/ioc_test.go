// Package core 提供 IoC 容器端到端集成测试。
//
// 测试场景：
//   - Bean 注册与获取
//   - 依赖注入
//   - 单例作用域
//   - 多个 Bean
package core

import (
	"sync/atomic"
	"testing"
)

// UserRepository 模拟用户仓库。
type UserRepository struct {
	data map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		data: map[string]string{
			"1": "Alice",
			"2": "Bob",
		},
	}
}

func (r *UserRepository) Find(id string) (string, bool) {
	name, ok := r.data[id]
	return name, ok
}

// UserService 模拟用户服务，依赖 UserRepository。
type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id string) (string, bool) {
	return s.repo.Find(id)
}

// TestIoC_BeanRegistration 测试 Bean 注册与获取。
func TestIoC_BeanRegistration(t *testing.T) {
	t.Parallel()

	c := NewContainer()

	repo := NewUserRepository()
	err := Register[*UserRepository](c,
		WithFactory[*UserRepository](func(_ ...any) (any, error) {
			return repo, nil
		}),
	)
	if err != nil {
		t.Fatalf("注册 UserRepository 失败: %v", err)
	}

	if !Has[*UserRepository](c, "") {
		t.Fatal("UserRepository 应该存在")
	}

	got, err := GetByName[*UserRepository](c, "")
	if err != nil {
		t.Fatalf("获取 UserRepository 失败: %v", err)
	}

	if got != repo {
		t.Fatal("获取的实例应该与注册的实例相同")
	}
}

// TestIoC_DependencyInjection 测试依赖注入。
func TestIoC_DependencyInjection(t *testing.T) {
	t.Parallel()

	c := NewContainer()

	repo := NewUserRepository()
	err := Register[*UserRepository](c,
		WithFactory[*UserRepository](func(_ ...any) (any, error) {
			return repo, nil
		}),
	)
	if err != nil {
		t.Fatalf("注册 UserRepository 失败: %v", err)
	}

	svc := NewUserService(repo)
	err = Register[*UserService](c,
		WithFactory[*UserService](func(_ ...any) (any, error) {
			return svc, nil
		}),
	)
	if err != nil {
		t.Fatalf("注册 UserService 失败: %v", err)
	}

	got, err := GetByName[*UserService](c, "")
	if err != nil {
		t.Fatalf("获取 UserService 失败: %v", err)
	}

	name, ok := got.GetUser("1")
	if !ok {
		t.Fatal("应该能找到用户 1")
	}
	if name != "Alice" {
		t.Fatalf("用户 1 应该是 Alice，got: %s", name)
	}
}

// TestIoC_SingletonScope 测试单例作用域。
func TestIoC_SingletonScope(t *testing.T) {
	t.Parallel()

	c := NewContainer()

	callCount := atomic.Int32{}
	err := Register[*UserRepository](c,
		WithFactory[*UserRepository](func(_ ...any) (any, error) {
			callCount.Add(1)
			return NewUserRepository(), nil
		}),
	)
	if err != nil {
		t.Fatalf("注册失败: %v", err)
	}

	_, _ = GetByName[*UserRepository](c, "")
	_, _ = GetByName[*UserRepository](c, "")
	_, _ = GetByName[*UserRepository](c, "")

	if callCount.Load() != 1 {
		t.Fatalf("单例应该只创建一次，实际创建: %d 次", callCount.Load())
	}
}

// TestIoC_BeanNotFound 测试获取不存在的 Bean。
func TestIoC_BeanNotFound(t *testing.T) {
	t.Parallel()

	c := NewContainer()

	_, err := GetByName[*UserRepository](c, "")
	if err == nil {
		t.Fatal("获取不存在的 Bean 应该返回错误")
	}
}

// TestIoC_MultipleBeans 测试多个同名 Bean。
func TestIoC_MultipleBeans(t *testing.T) {
	t.Parallel()

	c := NewContainer()

	err := Register[*UserRepository](c,
		WithFactory[*UserRepository](func(_ ...any) (any, error) {
			return &UserRepository{data: map[string]string{"1": "Primary"}}, nil
		}),
		WithName[*UserRepository]("primary"),
	)
	if err != nil {
		t.Fatalf("注册 primary 失败: %v", err)
	}

	err = Register[*UserRepository](c,
		WithFactory[*UserRepository](func(_ ...any) (any, error) {
			return &UserRepository{data: map[string]string{"1": "Secondary"}}, nil
		}),
		WithName[*UserRepository]("secondary"),
	)
	if err != nil {
		t.Fatalf("注册 secondary 失败: %v", err)
	}

	primary, err := GetByName[*UserRepository](c, "primary")
	if err != nil {
		t.Fatalf("获取 primary 失败: %v", err)
	}

	secondary, err := GetByName[*UserRepository](c, "secondary")
	if err != nil {
		t.Fatalf("获取 secondary 失败: %v", err)
	}

	name1, _ := primary.Find("1")
	name2, _ := secondary.Find("1")

	if name1 != "Primary" {
		t.Fatalf("primary 用户 1 应该是 Primary，got: %s", name1)
	}
	if name2 != "Secondary" {
		t.Fatalf("secondary 用户 1 应该是 Secondary，got: %s", name2)
	}
}

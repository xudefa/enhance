package config

import (
	"testing"
)

// priorityLoader 测试用 Loader 实现
type priorityLoader struct {
	name     string
	priority int
}

func (l *priorityLoader) Load(opts ...LoaderOption) (Config, error) {
	return NewConfig(), nil
}

func (l *priorityLoader) Priority() int {
	return l.priority
}

func (l *priorityLoader) SupportsWatch() bool {
	return false
}

func TestLoaderChain_Add(t *testing.T) {
	t.Parallel()
	chain := &LoaderChain{}
	chain.Add(&priorityLoader{name: "a", priority: 10})
	chain.Add(&priorityLoader{name: "b", priority: 5})

	if chain.Len() != 2 {
		t.Errorf("expected 2 loaders, got %d", chain.Len())
	}
}

func TestLoaderChain_Len(t *testing.T) {
	t.Parallel()
	chain := &LoaderChain{}
	if chain.Len() != 0 {
		t.Errorf("expected 0, got %d", chain.Len())
	}

	chain.Add(&priorityLoader{priority: 1})
	if chain.Len() != 1 {
		t.Errorf("expected 1, got %d", chain.Len())
	}
}

func TestLoaderChain_Less(t *testing.T) {
	t.Parallel()
	chain := &LoaderChain{
		loaders: []Loader{
			&priorityLoader{priority: 20},
			&priorityLoader{priority: 10},
		},
	}

	if !chain.Less(1, 0) {
		t.Error("expected Less(1,0) = true (10 < 20)")
	}
	if chain.Less(0, 1) {
		t.Error("expected Less(0,1) = false (20 < 10 is false)")
	}
}

func TestLoaderChain_Less_NilLoaders(t *testing.T) {
	t.Parallel()
	chain := &LoaderChain{
		loaders: []Loader{
			nil,
			&priorityLoader{priority: 10},
		},
	}

	if chain.Less(0, 1) {
		t.Error("expected Less(0,1) = false when first loader is nil")
	}
	if chain.Less(1, 0) {
		t.Error("expected Less(1,0) = false when second loader is nil")
	}
}

func TestLoaderChain_Swap(t *testing.T) {
	t.Parallel()
	a := &priorityLoader{name: "a", priority: 10}
	b := &priorityLoader{name: "b", priority: 20}
	chain := &LoaderChain{loaders: []Loader{a, b}}

	chain.Swap(0, 1)

	if chain.loaders[0] != b {
		t.Error("expected index 0 to be b after swap")
	}
	if chain.loaders[1] != a {
		t.Error("expected index 1 to be a after swap")
	}
}

func TestLoaderChain_Sorted(t *testing.T) {
	t.Parallel()
	chain := &LoaderChain{}
	chain.Add(&priorityLoader{name: "c", priority: 30})
	chain.Add(&priorityLoader{name: "a", priority: 10})
	chain.Add(&priorityLoader{name: "b", priority: 20})

	sorted := chain.Sorted()

	if len(sorted) != 3 {
		t.Fatalf("expected 3, got %d", len(sorted))
	}

	expectedPriorities := []int{10, 20, 30}
	for i, l := range sorted {
		if l.Priority() != expectedPriorities[i] {
			t.Errorf("index %d: expected priority %d, got %d", i, expectedPriorities[i], l.Priority())
		}
	}

	// 原始链不受影响
	if chain.loaders[0].Priority() != 30 {
		t.Error("original chain should not be modified by Sorted()")
	}
}

func TestWithPaths(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithPaths("/etc/app", "./config")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Paths) != 2 || model.Paths[0] != "/etc/app" || model.Paths[1] != "./config" {
		t.Errorf("expected paths ['/etc/app', './config'], got %v", model.Paths)
	}
}

func TestWithFileName(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithFileName("config.json")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.FileName != "config.json" {
		t.Errorf("expected FileName 'config.json', got %s", model.FileName)
	}
}

func TestWithLoaderFileType(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithLoaderFileType("yaml")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.FileType != "yaml" {
		t.Errorf("expected FileType 'yaml', got %s", model.FileType)
	}
}

func TestWithLoaderEnv(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithLoaderEnv("production")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Env != "production" {
		t.Errorf("expected Env 'production', got %s", model.Env)
	}
}

func TestWithPrefix(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithPrefix("APP_")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Prefix != "APP_" {
		t.Errorf("expected Prefix 'APP_', got %s", model.Prefix)
	}
}

func TestWithRemoteType(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithRemoteType("etcd")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.RemoteType != "etcd" {
		t.Errorf("expected RemoteType 'etcd', got %s", model.RemoteType)
	}
}

func TestWithEndpoints(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithEndpoints([]string{"127.0.0.1:2379", "127.0.0.2:2379"})
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(model.Endpoints) != 2 {
		t.Errorf("expected 2 endpoints, got %d", len(model.Endpoints))
	}
}

func TestWithLoaderKey(t *testing.T) {
	t.Parallel()
	model := &LoaderModel{}
	opt := WithLoaderKey("app.config")
	err := opt(model)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if model.Key != "app.config" {
		t.Errorf("expected Key 'app.config', got %s", model.Key)
	}
}

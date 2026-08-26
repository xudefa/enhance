package aop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewAopBeanScanner_NilContainer(t *testing.T) {
	t.Parallel()

	scanner := NewAopBeanScanner(nil)
	if scanner == nil {
		t.Fatal("expected non-nil scanner")
	}
	if scanner.container == nil {
		t.Error("expected container to be auto-created")
	}
}

func TestAopBeanScanner_EnableDisable(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	if !scanner.IsEnabled() {
		t.Error("expected scanner to be enabled by default")
	}

	scanner.Disable()
	if scanner.IsEnabled() {
		t.Error("expected scanner to be disabled after Disable()")
	}

	scanner.Enable()
	if !scanner.IsEnabled() {
		t.Error("expected scanner to be enabled after Enable()")
	}
}

func TestAopBeanScanner_GetContainer(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	if scanner.GetContainer() != container {
		t.Error("expected GetContainer to return original container")
	}
}

func TestAopBeanScanner_Scan_NonExistentPath(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	err := scanner.Scan("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Fatal("expected error for non-existent path")
	}
}

func TestAopBeanScanner_Scan_Disabled(t *testing.T) {
	t.Parallel()

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)
	scanner.Disable()

	// 扫描非存在路径，但因为禁用应该返回 nil
	err := scanner.Scan("/nonexistent/path")
	if err != nil {
		t.Fatalf("expected no error when scanner is disabled, got: %v", err)
	}
}

func TestAopBeanScanner_Scan_EmptyDir(t *testing.T) {
	t.Parallel()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "aop-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	err = scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error scanning empty dir: %v", err)
	}
}

func TestAopBeanScanner_Scan_WithGoFile(t *testing.T) {
	t.Parallel()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "aop-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建带有 @AopProxy 注解的 Go 文件
	goFile := filepath.Join(tmpDir, "test_service.go")
	content := `package test

// @AopProxy
type TestService struct {
}

func (s *TestService) DoWork() {}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	err = scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error scanning dir with go file: %v", err)
	}
}

func TestAopBeanScanner_Scan_ExcludedFiles(t *testing.T) {
	t.Parallel()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "aop-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建测试文件（应该被排除）
	testFile := filepath.Join(tmpDir, "test_service_test.go")
	content := `package test

// @AopProxy
type TestService struct {}
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 创建 AOP 生成文件（应该被排除）
	aopFile := filepath.Join(tmpDir, "test_service_aop.go")
	if err := os.WriteFile(aopFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write aop file: %v", err)
	}

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	err = scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error scanning dir with excluded files: %v", err)
	}
}

func TestAopBeanScanner_Scan_InvalidGoFile(t *testing.T) {
	t.Parallel()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "aop-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 创建无效的 Go 文件（应该被跳过）
	invalidFile := filepath.Join(tmpDir, "invalid.go")
	content := `this is not valid go syntax {{{`
	if err := os.WriteFile(invalidFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write invalid file: %v", err)
	}

	container := NewAopContainer(nil)
	scanner := NewAopBeanScanner(container)

	err = scanner.Scan(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error scanning dir with invalid go file: %v", err)
	}
}

func TestGlobalAopBeanScanner(t *testing.T) {
	t.Parallel()

	if GlobalAopBeanScanner == nil {
		t.Fatal("expected non-nil GlobalAopBeanScanner")
	}
}

func TestScanAopBeans(t *testing.T) {
	t.Parallel()

	// 创建临时目录
	tmpDir, err := os.MkdirTemp("", "aop-scanner-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	err = ScanAopBeans(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error from ScanAopBeans: %v", err)
	}
}

func TestAutoScan(t *testing.T) {
	t.Parallel()

	// AutoScan 扫描当前目录，应该能正常工作
	err := AutoScan()
	if err != nil {
		t.Fatalf("unexpected error from AutoScan: %v", err)
	}
}

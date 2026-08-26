package generator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	t.Parallel()
	p := NewParser()
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
	if p.fset == nil {
		t.Error("fset should not be nil")
	}
	if len(p.aspects) != 0 {
		t.Errorf("aspects should be empty, got %d", len(p.aspects))
	}
	if len(p.proxies) != 0 {
		t.Errorf("proxies should be empty, got %d", len(p.proxies))
	}
}

func TestParser_ParseDir_NonexistentDir(t *testing.T) {
	t.Parallel()
	p := NewParser()
	err := p.ParseDir("/nonexistent/path/that/does/not/exist")
	if err == nil {
		t.Error("expected error for nonexistent directory")
	}
}

func TestParser_ParseDir_SkipsTestFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	src := `package testpkg

// @AopProxy
type MyService struct{}

func (s *MyService) Do() {}
`
	testSrc := `package testpkg

// @AopProxy
type TestHelper struct{}

func (s *TestHelper) Do() {}
`
	_ = os.WriteFile(filepath.Join(dir, "service.go"), []byte(src), 0644)
	_ = os.WriteFile(filepath.Join(dir, "service_test.go"), []byte(testSrc), 0644)

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	if len(p.proxies) != 1 {
		t.Errorf("expected 1 proxy (test files skipped), got %d", len(p.proxies))
	}
}

func TestParser_ParseDir_SkipsAopFiles(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()

	normalSrc := `package testpkg

// @AopProxy
type RealService struct{}

func (s *RealService) Do() {}
`
	aopSrc := `package testpkg

// @AopProxy
type GeneratedProxy struct{}

func (s *GeneratedProxy) Do() {}
`
	_ = os.WriteFile(filepath.Join(dir, "service.go"), []byte(normalSrc), 0644)
	_ = os.WriteFile(filepath.Join(dir, "service_aop.go"), []byte(aopSrc), 0644)

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	if len(p.proxies) != 1 {
		t.Errorf("expected 1 proxy (_aop.go skipped), got %d", len(p.proxies))
	}
}

func TestParser_ParseDir_EmptyDir(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}
	if len(p.aspects) != 0 || len(p.proxies) != 0 {
		t.Error("expected empty results for empty directory")
	}
}

func TestParser_GetAspects_AfterParse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `package testpkg

// @Aspect(order=3)
type TracingAspect struct{}
`
	_ = os.WriteFile(filepath.Join(dir, "aspect.go"), []byte(src), 0644)

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	aspects := p.GetAspects()
	if len(aspects) != 1 {
		t.Fatalf("expected 1 aspect, got %d", len(aspects))
	}
	if aspects[0].Name != "TracingAspect" {
		t.Errorf("Name = %q, want %q", aspects[0].Name, "TracingAspect")
	}
	if aspects[0].Order != 3 {
		t.Errorf("Order = %d, want 3", aspects[0].Order)
	}
}

func TestParser_GetInterfaces_AfterParse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `package testpkg

// @ProxyInterface
type Greeter interface {
	Hello(name string) string
	Goodbye()
}
`
	_ = os.WriteFile(filepath.Join(dir, "iface.go"), []byte(src), 0644)

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	intfs := p.GetInterfaces()
	if len(intfs) != 1 {
		t.Fatalf("expected 1 interface, got %d", len(intfs))
	}
	if intfs[0].Name != "Greeter" {
		t.Errorf("Name = %q, want %q", intfs[0].Name, "Greeter")
	}
	if len(intfs[0].Methods) != 2 {
		t.Errorf("expected 2 methods, got %d", len(intfs[0].Methods))
	}
}

func TestParser_GetFuncs_AfterParse(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `package testpkg

// @Before(Service.Save)
func LogBefore(jp any) {}
`
	_ = os.WriteFile(filepath.Join(dir, "func.go"), []byte(src), 0644)

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir: %v", err)
	}

	funcs := p.GetFuncs()
	if len(funcs) != 1 {
		t.Fatalf("expected 1 func, got %d", len(funcs))
	}
	if funcs[0].FuncName != "LogBefore" {
		t.Errorf("FuncName = %q, want %q", funcs[0].FuncName, "LogBefore")
	}
	if !funcs[0].IsFunc {
		t.Error("expected IsFunc=true")
	}
}

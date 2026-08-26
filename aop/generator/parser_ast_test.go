package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestParser_ExtractAdviceTypeVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	tests := []struct {
		text string
		want AdviceType
	}{
		{"@Before(Service.Method)", AdviceBefore},
		{"@After(Service.Method)", AdviceAfter},
		{"@Around(Service.Method)", AdviceAround},
		{"@AfterReturning(Service.Method)", AdviceAfterReturning},
		{"@AfterThrowing(Service.Method)", AdviceAfterThrowing},
		{"@before(Svc.Foo)", AdviceBefore},
		{"no annotation here", ""},
		{"", ""},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.text, func(t *testing.T) {
			t.Parallel()
			got := p.extractAdviceType(tt.text)
			if got != tt.want {
				t.Errorf("extractAdviceType(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestParser_ExtractTargetsVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"single target", `@Before(Service.Method)`, []string{"Service.Method"}},
		{"multiple targets", `@Before(Service.Method, Other.Method)`, []string{"Service.Method", "Other.Method"}},
		{"quoted targets", `@Before("Service.Method")`, []string{"Service.Method"}},
		{"no parens", `@Before`, nil},
		{"empty parens", `@Before()`, nil},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := p.extractTargets(tt.text)
			if len(got) != len(tt.want) {
				t.Fatalf("extractTargets(%q) = %v, want %v", tt.text, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParser_ExprToStringVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	tests := []struct {
		name string
		src  string
		want string
	}{
		{"ident", `package p; type T struct { X int }`, "int"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			for _, decl := range f.Decls {
				gd, ok := decl.(*ast.GenDecl)
				if !ok {
					continue
				}
				for _, spec := range gd.Specs {
					ts, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					st, ok := ts.Type.(*ast.StructType)
					if !ok {
						continue
					}
					for _, field := range st.Fields.List {
						got := p.exprToString(field.Type)
						if got != tt.want {
							t.Errorf("exprToString() = %q, want %q", got, tt.want)
						}
					}
				}
			}
		})
	}
}

func TestParser_ResolveRecvTypeVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	tests := []struct {
		name     string
		src      string
		wantType string
	}{
		{"pointer receiver", `package p; type T struct{}; func (t *T) Do() {}`, "T"},
		{"value receiver", `package p; type T struct{}; func (t T) Do() {}`, "T"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, "test.go", tt.src, 0)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}
			for _, decl := range f.Decls {
				fd, ok := decl.(*ast.FuncDecl)
				if !ok || fd.Recv == nil {
					continue
				}
				got := p.resolveRecvType(fd.Recv)
				if got != tt.wantType {
					t.Errorf("resolveRecvType() = %q, want %q", got, tt.wantType)
				}
			}
		})
	}
}

func TestParser_ParseAspectAnnotationVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	p.parseAspectAnnotation(`@Aspect(order=5)`, "MyAspect", "mypkg")
	a, ok := p.aspects["MyAspect"]
	if !ok {
		t.Fatal("expected aspect to be registered")
	}
	if a.Order != 5 {
		t.Errorf("Order = %d, want 5", a.Order)
	}
	if a.Package != "mypkg" {
		t.Errorf("Package = %q, want %q", a.Package, "mypkg")
	}
}

func TestParser_ParseProxyAnnotationVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	p.parseProxyAnnotation(`@AopProxy`, "MyProxy", "pkg", "pkg/file.go")
	proxy, ok := p.proxies["MyProxy"]
	if !ok {
		t.Fatal("expected proxy to be registered")
	}
	if proxy.BeanID != "myProxy" {
		t.Errorf("BeanID = %q, want %q", proxy.BeanID, "myProxy")
	}
	if proxy.FilePath != "pkg/file.go" {
		t.Errorf("FilePath = %q", proxy.FilePath)
	}
}

func TestParser_ParseProxyAnnotation_CustomBeanIDVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	p.parseProxyAnnotation(`@AopProxy(beanId="customID")`, "MyProxy", "pkg", "pkg/file.go")
	proxy, ok := p.proxies["MyProxy"]
	if !ok {
		t.Fatal("expected proxy")
	}
	if proxy.BeanID != "customID" {
		t.Errorf("BeanID = %q, want %q", proxy.BeanID, "customID")
	}
}

func TestParser_ParseInterfaceAnnotationVariants(t *testing.T) {
	t.Parallel()
	p := NewParser()
	p.parseInterfaceAnnotation(`@ProxyInterface`, "MyService", "pkg", "pkg/service.go")
	intf, ok := p.interfaces["MyService"]
	if !ok {
		t.Fatal("expected interface to be registered")
	}
	if intf.BeanID != "myService" {
		t.Errorf("BeanID = %q, want %q", intf.BeanID, "myService")
	}
}

func TestParser_GetAspects_EmptyAdvanced(t *testing.T) {
	t.Parallel()
	p := NewParser()
	if got := p.GetAspects(); len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestParser_GetProxies_EmptyAdvanced(t *testing.T) {
	t.Parallel()
	p := NewParser()
	if got := p.GetProxies(); len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestParser_GetFuncs_EmptyAdvanced(t *testing.T) {
	t.Parallel()
	p := NewParser()
	if got := p.GetFuncs(); len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestParser_GetInterfaces_EmptyAdvanced(t *testing.T) {
	t.Parallel()
	p := NewParser()
	if got := p.GetInterfaces(); len(got) != 0 {
		t.Errorf("expected empty, got %d", len(got))
	}
}

func TestParser_ParseDir_IntegrationAdvanced(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	src := `package testpkg

// @Aspect(order=1)
type LoggingAspect struct{}

// @AopProxy
type UserService struct{}

func (s *UserService) GetUser(id int64) string { return "" }
`
	if err := os.WriteFile(filepath.Join(dir, "service.go"), []byte(src), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}

	p := NewParser()
	if err := p.ParseDir(dir); err != nil {
		t.Fatalf("ParseDir error: %v", err)
	}

	if len(p.aspects) != 1 {
		t.Errorf("expected 1 aspect, got %d", len(p.aspects))
	}
	if len(p.proxies) != 1 {
		t.Errorf("expected 1 proxy, got %d", len(p.proxies))
	}

	proxies := p.GetProxies()
	if len(proxies) != 1 {
		t.Fatalf("expected 1 proxy, got %d", len(proxies))
	}
	if proxies[0].BeanID != "userService" {
		t.Errorf("BeanID = %q, want %q", proxies[0].BeanID, "userService")
	}
}

package core

import (
	"reflect"
	"testing"
)

func TestGenerate_WithPointerType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf((*TestService)(nil))

	beanID := container.Generate(typ)
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("expected %q, got %q", expected, beanID)
	}
}

func TestGenerate_WithNonPointerType(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf(TestService{})

	beanID := container.Generate(typ)
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("expected %q, got %q", expected, beanID)
	}
}

func TestGenerate_WithEmptyCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf(TestService{})

	beanID := container.Generate(typ, "")
	expected := "github.com/xudefa/enhance/core.TestService"
	if beanID != expected {
		t.Errorf("expected %q, got %q", expected, beanID)
	}
}

func TestGenerate_WithStandardFormatName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	typ := reflect.TypeOf(TestService{})
	standardName := "github.com/xudefa/enhance/core.TestService#custom"

	beanID := container.Generate(typ, standardName)
	if beanID != standardName {
		t.Errorf("expected %q, got %q", standardName, beanID)
	}
}

func TestParse_NoDot(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	pkg, typ, custom := container.Parse("SimpleType")
	if pkg != "" {
		t.Errorf("expected empty pkg, got %q", pkg)
	}
	if typ != "SimpleType" {
		t.Errorf("expected type 'SimpleType', got %q", typ)
	}
	if custom != "" {
		t.Errorf("expected empty custom, got %q", custom)
	}
}

func TestParse_WithCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	pkg, typ, custom := container.Parse("github.com/pkg.MyType#myBean")
	if pkg != "github.com/pkg" {
		t.Errorf("expected pkg 'github.com/pkg', got %q", pkg)
	}
	if typ != "MyType" {
		t.Errorf("expected type 'MyType', got %q", typ)
	}
	if custom != "myBean" {
		t.Errorf("expected custom 'myBean', got %q", custom)
	}
}

func TestParse_WithoutCustomName(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	pkg, typ, custom := container.Parse("github.com/pkg.MyType")
	if pkg != "github.com/pkg" {
		t.Errorf("expected pkg 'github.com/pkg', got %q", pkg)
	}
	if typ != "MyType" {
		t.Errorf("expected type 'MyType', got %q", typ)
	}
	if custom != "" {
		t.Errorf("expected empty custom, got %q", custom)
	}
}

func TestParse_MultipleHashSigns(t *testing.T) {
	t.Parallel()
	container := NewContainer()
	pkg, typ, custom := container.Parse("github.com/pkg.MyType#part1#part2")
	if pkg != "github.com/pkg" {
		t.Errorf("expected pkg 'github.com/pkg', got %q", pkg)
	}
	if typ != "MyType" {
		t.Errorf("expected type 'MyType', got %q", typ)
	}
	if custom != "part1#part2" {
		t.Errorf("expected custom 'part1#part2', got %q", custom)
	}
}

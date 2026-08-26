package binding

import (
	"reflect"
	"testing"
)

func TestInjectByTypeEmptyName(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"ptr-TestService": &TestService{Name: "by-type"},
		},
		types: map[reflect.Type][]string{
			reflect.TypeOf((*TestService)(nil)): {"ptr-TestService"},
		},
	}

	svc, err := Inject[*TestService](mock, "")
	if err != nil {
		t.Fatalf("Inject by type failed: %v", err)
	}
	if svc.Name != "by-type" {
		t.Errorf("expected 'by-type', got %q", svc.Name)
	}
}

func TestInjectByNameNotFound(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{beans: map[string]any{}}

	_, err := Inject[*TestService](mock, "nonexistent")
	if err == nil {
		t.Error("expected error for nonexistent bean")
	}
}

func TestInjectByTypeNoMatch(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{},
		types: map[reflect.Type][]string{},
	}

	_, err := Inject[*TestService](mock, "")
	if err == nil {
		t.Error("expected error when no bean of type exists")
	}
}

func TestInjectTypeMismatch(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{
		beans: map[string]any{
			"wrong": "not a TestService",
		},
	}

	_, err := Inject[*TestService](mock, "wrong")
	if err == nil {
		t.Error("expected error for type mismatch")
	}
}

func TestMustInjectByNameNotFound(t *testing.T) {
	t.Parallel()
	mock := &mockBeanGet{beans: map[string]any{}}

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for nonexistent bean")
		}
	}()
	MustInject[*TestService](mock, "nonexistent")
}

func TestInjectOption_WithRequired(t *testing.T) {
	t.Parallel()
	opt := WithRequired()
	cfg := &injectConfig{}
	opt(cfg)
	if !cfg.Required {
		t.Error("expected Required to be true")
	}
}

func TestInjectOption_WithOptional(t *testing.T) {
	t.Parallel()
	opt := WithOptional()
	cfg := &injectConfig{Required: true}
	opt(cfg)
	if cfg.Required {
		t.Error("expected Required to be false")
	}
}

func TestInjectOption_ChainOrder(t *testing.T) {
	t.Parallel()
	cfg := &injectConfig{}
	opts := []InjectOption{WithRequired(), WithOptional(), WithRequired()}
	for _, opt := range opts {
		opt(cfg)
	}
	if !cfg.Required {
		t.Error("expected Required to be true after last WithRequired")
	}
}

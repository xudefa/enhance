package registry

import (
	"errors"
	"reflect"
	"testing"
)

func TestRegisterInstance(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func(reg BeanRegistry)
		instance  any
		typ       reflect.Type
		wantErr   bool
		checkFunc func(t *testing.T, reg BeanRegistry)
	}{
		{
			name:     "success registers primary singleton",
			instance: &TestBean{Value: "inst"},
			typ:      reflect.TypeOf((*TestBean)(nil)),
			checkFunc: func(t *testing.T, reg BeanRegistry) {
				t.Helper()
				def, ok := reg.GetDefinition("id-inst")
				if !ok {
					t.Fatal("expected definition to exist")
				}
				if !def.Primary {
					t.Error("expected instance bean to be primary")
				}
				if def.Scope != Singleton {
					t.Errorf("expected singleton scope, got %v", def.Scope)
				}
				bean, err := def.Factory()
				if err != nil {
					t.Fatalf("factory failed: %v", err)
				}
				got, ok := bean.(*TestBean)
				if !ok || got.Value != "inst" {
					t.Errorf("unexpected factory result: %v", bean)
				}
				primaryID, ok := reg.GetPrimaryByType(reflect.TypeOf((*TestBean)(nil)))
				if !ok || primaryID != "id-inst" {
					t.Errorf("expected id-inst as primary, got %q", primaryID)
				}
			},
		},
		{
			name: "duplicate instance id returns error",
			setup: func(reg BeanRegistry) {
				if err := reg.RegisterInstance(&TestBean{Value: "first"}, reflect.TypeOf((*TestBean)(nil)), "id-inst"); err != nil {
					t.Errorf("setup register failed: %v", err)
				}
			},
			instance: &TestBean{Value: "second"},
			typ:      reflect.TypeOf((*TestBean)(nil)),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg := NewBeanRegistry()
			if tt.setup != nil {
				tt.setup(reg)
			}
			err := reg.RegisterInstance(tt.instance, tt.typ, "id-inst")
			if (err != nil) != tt.wantErr {
				t.Fatalf("RegisterInstance error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, reg)
			}
		})
	}
}

func TestListBeansAndInstances(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	_ = reg.Register(BeanDef{Type: typ, Name: "b1", Scope: Singleton}, "b1")
	_ = reg.Register(BeanDef{Type: typ, Name: "b2", Scope: Prototype}, "b2")
	reg.SetInstance("b1", &TestBean{Value: "i1"})
	reg.SetInstance("b2", &TestBean{Value: "i2"})

	beans := reg.ListBeans()
	if len(beans) != 2 {
		t.Fatalf("expected 2 beans, got %d", len(beans))
	}
	if beans["b1"].Name != "b1" || beans["b2"].Scope != Prototype {
		t.Errorf("unexpected bean definitions: %+v", beans)
	}

	instances := reg.ListInstances()
	if len(instances) != 2 {
		t.Fatalf("expected 2 instances, got %d", len(instances))
	}
	if v, ok := instances["b1"].(*TestBean); !ok || v.Value != "i1" {
		t.Errorf("unexpected instance for b1: %v", instances["b1"])
	}

	beans["b1"].Name = "mutated"
	delete(beans, "b2")

	if again := reg.ListBeans(); again["b1"].Name != "b1" || len(again) != 2 {
		t.Error("ListBeans must return copies, mutations must not leak")
	}

	instances["b1"] = nil
	delete(instances, "b2")

	if again := reg.ListInstances(); len(again) != 2 {
		t.Error("ListInstances must return a snapshot copy")
	}
}

func TestRegisterDifferentDefinitionReturnsErrAlreadyExists(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))
	first := BeanDef{Type: typ, Name: "svc", Scope: Singleton}
	second := BeanDef{Type: typ, Name: "svc", Scope: Prototype}

	if err := reg.Register(first, "svc"); err != nil {
		t.Fatalf("first register failed: %v", err)
	}

	err := reg.Register(second, "svc")
	if !errors.Is(err, ErrBeanAlreadyExists) {
		t.Fatalf("expected ErrBeanAlreadyExists, got %v", err)
	}
}

func TestRegisterCustomNameConflict(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		firstID    string
		firstName  string
		secondID   string
		secondName string
		wantErr    bool
	}{
		{"plain name conflict", "id-1", "alpha", "id-2", "alpha", true},
		{"suffix conflict via #name", "id-1", "pkg1.Svc#shared", "id-2", "pkg2.Svc#shared", true},
		{"distinct names ok", "id-1", "pkg1.Svc#a", "id-2", "pkg2.Svc#b", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			reg := NewBeanRegistry()
			typ := reflect.TypeOf((*TestBean)(nil))

			if err := reg.Register(BeanDef{Type: typ, Name: tt.firstName}, tt.firstID); err != nil {
				t.Fatalf("first register failed: %v", err)
			}
			err := reg.Register(BeanDef{Type: typ, Name: tt.secondName}, tt.secondID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFuncPtr(t *testing.T) {
	t.Parallel()

	var nilFn func(bean any) error
	fn := func(bean any) error { return nil }

	tests := []struct {
		name string
		in   any
		want uintptr
	}{
		{"nil interface", nil, 0},
		{"non-func kind", "not-a-func", 0},
		{"nil func value", nilFn, 0},
		{"valid func", fn, reflect.ValueOf(fn).Pointer()},
		{"same func twice equal", fn, reflect.ValueOf(fn).Pointer()},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := funcPtr(tt.in); got != tt.want {
				t.Errorf("funcPtr() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNormalizeScope(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   Scope
		want Scope
	}{
		{"empty maps to singleton", "", Singleton},
		{"singleton kept", Singleton, Singleton},
		{"prototype kept", Prototype, Prototype},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := normalizeScope(tt.in); got != tt.want {
				t.Errorf("normalizeScope(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSameBeanDefinitionIgnoresEmptyScopeAndFactory(t *testing.T) {
	t.Parallel()
	reg := NewBeanRegistry()

	typ := reflect.TypeOf((*TestBean)(nil))

	if err := reg.Register(BeanDef{Type: typ}, "svc"); err != nil {
		t.Fatalf("register failed: %v", err)
	}

	if err := reg.Register(BeanDef{Type: typ, Scope: Singleton}, "svc"); err != nil {
		t.Errorf("empty scope should equal singleton, got %v", err)
	}
}

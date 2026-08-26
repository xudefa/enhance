package generator

import "testing"

func TestAdviceAdapterTemplateData_Fields(t *testing.T) {
	t.Parallel()
	d := AdviceAdapterTemplateData{
		Package: "mypkg",
		Adapters: []AdviceAdapterData{
			{
				FuncName:   "adaptFunc",
				AspectType: "MyAspect",
				MethodName: "Do",
				IsAround:   true,
				HasReturn:  true,
			},
		},
	}
	if d.Package != "mypkg" {
		t.Errorf("Package = %q", d.Package)
	}
	if len(d.Adapters) != 1 {
		t.Fatalf("Adapters len = %d, want 1", len(d.Adapters))
	}
	a := d.Adapters[0]
	if a.FuncName != "adaptFunc" || a.AspectType != "MyAspect" || a.MethodName != "Do" || !a.IsAround || !a.HasReturn {
		t.Errorf("unexpected adapter: %+v", a)
	}
}

func TestAspectTemplateData_Fields(t *testing.T) {
	t.Parallel()
	a := AspectTemplateData{
		MethodName:       "Do",
		AdviceType:       "Before",
		AdviceFunc:       "logBefore",
		Order:            2,
		AspectName:       "LogAspect",
		AspectMethodName: "Log",
	}
	if a.MethodName != "Do" || a.AdviceType != "Before" || a.Order != 2 {
		t.Errorf("unexpected: %+v", a)
	}
}

func TestAdviceBindingData_Fields(t *testing.T) {
	t.Parallel()
	d := AdviceBindingData{
		AdviceFunc: "bindFunc",
		HasParams:  true,
	}
	if d.AdviceFunc != "bindFunc" || !d.HasParams {
		t.Errorf("unexpected: %+v", d)
	}
}

func TestTemplateConstantsExist(t *testing.T) {
	t.Parallel()
	if simpleProxyTemplate == "" {
		t.Error("simpleProxyTemplate should not be empty")
	}
	if staticAopProxyTemplate == "" {
		t.Error("staticAopProxyTemplate should not be empty")
	}
	if staticInterfaceProxyTemplate == "" {
		t.Error("staticInterfaceProxyTemplate should not be empty")
	}
	if aopProxyTemplate == "" {
		t.Error("aopProxyTemplate should not be empty")
	}
	if adviceAdapterTemplate == "" {
		t.Error("adviceAdapterTemplate should not be empty")
	}
}

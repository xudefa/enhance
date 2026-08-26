package generator

import "testing"

func TestAdviceTypeConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  AdviceType
		want string
	}{
		{"AdviceBefore", AdviceBefore, "before"},
		{"AdviceAfter", AdviceAfter, "after"},
		{"AdviceAround", AdviceAround, "around"},
		{"AdviceAfterReturning", AdviceAfterReturning, "after_returning"},
		{"AdviceAfterThrowing", AdviceAfterThrowing, "after_throwing"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if string(tt.got) != tt.want {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestProxyInfo_Fields(t *testing.T) {
	t.Parallel()
	pi := ProxyInfo{
		Name:     "UserServiceProxy",
		Package:  "service",
		FilePath: "service/user.go",
		Target:   "UserService",
		Methods: []MethodInfo{
			{Name: "GetUser", Receiver: "UserService", Exported: true},
		},
		Aspects: []AspectInfo{
			{Name: "LoggingAspect", Order: 1},
		},
		BeanID: "userServiceProxy",
	}
	if pi.Name != "UserServiceProxy" {
		t.Errorf("Name = %q, want %q", pi.Name, "UserServiceProxy")
	}
	if len(pi.Methods) != 1 {
		t.Errorf("Methods len = %d, want 1", len(pi.Methods))
	}
	if len(pi.Aspects) != 1 {
		t.Errorf("Aspects len = %d, want 1", len(pi.Aspects))
	}
}

func TestInterfaceInfo_Fields(t *testing.T) {
	t.Parallel()
	ii := InterfaceInfo{
		Name:     "ServiceInterface",
		Package:  "service",
		FilePath: "service/service.go",
		Methods: []MethodInfo{
			{Name: "Execute", Receiver: "ServiceInterface", Exported: true},
		},
		BeanID: "serviceInterface",
		Aspects: []AspectInfo{
			{Name: "TracingAspect", Order: 0},
		},
	}
	if ii.Name != "ServiceInterface" {
		t.Errorf("Name = %q, want %q", ii.Name, "ServiceInterface")
	}
	if ii.BeanID != "serviceInterface" {
		t.Errorf("BeanID = %q, want %q", ii.BeanID, "serviceInterface")
	}
	if len(ii.Methods) != 1 || ii.Methods[0].Name != "Execute" {
		t.Errorf("Methods = %+v", ii.Methods)
	}
}

func TestMethodInfo_Fields(t *testing.T) {
	t.Parallel()
	m := MethodInfo{
		Name:     "Save",
		Receiver: "Repository",
		Params:   []ParamInfo{{Name: "data", Type: "string"}},
		Results:  []ParamInfo{{Name: "", Type: "error"}},
		Exported: true,
	}
	if m.Name != "Save" {
		t.Errorf("Name = %q", m.Name)
	}
	if !m.Exported {
		t.Error("expected Exported=true")
	}
	if len(m.Params) != 1 || m.Params[0].Type != "string" {
		t.Errorf("Params = %+v", m.Params)
	}
	if len(m.Results) != 1 || m.Results[0].Type != "error" {
		t.Errorf("Results = %+v", m.Results)
	}
}

func TestParamInfo_Fields(t *testing.T) {
	t.Parallel()
	p := ParamInfo{Name: "id", Type: "int64"}
	if p.Name != "id" || p.Type != "int64" {
		t.Errorf("ParamInfo = %+v", p)
	}
}

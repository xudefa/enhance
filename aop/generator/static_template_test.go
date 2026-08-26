package generator

import (
	"strings"
	"testing"
)

func TestRenderStaticProxyWithAspects(t *testing.T) {
	t.Parallel()
	data := &StaticProxyTemplateData{
		Package:    "mypkg",
		ProxyName:  "SvcProxy",
		TargetName: "Svc",
		SourceFile: "svc.go",
		Methods: []StaticMethodInfo{
			{
				Name:           "Run",
				ParamsStr:      "",
				ResultsStr:     "",
				ArgsList:       "",
				HasReturnValue: false,
				HasAspects:     true,
				BeforeAdvices: []AdviceRef{
					{AdviceField: "logBefore"},
				},
			},
		},
	}
	got, err := RenderStaticProxy(data)
	if err != nil {
		t.Fatalf("RenderStaticProxy: %v", err)
	}
	if !strings.Contains(got, "Before") {
		t.Error("expected Before advice comment in output")
	}
}

func TestStaticProxyTemplateData_FieldsAdvanced(t *testing.T) {
	t.Parallel()
	d := StaticProxyTemplateData{
		Package:    "pkg",
		ProxyName:  "Proxy",
		TargetName: "Target",
		SourceFile: "target.go",
	}
	if d.Package != "pkg" || d.ProxyName != "Proxy" || d.TargetName != "Target" || d.SourceFile != "target.go" {
		t.Errorf("unexpected field values: %+v", d)
	}
}

func TestInterfaceProxyTemplateData_FieldsAdvanced(t *testing.T) {
	t.Parallel()
	d := InterfaceProxyTemplateData{
		Package:       "pkg",
		ProxyName:     "Proxy",
		InterfaceName: "Interface",
		SourceFile:    "iface.go",
	}
	if d.InterfaceName != "Interface" {
		t.Errorf("InterfaceName = %q", d.InterfaceName)
	}
}

func TestAdviceField_FieldsAdvanced(t *testing.T) {
	t.Parallel()
	af := AdviceField{Name: "logBefore", Type: "func(aop.JoinPoint)"}
	if af.Name != "logBefore" || af.Type != "func(aop.JoinPoint)" {
		t.Errorf("AdviceField = %+v", af)
	}
}

func TestAdviceRef_FieldsAdvanced(t *testing.T) {
	t.Parallel()
	ar := AdviceRef{AdviceField: "myAdvice"}
	if ar.AdviceField != "myAdvice" {
		t.Errorf("AdviceRef = %+v", ar)
	}
}

func TestStaticMethodInfo_FieldsAdvanced(t *testing.T) {
	t.Parallel()
	m := StaticMethodInfo{
		Name:           "Do",
		HasAspects:     true,
		HasReturnValue: true,
		HasParams:      true,
	}
	if !m.HasAspects || !m.HasReturnValue || !m.HasParams {
		t.Errorf("unexpected fields: %+v", m)
	}
}

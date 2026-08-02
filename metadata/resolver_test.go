package metadata

import (
	"reflect"
	"testing"
)

// TestUser 测试用结构体
type TestUser struct {
	Name    string `metadata:"required:field=name,minLength=1,maxLength=100"`
	Email   string `metadata:"required:field=email,pattern=^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$"`
	Age     int    `metadata:"optional:field=age,min=0,max=150"`
	Address string `metadata:"optional:field=address"`
}

func TestTagAnnotationResolver_ResolveAnnotations(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	if len(annotations) != 4 {
		t.Fatalf("expected 4 annotations, got %d", len(annotations))
	}

	// 验证第一个注解
	if annotations[0].Name != "required" {
		t.Errorf("expected annotation name 'required', got '%s'", annotations[0].Name)
	}
	if field, ok := GetStringAttribute(annotations[0], "field"); !ok || field != "name" {
		t.Errorf("expected field='name', got '%v'", field)
	}
	if minLen, ok := GetIntAttribute(annotations[0], "minLength"); !ok || minLen != 1 {
		t.Errorf("expected minLength=1, got %d", minLen)
	}
	if maxLen, ok := GetIntAttribute(annotations[0], "maxLength"); !ok || maxLen != 100 {
		t.Errorf("expected maxLength=100, got %d", maxLen)
	}
}

func TestTagAnnotationResolver_GetAnnotation(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	// 查找 required 注解
	var found bool
	for _, ann := range annotations {
		if ann.Name == "required" {
			if field, ok := GetStringAttribute(ann, "field"); !ok || field != "name" {
				t.Errorf("expected field='name', got '%v'", field)
			}
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find 'required' annotation")
	}
}

func TestTagAnnotationResolver_HasAnnotation(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	// 检查是否有 required 注解
	hasRequired := false
	hasNonexistent := false
	for _, ann := range annotations {
		if ann.Name == "required" {
			hasRequired = true
		}
		if ann.Name == "nonexistent" {
			hasNonexistent = true
		}
	}

	if !hasRequired {
		t.Error("expected to have 'required' annotation")
	}

	if hasNonexistent {
		t.Error("expected not to have 'nonexistent' annotation")
	}
}

func TestTagAnnotationResolver_GetFieldAnnotations(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	// 查找 Name 字段的注解
	var nameAnnotations []Annotation
	for _, ann := range annotations {
		if field, ok := GetStringAttribute(ann, "field"); ok && field == "name" {
			nameAnnotations = append(nameAnnotations, ann)
		}
	}

	if len(nameAnnotations) != 1 {
		t.Fatalf("expected 1 annotation for Name field, got %d", len(nameAnnotations))
	}

	if nameAnnotations[0].Name != "required" {
		t.Errorf("expected annotation name 'required', got '%s'", nameAnnotations[0].Name)
	}
}

func TestTagAnnotationResolver_GetFieldAnnotation(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	// 查找 Name 字段的 required 注解
	var found bool
	for _, ann := range annotations {
		if field, ok := GetStringAttribute(ann, "field"); ok && field == "name" && ann.Name == "required" {
			if ann.Name != "required" {
				t.Errorf("expected annotation name 'required', got '%s'", ann.Name)
			}
			found = true
			break
		}
	}

	if !found {
		t.Error("expected to find 'required' annotation for Name field")
	}
}

func TestTagAnnotationResolver_EmptyTag(t *testing.T) {
	t.Parallel()

	type EmptyStruct struct {
		Field string `metadata:""`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(EmptyStruct{}))

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_NoTag(t *testing.T) {
	t.Parallel()

	type NoTagStruct struct {
		Field string
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(NoTagStruct{}))

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_NonStruct(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf("string"))

	if len(annotations) != 0 {
		t.Errorf("expected 0 annotations for non-struct type, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_PointerType(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(&TestUser{}))

	if len(annotations) != 4 {
		t.Fatalf("expected 4 annotations for pointer type, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_Cache(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")

	// 第一次解析
	annotations1 := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	// 第二次解析（应该使用缓存）
	annotations2 := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))

	if len(annotations1) != len(annotations2) {
		t.Errorf("cached annotations length mismatch")
	}
}

func TestTagAnnotationResolver_CacheIsolation(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("metadata")

	first := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))
	for i := range first {
		first[i].Name = "corrupted"
		first[i].Attributes["corrupted"] = true
	}

	second := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))
	if len(second) == 0 {
		t.Fatal("expected annotations from cache")
	}
	for _, ann := range second {
		if ann.Name == "corrupted" {
			t.Error("expected cache not to be corrupted by caller mutation")
		}
		if _, ok := ann.Attributes["corrupted"]; ok {
			t.Error("expected cache attributes not to be corrupted by caller mutation")
		}
	}
}

func TestTagAnnotationResolver_DefaultTagName(t *testing.T) {
	t.Parallel()

	resolver := NewTagAnnotationResolver("")
	// 验证默认 tag name 是 metadata
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(TestUser{}))
	if len(annotations) != 4 {
		t.Errorf("expected 4 annotations with default tag name, got %d", len(annotations))
	}
}

func TestTagAnnotationResolver_BoolAttribute(t *testing.T) {
	t.Parallel()

	type BoolStruct struct {
		Field string `metadata:"required:enabled=true,disabled=false"`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(BoolStruct{}))

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	if enabled, ok := GetBoolAttribute(ann, "enabled"); !ok || !enabled {
		t.Errorf("expected enabled=true, got %v", enabled)
	}

	if disabled, ok := GetBoolAttribute(ann, "disabled"); !ok || disabled {
		t.Errorf("expected disabled=false, got %v", disabled)
	}
}

func TestTagAnnotationResolver_FloatAttribute(t *testing.T) {
	t.Parallel()

	type FloatStruct struct {
		Field string `metadata:"config:threshold=0.75"`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(FloatStruct{}))

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	threshold, ok := ann.Attributes["threshold"]
	if !ok {
		t.Fatal("expected 'threshold' attribute")
	}

	if f, ok := threshold.(float64); !ok || f != 0.75 {
		t.Errorf("expected threshold=0.75, got %v", threshold)
	}
}

func TestTagAnnotationResolver_GetIntAttribute_FromFloat(t *testing.T) {
	t.Parallel()

	type FloatStruct struct {
		Field string `metadata:"config:count=42"`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(FloatStruct{}))

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	count, ok := GetIntAttribute(ann, "count")
	if !ok {
		t.Fatal("expected 'count' attribute")
	}

	if count != 42 {
		t.Errorf("expected count=42, got %d", count)
	}
}

func TestTagAnnotationResolver_GetStringAttribute_Conversion(t *testing.T) {
	t.Parallel()

	type MixedStruct struct {
		Field string `metadata:"config:value=123"`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(MixedStruct{}))

	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation, got %d", len(annotations))
	}

	ann := annotations[0]
	// GetIntAttribute 应该能正确解析整数
	value, ok := GetIntAttribute(ann, "value")
	if !ok {
		t.Fatal("expected 'value' attribute")
	}

	if value != 123 {
		t.Errorf("expected value=123, got %d", value)
	}
}

func TestTagAnnotationResolver_UnexportedField(t *testing.T) {
	t.Parallel()

	type UnexportedStruct struct {
		_          string `metadata:"required:field=exported"` //nolint:unused // 测试用未导出字段
		Unexported string `metadata:"optional:field=unexported"`
	}

	resolver := NewTagAnnotationResolver("metadata")
	annotations := resolver.ResolveAnnotations(reflect.TypeOf(UnexportedStruct{}))

	// 只应该解析导出字段
	if len(annotations) != 1 {
		t.Fatalf("expected 1 annotation (only exported field), got %d", len(annotations))
	}

	if annotations[0].Name != "optional" {
		t.Errorf("expected annotation name 'optional', got '%s'", annotations[0].Name)
	}
}

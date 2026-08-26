package binding

import (
	"reflect"
	"testing"
	"time"
)

func TestValueResolverFunc_Resolve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		key       string
		wantValue string
		wantOK    bool
	}{
		{"existing key", "timeout", "30", true},
		{"missing key", "missing", "", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolver := ValueResolverFunc(func(key string) (string, bool) {
				if key == "timeout" {
					return "30", true
				}
				return "", false
			})
			got, ok := resolver.Resolve(tt.key)
			if got != tt.wantValue || ok != tt.wantOK {
				t.Errorf("Resolve(%q) = (%q, %v), want (%q, %v)", tt.key, got, ok, tt.wantValue, tt.wantOK)
			}
		})
	}
}

func TestSetFieldValue_ConverterReturnsNil(t *testing.T) {
	t.Parallel()
	type Target struct {
		Val int
	}
	conv := &testConverter{
		convertFunc: func(value string, targetType string) (any, error) {
			return nil, nil
		},
	}
	binder := &defaultBinder{converter: conv}
	target := &Target{}
	val := reflect.ValueOf(target).Elem().FieldByName("Val")

	err := binder.setFieldValue(val, "42", reflect.TypeOf(0))
	if err == nil {
		t.Error("expected error when converter returns nil")
	}
}

func TestSetFieldValue_ConverterReturnsIncompatibleType(t *testing.T) {
	t.Parallel()
	type Target struct {
		Val int
	}
	conv := &testConverter{
		convertFunc: func(value string, targetType string) (any, error) {
			return "string", nil // string is not assignable to int
		},
	}
	binder := &defaultBinder{converter: conv}
	target := &Target{}
	val := reflect.ValueOf(target).Elem().FieldByName("Val")

	err := binder.setFieldValue(val, "42", reflect.TypeOf(0))
	if err == nil {
		t.Error("expected error for incompatible converter return type")
	}
}

func TestSetFieldValue_IntOverflow(t *testing.T) {
	t.Parallel()
	type Target struct {
		Val int8
	}
	binder := &defaultBinder{}
	target := &Target{}
	val := reflect.ValueOf(target).Elem().FieldByName("Val")

	err := binder.setFieldValue(val, "200", reflect.TypeOf(int8(0)))
	if err == nil {
		t.Error("expected overflow error for int8")
	}
}

func TestSetFieldValue_UintOverflow(t *testing.T) {
	t.Parallel()
	type Target struct {
		Val uint8
	}
	binder := &defaultBinder{}
	target := &Target{}
	val := reflect.ValueOf(target).Elem().FieldByName("Val")

	err := binder.setFieldValue(val, "300", reflect.TypeOf(uint8(0)))
	if err == nil {
		t.Error("expected overflow error for uint8")
	}
}

func TestNewBinder(t *testing.T) {
	t.Parallel()
	binder := NewBinder()
	if binder == nil {
		t.Fatal("expected non-nil binder")
	}
}

func TestNewTypeConverter(t *testing.T) {
	t.Parallel()
	converter := NewTypeConverter()
	if converter == nil {
		t.Fatal("expected non-nil converter")
	}

	tests := []struct {
		name       string
		value      string
		targetType string
		wantErr    bool
	}{
		{"string", "hello", "string", false},
		{"int", "42", "int", false},
		{"int64", "100", "int64", false},
		{"float64", "3.14", "float64", false},
		{"bool", "true", "bool", false},
		{"duration", "5s", "time.Duration", false},
		{"unsupported", "val", "custom", true},
		{"invalid int", "abc", "int", true},
		{"invalid float", "abc", "float64", true},
		{"invalid bool", "abc", "bool", true},
		{"invalid duration", "abc", "time.Duration", true},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := converter.Convert(tt.value, tt.targetType)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert(%q, %q) error = %v, wantErr %v", tt.value, tt.targetType, err, tt.wantErr)
			}
		})
	}
}

func TestSetFieldValue_DurationTypes(t *testing.T) {
	t.Parallel()
	type Target struct {
		Timeout time.Duration
	}
	binder := &defaultBinder{}
	target := &Target{}
	val := reflect.ValueOf(target).Elem().FieldByName("Timeout")

	err := binder.setFieldValue(val, "30s", reflect.TypeOf(time.Duration(0)))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if target.Timeout != 30*time.Second {
		t.Errorf("expected 30s, got %v", target.Timeout)
	}
}

func TestSetFieldValue_AllPrimitiveTypes(t *testing.T) {
	t.Parallel()
	type Target struct {
		S string
		I int
		U uint
		F float64
		B bool
	}
	binder := &defaultBinder{}

	tests := []struct {
		name      string
		value     string
		fieldName string
	}{
		{"string", "hello", "S"},
		{"int", "42", "I"},
		{"uint", "42", "U"},
		{"float64", "3.14", "F"},
		{"bool", "true", "B"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target := &Target{}
			val := reflect.ValueOf(target).Elem().FieldByName(tt.fieldName)
			err := binder.setFieldValue(val, tt.value, val.Type())
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

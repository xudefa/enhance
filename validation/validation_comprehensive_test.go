package validation

import (
	"reflect"
	"regexp"
	"sync"
	"testing"
)

func TestValidatorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("Register and Get", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()

		v := &mockCustomValidator{valid: true, msg: "ok"}
		r.Register("test", v)

		got, ok := r.Get("test")
		if !ok {
			t.Fatal("expected to find validator")
		}
		if got != v {
			t.Error("expected same validator instance")
		}
	})

	t.Run("Get missing returns false", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()
		_, ok := r.Get("nonexistent")
		if ok {
			t.Error("expected false for missing validator")
		}
	})

	t.Run("RegisterFunc and GetFunc", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()

		fn := func(field reflect.Value, param string) (bool, string) {
			return true, ""
		}
		r.RegisterFunc("myFunc", fn)

		got, ok := r.GetFunc("myFunc")
		if !ok {
			t.Fatal("expected to find func validator")
		}
		if got == nil {
			t.Error("expected non-nil function")
		}
	})

	t.Run("GetFunc missing returns false", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()
		_, ok := r.GetFunc("nonexistent")
		if ok {
			t.Error("expected false for missing func validator")
		}
	})

	t.Run("Unregister removes validator", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()

		r.Register("test", &mockCustomValidator{})
		r.RegisterFunc("test", func(reflect.Value, string) (bool, string) { return true, "" })

		r.Unregister("test")

		_, ok := r.Get("test")
		if ok {
			t.Error("expected validator to be removed")
		}
		_, ok = r.GetFunc("test")
		if ok {
			t.Error("expected func validator to be removed")
		}
	})

	t.Run("concurrent access", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()

		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func(idx int) {
				defer wg.Done()
				name := "v" + string(rune('0'+idx%10))
				r.Register(name, &mockCustomValidator{})
				r.Get(name)
				r.RegisterFunc(name, func(reflect.Value, string) (bool, string) { return true, "" })
				r.GetFunc(name)
			}(i)
		}
		wg.Wait()
	})
}

func TestPool(t *testing.T) {
	t.Parallel()

	t.Run("acquire and release validation errors", func(t *testing.T) {
		t.Parallel()
		p := acquireValidationErrors()
		if p == nil {
			t.Fatal("expected non-nil pool slice")
		}

		*p = append(*p, ValidationError{Field: "test", Message: "error"})

		releaseValidationErrors(p)
	})

	t.Run("acquire returns reset slice", func(t *testing.T) {
		t.Parallel()
		p1 := acquireValidationErrors()
		*p1 = append(*p1, ValidationError{Field: "test"})
		releaseValidationErrors(p1)

		p2 := acquireValidationErrors()
		if len(*p2) != 0 {
			t.Errorf("expected empty slice after reset, got length %d", len(*p2))
		}
		releaseValidationErrors(p2)
	})
}

func TestCompileRegex(t *testing.T) {
	t.Parallel()

	t.Run("valid pattern compiles", func(t *testing.T) {
		t.Parallel()
		re := compileRegex(`^[a-z]+$`)
		if re == nil {
			t.Fatal("expected non-nil regexp")
		}
		if !re.MatchString("hello") {
			t.Error("expected 'hello' to match")
		}
	})

	t.Run("invalid pattern returns nil", func(t *testing.T) {
		t.Parallel()
		re := compileRegex(`[invalid`)
		if re != nil {
			t.Error("expected nil for invalid pattern")
		}
	})

	t.Run("cached pattern reuse", func(t *testing.T) {
		t.Parallel()
		re1 := compileRegex(`^test$`)
		re2 := compileRegex(`^test$`)
		if re1 != re2 {
			t.Error("expected same instance from cache")
		}
	})
}

func TestUnsafeFieldValues(t *testing.T) {
	t.Parallel()

	t.Run("getFieldValueUnsafe with valid values", func(t *testing.T) {
		t.Parallel()

		type TestStruct struct {
			Str     string
			Int     int
			Int8    int8
			Int16   int16
			Int32   int32
			Int64   int64
			Uint    uint
			Uint8   uint8
			Uint16  uint16
			Uint32  uint32
			Uint64  uint64
			Float32 float32
			Float64 float64
			Bool    bool
		}

		s := TestStruct{
			Str:     "hello",
			Int:     42,
			Int8:    8,
			Int16:   16,
			Int32:   32,
			Int64:   64,
			Uint:    10,
			Uint8:   80,
			Uint16:  160,
			Uint32:  320,
			Uint64:  640,
			Float32: 3.14,
			Float64: 2.718,
			Bool:    true,
		}

		rv := reflect.ValueOf(&s).Elem()

		tests := []struct {
			name      string
			fieldName string
			val       any
		}{
			{"string", "Str", "hello"},
			{"int", "Int", int64(42)},
			{"int8", "Int8", int8(8)},
			{"int16", "Int16", int16(16)},
			{"int32", "Int32", int32(32)},
			{"int64", "Int64", int64(64)},
			{"uint", "Uint", uint64(10)},
			{"uint8", "Uint8", uint8(80)},
			{"uint16", "Uint16", uint16(160)},
			{"uint32", "Uint32", uint32(320)},
			{"uint64", "Uint64", uint64(640)},
			{"float32", "Float32", float32(3.14)},
			{"float64", "Float64", float64(2.718)},
			{"bool", "Bool", true},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				t.Parallel()
				field := rv.FieldByName(tt.fieldName)
				got := getFieldValueUnsafe(field)
				if !reflect.DeepEqual(got, tt.val) {
					t.Errorf("expected %v (%T), got %v (%T)", tt.val, tt.val, got, got)
				}
			})
		}
	})

	t.Run("getFieldValueUnsafe with invalid value", func(t *testing.T) {
		t.Parallel()
		got := getFieldValueUnsafe(reflect.Value{})
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("fieldValueInterface with nil pointer", func(t *testing.T) {
		t.Parallel()
		var p *string
		rv := reflect.ValueOf(&p).Elem()
		got := fieldValueInterface(rv)
		if got != nil {
			t.Errorf("expected nil for nil pointer, got %v", got)
		}
	})

	t.Run("fieldValueInterface with non-nil pointer", func(t *testing.T) {
		t.Parallel()
		s := "hello"
		p := &s
		rv := reflect.ValueOf(&p).Elem()
		got := fieldValueInterface(rv)
		if got == nil {
			t.Error("expected non-nil for non-nil pointer")
		}
	})

	t.Run("unpackNonNilValue with addressable value", func(t *testing.T) {
		t.Parallel()
		type S struct{ X int }
		s := S{X: 42}
		rv := reflect.ValueOf(&s).Elem().FieldByName("X")
		got := unpackNonNilValue(rv)
		if got == nil {
			t.Error("expected non-nil value")
		}
	})
}

func TestSplitRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		tag      string
		expected []string
	}{
		{
			name:     "simple rules",
			tag:      "required,min=3,max=10",
			expected: []string{"required", "min=3", "max=10"},
		},
		{
			name:     "regexp at end",
			tag:      "required,regexp=^[a-z]+$",
			expected: []string{"required", "regexp=^[a-z]+$"},
		},
		{
			name:     "regexp with commas in pattern",
			tag:      "required,regexp=^\\d{1,3}$",
			expected: []string{"required", "regexp=^\\d{1,3}$"},
		},
		{
			name:     "empty tag",
			tag:      "",
			expected: []string{""},
		},
		{
			name:     "single rule",
			tag:      "required",
			expected: []string{"required"},
		},
		{
			name:     "regexp with multiple commas in pattern",
			tag:      "regexp=^(a|b|c)$,required",
			expected: []string{"regexp=^(a|b|c)$,required"},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitRules(tt.tag)
			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d rules, got %d: %v", len(tt.expected), len(got), got)
			}
			for i := range got {
				if got[i] != tt.expected[i] {
					t.Errorf("rule %d: expected '%s', got '%s'", i, tt.expected[i], got[i])
				}
			}
		})
	}
}

func TestValidationErrorEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("ValidationError.Error() format", func(t *testing.T) {
		t.Parallel()
		ve := ValidationError{Field: "name", Message: "is required"}
		got := ve.Error()
		if got != "name: is required" {
			t.Errorf("expected 'name: is required', got '%s'", got)
		}
	})

	t.Run("ValidationErrors.Error() joins multiple", func(t *testing.T) {
		t.Parallel()
		errs := ValidationErrors{
			{Field: "name", Message: "is required"},
			{Field: "email", Message: "invalid format"},
		}
		got := errs.Error()
		if got != "name: is required; email: invalid format" {
			t.Errorf("unexpected error string: '%s'", got)
		}
	})

	t.Run("empty ValidationErrors", func(t *testing.T) {
		t.Parallel()
		errs := ValidationErrors{}
		got := errs.Error()
		if got != "" {
			t.Errorf("expected empty string, got '%s'", got)
		}
	})
}

func TestNewTagValidatorWithRegistry(t *testing.T) {
	t.Parallel()

	t.Run("creates validator with registry", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()
		v := NewTagValidatorWithRegistry(r)
		if v == nil {
			t.Fatal("expected non-nil validator")
		}
		if v.registry != r {
			t.Error("expected same registry instance")
		}
	})

	t.Run("custom func validator is used", func(t *testing.T) {
		t.Parallel()
		r := NewValidatorRegistry()
		r.RegisterFunc("custom", func(field reflect.Value, param string) (bool, string) {
			if field.Kind() != reflect.String {
				return false, "must be string"
			}
			if field.String() == "forbidden" {
				return false, "value is forbidden"
			}
			return true, ""
		})

		v := NewTagValidatorWithRegistry(r)

		type TestStruct struct {
			Value string `validate:"custom"`
		}

		err := v.Validate(TestStruct{Value: "allowed"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = v.Validate(TestStruct{Value: "forbidden"})
		if err == nil {
			t.Error("expected error for forbidden value")
		}
	})
}

func TestValidateField(t *testing.T) {
	t.Parallel()

	t.Run("gt validation for different types", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			IntVal  int    `validate:"gt=5"`
			StrVal  string `validate:"gt=3"`
			UintVal uint   `validate:"gt=10"`
		}

		err := v.Validate(S{IntVal: 10, StrVal: "hello", UintVal: 20})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{IntVal: 3, StrVal: "hi", UintVal: 5})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("gte validation", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"gte=5"`
		}

		err := v.Validate(S{Val: 5})
		if err != nil {
			t.Errorf("expected valid for equal value, got %v", err)
		}

		err = v.Validate(S{Val: 3})
		if err == nil {
			t.Error("expected invalid for lesser value")
		}
	})

	t.Run("lt validation", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"lt=10"`
		}

		err := v.Validate(S{Val: 5})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Val: 15})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("lte validation", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"lte=10"`
		}

		err := v.Validate(S{Val: 10})
		if err != nil {
			t.Errorf("expected valid for equal value, got %v", err)
		}

		err = v.Validate(S{Val: 15})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("len validation for string", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val string `validate:"len=5"`
		}

		err := v.Validate(S{Val: "hello"})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Val: "hi"})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("len validation for slice", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Items []int `validate:"len=3"`
		}

		err := v.Validate(S{Items: []int{1, 2, 3}})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Items: []int{1, 2}})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("invalid regex pattern returns error", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val string `validate:"regexp=[invalid"`
		}

		err := v.Validate(S{Val: "test"})
		if err == nil {
			t.Error("expected error for invalid regex")
		}
	})

	t.Run("oneof with int", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Status int `validate:"oneof=1 2 3"`
		}

		err := v.Validate(S{Status: 2})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Status: 5})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("oneof with uint", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Status uint `validate:"oneof=1 2 3"`
		}

		err := v.Validate(S{Status: 2})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Status: 5})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("oneof with float", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Score float64 `validate:"oneof=1.0 2.0 3.0"`
		}

		err := v.Validate(S{Score: 2.0})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}

		err = v.Validate(S{Score: 5.0})
		if err == nil {
			t.Error("expected invalid")
		}
	})

	t.Run("non-string email field", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"email"`
		}

		err := v.Validate(S{Val: 42})
		if err == nil {
			t.Error("expected error for non-string email field")
		}
	})

	t.Run("non-string URL field", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"url"`
		}

		err := v.Validate(S{Val: 42})
		if err == nil {
			t.Error("expected error for non-string URL field")
		}
	})

	t.Run("non-string IP field", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type S struct {
			Val int `validate:"ip"`
		}

		err := v.Validate(S{Val: 42})
		if err == nil {
			t.Error("expected error for non-string IP field")
		}
	})

	t.Run("required with non-zero struct type", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()

		type Inner struct{ Name string }
		type S struct {
			Val Inner `validate:"required"`
		}

		err := v.Validate(S{Val: Inner{Name: "test"}})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}
	})
}

func TestValidateEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil object", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()
		err := v.Validate(nil)
		if err != nil {
			t.Errorf("expected nil error for nil object, got %v", err)
		}
	})

	t.Run("nil pointer to struct", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()
		type S struct{ Name string `validate:"required"` }
		var s *S
		err := v.Validate(s)
		if err != nil {
			t.Errorf("expected nil error for nil pointer, got %v", err)
		}
	})

	t.Run("non-struct type", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()
		err := v.Validate("not a struct")
		if err == nil {
			t.Error("expected error for non-struct type")
		}
	})

	t.Run("struct with no validate tags", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()
		type S struct {
			Name string
			Age  int
		}
		err := v.Validate(S{Name: "test", Age: 25})
		if err != nil {
			t.Errorf("expected nil error for struct without tags, got %v", err)
		}
	})

	t.Run("struct with JSON tag for field name", func(t *testing.T) {
		t.Parallel()
		v := NewTagValidator()
		type S struct {
			Name string `json:"user_name" validate:"required"`
		}
		err := v.Validate(S{})
		if err == nil {
			t.Error("expected error for empty required field")
		}
	})
}

func TestValidateStructFunction(t *testing.T) {
	t.Parallel()

	t.Run("valid struct", func(t *testing.T) {
		t.Parallel()
		type S struct {
			Name string `validate:"required,min=2"`
		}
		err := ValidateStruct(S{Name: "hello"})
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("invalid struct", func(t *testing.T) {
		t.Parallel()
		type S struct {
			Name string `validate:"required,min=2"`
		}
		err := ValidateStruct(S{Name: "a"})
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestValidateFunction(t *testing.T) {
	t.Parallel()

	t.Run("valid string", func(t *testing.T) {
		t.Parallel()
		err := Validate("hello", "required")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("valid int", func(t *testing.T) {
		t.Parallel()
		err := Validate(50, "min=1,max=100")
		if err != nil {
			t.Errorf("expected nil, got %v", err)
		}
	})

	t.Run("invalid int", func(t *testing.T) {
		t.Parallel()
		err := Validate(200, "max=100")
		if err == nil {
			t.Error("expected error")
		}
	})
}

func TestRequiredValidationEdgeCases(t *testing.T) {
	t.Parallel()
	v := NewTagValidator()

	type S struct {
		BoolVal    bool   `validate:"required"`
		FloatVal   float64 `validate:"required"`
		UintVal    uint   `validate:"required"`
		StringVal  string `validate:"required"`
	}

	t.Run("zero values fail required", func(t *testing.T) {
		t.Parallel()
		err := v.Validate(S{})
		if err == nil {
			t.Error("expected error for zero values")
		}
	})

	t.Run("non-zero values pass required", func(t *testing.T) {
		t.Parallel()
		err := v.Validate(S{
			BoolVal:   true,
			FloatVal:  1.0,
			UintVal:   1,
			StringVal: "x",
		})
		if err != nil {
			t.Errorf("expected valid, got %v", err)
		}
	})
}

// mockCustomValidator is a test mock for CustomValidator interface
type mockCustomValidator struct {
	valid bool
	msg   string
}

func (m *mockCustomValidator) Validate(field reflect.Value, param string) (bool, string) {
	return m.valid, m.msg
}

// Ensure regexp is used to prevent unused import errors
var _ = regexp.MustCompile

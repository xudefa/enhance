package validation

import (
	"reflect"
	"regexp"
	"sync"
	"testing"
)

func testValidatorRegistryRegisterGet(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()

	customValidator := &mockCustomValidator{valid: true, msg: "ok"}
	registry.Register("test", customValidator)

	got, ok := registry.Get("test")
	if !ok {
		t.Fatal("expected to find validator")
	}
	if got != customValidator {
		t.Error("expected same validator instance")
	}
}

func testValidatorRegistryGetMissing(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	_, ok := registry.Get("nonexistent")
	if ok {
		t.Error("expected false for missing validator")
	}
}

func testValidatorRegistryRegisterFuncGetFunc(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()

	fn := func(field reflect.Value, param string) (bool, string) {
		return true, ""
	}
	registry.RegisterFunc("myFunc", fn)

	got, ok := registry.GetFunc("myFunc")
	if !ok {
		t.Fatal("expected to find func validator")
	}
	if got == nil {
		t.Error("expected non-nil function")
	}
}

func testValidatorRegistryGetFuncMissing(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()
	_, ok := registry.GetFunc("nonexistent")
	if ok {
		t.Error("expected false for missing func validator")
	}
}

func testValidatorRegistryUnregister(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()

	registry.Register("test", &mockCustomValidator{})
	registry.RegisterFunc("test", func(reflect.Value, string) (bool, string) { return true, "" })

	registry.Unregister("test")

	_, ok := registry.Get("test")
	if ok {
		t.Error("expected validator to be removed")
	}
	_, ok = registry.GetFunc("test")
	if ok {
		t.Error("expected func validator to be removed")
	}
}

func testValidatorRegistryConcurrent(t *testing.T) {
	t.Parallel()
	registry := NewValidatorRegistry()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			name := "v" + string(rune('0'+idx%10))
			registry.Register(name, &mockCustomValidator{})
			registry.Get(name)
			registry.RegisterFunc(name, func(reflect.Value, string) (bool, string) { return true, "" })
			registry.GetFunc(name)
		}(i)
	}
	wg.Wait()
}

func TestValidatorRegistry(t *testing.T) {
	t.Parallel()

	t.Run("Register and Get", testValidatorRegistryRegisterGet)
	t.Run("Get missing returns false", testValidatorRegistryGetMissing)
	t.Run("RegisterFunc and GetFunc", testValidatorRegistryRegisterFuncGetFunc)
	t.Run("GetFunc missing returns false", testValidatorRegistryGetFuncMissing)
	t.Run("Unregister removes validator", testValidatorRegistryUnregister)
	t.Run("concurrent access", testValidatorRegistryConcurrent)
}

func TestPool(t *testing.T) {
	t.Parallel()

	t.Run("acquire and release validation errors", func(t *testing.T) {
		t.Parallel()
		errSlice := acquireValidationErrors()
		if errSlice == nil {
			t.Fatal("expected non-nil pool slice")
		}

		*errSlice = append(*errSlice, ValidationError{Field: "test", Message: "error"})

		releaseValidationErrors(errSlice)
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

func testUnsafeFieldValueCases() []struct {
	name      string
	fieldName string
	val       any
} {
	return []struct {
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
}

func testUnsafeFieldValuesValid(t *testing.T) {
	t.Parallel()

	rv := reflect.ValueOf(newTestStructSample()).Elem()

	for _, tt := range testUnsafeFieldValueCases() {
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
}

type testFieldStruct struct {
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

func newTestStructSample() *testFieldStruct {
	return &testFieldStruct{
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
}

func TestUnsafeFieldValues(t *testing.T) {
	t.Parallel()

	t.Run("getFieldValueUnsafe with valid values", testUnsafeFieldValuesValid)

	t.Run("getFieldValueUnsafe with invalid value", func(t *testing.T) {
		t.Parallel()
		got := getFieldValueUnsafe(reflect.Value{})
		if got != nil {
			t.Errorf("expected nil, got %v", got)
		}
	})

	t.Run("fieldValueInterface with nil pointer", func(t *testing.T) {
		t.Parallel()
		var ptr *string
		rv := reflect.ValueOf(&ptr).Elem()
		got := fieldValueInterface(rv)
		if got != nil {
			t.Errorf("expected nil for nil pointer, got %v", got)
		}
	})

	t.Run("fieldValueInterface with non-nil pointer", func(t *testing.T) {
		t.Parallel()
		sample := "hello"
		ptr := &sample
		rv := reflect.ValueOf(&ptr).Elem()
		got := fieldValueInterface(rv)
		if got == nil {
			t.Error("expected non-nil for non-nil pointer")
		}
	})

	t.Run("unpackNonNilValue with addressable value", func(t *testing.T) {
		t.Parallel()
		type S struct{ X int }
		sample := S{X: 42}
		rv := reflect.ValueOf(&sample).Elem().FieldByName("X")
		got := unpackNonNilValue(rv)
		if got == nil {
			t.Error("expected non-nil value")
		}
	})
}

func testSplitRuleCases() []struct {
	name     string
	tag      string
	expected []string
} {
	return []struct {
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
}

func TestSplitRules(t *testing.T) {
	t.Parallel()

	for _, tt := range testSplitRuleCases() {
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

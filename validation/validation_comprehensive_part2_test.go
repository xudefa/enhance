package validation

import (
	"reflect"
	"testing"
)

func TestNewTagValidatorWithRegistry(t *testing.T) {
	t.Parallel()

	t.Run("creates validator with registry", func(t *testing.T) {
		t.Parallel()
		registry := NewValidatorRegistry()
		validator := NewTagValidatorWithRegistry(registry)
		if validator == nil {
			t.Fatal("expected non-nil validator")
		}
		if validator.registry != registry {
			t.Error("expected same registry instance")
		}
	})

	t.Run("custom func validator is used", func(t *testing.T) {
		t.Parallel()
		registry := NewValidatorRegistry()
		registry.RegisterFunc("custom", func(field reflect.Value, param string) (bool, string) {
			if field.Kind() != reflect.String {
				return false, "must be string"
			}
			if field.String() == "forbidden" {
				return false, "value is forbidden"
			}
			return true, ""
		})

		validator := NewTagValidatorWithRegistry(registry)

		type TestStruct struct {
			Value string `validate:"custom"`
		}

		err := validator.Validate(TestStruct{Value: "allowed"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		err = validator.Validate(TestStruct{Value: "forbidden"})
		if err == nil {
			t.Error("expected error for forbidden value")
		}
	})
}

func testValidateFieldGt(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		IntVal  int    `validate:"gt=5"`
		StrVal  string `validate:"gt=3"`
		UintVal uint   `validate:"gt=10"`
	}

	err := validator.Validate(S{IntVal: 10, StrVal: "hello", UintVal: 20})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{IntVal: 3, StrVal: "hi", UintVal: 5})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldGte(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"gte=5"`
	}

	err := validator.Validate(S{Val: 5})
	if err != nil {
		t.Errorf("expected valid for equal value, got %v", err)
	}

	err = validator.Validate(S{Val: 3})
	if err == nil {
		t.Error("expected invalid for lesser value")
	}
}

func testValidateFieldLt(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"lt=10"`
	}

	err := validator.Validate(S{Val: 5})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Val: 15})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldLte(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"lte=10"`
	}

	err := validator.Validate(S{Val: 10})
	if err != nil {
		t.Errorf("expected valid for equal value, got %v", err)
	}

	err = validator.Validate(S{Val: 15})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldLenString(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val string `validate:"len=5"`
	}

	err := validator.Validate(S{Val: "hello"})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Val: "hi"})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldLenSlice(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Items []int `validate:"len=3"`
	}

	err := validator.Validate(S{Items: []int{1, 2, 3}})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Items: []int{1, 2}})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldInvalidRegex(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val string `validate:"regexp=[invalid"`
	}

	err := validator.Validate(S{Val: "test"})
	if err == nil {
		t.Error("expected error for invalid regex")
	}
}

func testValidateFieldOneOfInt(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Status int `validate:"oneof=1 2 3"`
	}

	err := validator.Validate(S{Status: 2})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Status: 5})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldOneOfUint(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Status uint `validate:"oneof=1 2 3"`
	}

	err := validator.Validate(S{Status: 2})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Status: 5})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldOneOfFloat(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Score float64 `validate:"oneof=1.0 2.0 3.0"`
	}

	err := validator.Validate(S{Score: 2.0})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}

	err = validator.Validate(S{Score: 5.0})
	if err == nil {
		t.Error("expected invalid")
	}
}

func testValidateFieldEmailField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"email"`
	}

	err := validator.Validate(S{Val: 42})
	if err == nil {
		t.Error("expected error for non-string email field")
	}
}

func testValidateFieldURLField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"url"`
	}

	err := validator.Validate(S{Val: 42})
	if err == nil {
		t.Error("expected error for non-string URL field")
	}
}

func testValidateFieldIPField(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type S struct {
		Val int `validate:"ip"`
	}

	err := validator.Validate(S{Val: 42})
	if err == nil {
		t.Error("expected error for non-string IP field")
	}
}

func testValidateFieldRequiredStruct(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()

	type Inner struct{ Name string }
	type S struct {
		Val Inner `validate:"required"`
	}

	err := validator.Validate(S{Val: Inner{Name: "test"}})
	if err != nil {
		t.Errorf("expected valid, got %v", err)
	}
}

func TestValidateField(t *testing.T) {
	t.Parallel()
	t.Run("gt validation for different types", testValidateFieldGt)
	t.Run("gte validation", testValidateFieldGte)
	t.Run("lt validation", testValidateFieldLt)
	t.Run("lte validation", testValidateFieldLte)
	t.Run("len validation for string", testValidateFieldLenString)
	t.Run("len validation for slice", testValidateFieldLenSlice)
	t.Run("invalid regex pattern returns error", testValidateFieldInvalidRegex)
	t.Run("oneof with int", testValidateFieldOneOfInt)
	t.Run("oneof with uint", testValidateFieldOneOfUint)
	t.Run("oneof with float", testValidateFieldOneOfFloat)
	t.Run("non-string email field", testValidateFieldEmailField)
	t.Run("non-string URL field", testValidateFieldURLField)
	t.Run("non-string IP field", testValidateFieldIPField)
	t.Run("required with non-zero struct type", testValidateFieldRequiredStruct)
}

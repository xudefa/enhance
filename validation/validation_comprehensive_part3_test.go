package validation

import (
	"testing"
)

func testValidateEdgeNilObject(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()
	err := validator.Validate(nil)
	if err != nil {
		t.Errorf("expected nil error for nil object, got %v", err)
	}
}

func testValidateEdgeNilPointer(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()
	type S struct {
		Name string `validate:"required"`
	}
	var s *S
	err := validator.Validate(s)
	if err != nil {
		t.Errorf("expected nil error for nil pointer, got %v", err)
	}
}

func testValidateEdgeNonStruct(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()
	err := validator.Validate("not a struct")
	if err == nil {
		t.Error("expected error for non-struct type")
	}
}

func testValidateEdgeNoTags(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()
	type S struct {
		Name string
		Age  int
	}
	err := validator.Validate(S{Name: "test", Age: 25})
	if err != nil {
		t.Errorf("expected nil error for struct without tags, got %v", err)
	}
}

func testValidateEdgeJSONTag(t *testing.T) {
	t.Parallel()
	validator := NewTagValidator()
	type S struct {
		Name string `json:"user_name" validate:"required"`
	}
	err := validator.Validate(S{})
	if err == nil {
		t.Error("expected error for empty required field")
	}
}

func TestValidateEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("nil object", testValidateEdgeNilObject)
	t.Run("nil pointer to struct", testValidateEdgeNilPointer)
	t.Run("non-struct type", testValidateEdgeNonStruct)
	t.Run("struct with no validate tags", testValidateEdgeNoTags)
	t.Run("struct with JSON tag for field name", testValidateEdgeJSONTag)
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
	validator := NewTagValidator()

	type S struct {
		BoolVal   bool    `validate:"required"`
		FloatVal  float64 `validate:"required"`
		UintVal   uint    `validate:"required"`
		StringVal string  `validate:"required"`
	}

	t.Run("zero values fail required", func(t *testing.T) {
		t.Parallel()
		err := validator.Validate(S{})
		if err == nil {
			t.Error("expected error for zero values")
		}
	})

	t.Run("non-zero values pass required", func(t *testing.T) {
		t.Parallel()
		err := validator.Validate(S{
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

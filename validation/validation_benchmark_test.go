package validation

import (
	"testing"
)

// BenchmarkTagValidator 测试标签验证器性能
func BenchmarkTagValidator(b *testing.B) {
	type TestStruct struct {
		Name  string `validate:"required,min=3,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=1,max=150"`
	}

	validator := NewTagValidator()

	b.Run("Valid-Struct", func(b *testing.B) {
		obj := &TestStruct{
			Name:  "John Doe",
			Email: "john@example.com",
			Age:   30,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(obj)
		}
	})

	b.Run("Invalid-Struct", func(b *testing.B) {
		obj := &TestStruct{
			Name:  "Jo",
			Email: "invalid-email",
			Age:   200,
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(obj)
		}
	})
}

// BenchmarkTagValidator_Concurrent 测试并发验证性能
func BenchmarkTagValidator_Concurrent(b *testing.B) {
	type TestStruct struct {
		Name  string `validate:"required,min=3,max=50"`
		Email string `validate:"required,email"`
	}

	validator := NewTagValidator()
	obj := &TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = validator.Validate(obj)
		}
	})
}

// BenchmarkTagValidator_DifferentRules 测试不同验证规则的性能
func BenchmarkTagValidator_DifferentRules(b *testing.B) {
	type SimpleStruct struct {
		Name string `validate:"required"`
	}

	type ComplexStruct struct {
		Name    string `validate:"required,min=3,max=50"`
		Email   string `validate:"required,email"`
		Age     int    `validate:"min=1,max=150"`
		Phone   string `validate:"required"`
		Address string `validate:"min=10,max=200"`
	}

	validator := NewTagValidator()

	b.Run("Simple-Required", func(b *testing.B) {
		obj := &SimpleStruct{Name: "Test"}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(obj)
		}
	})

	b.Run("Complex-Multiple-Rules", func(b *testing.B) {
		obj := &ComplexStruct{
			Name:    "John Doe",
			Email:   "john@example.com",
			Age:     30,
			Phone:   "1234567890",
			Address: "123 Main Street, City, Country",
		}
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = validator.Validate(obj)
		}
	})
}

// BenchmarkRuleBuilder 测试规则构建器性能
func BenchmarkRuleBuilder(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRuleBuilder().
				Required().
				Min(3).
				Max(50).
				Build()
		}
	})

	b.Run("Complex", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = NewRuleBuilder().
				Required().
				Email().
				Min(3).
				Max(100).
				Build()
		}
	})
}

// BenchmarkValidateStruct 测试 ValidateStruct 函数性能
func BenchmarkValidateStruct(b *testing.B) {
	type TestStruct struct {
		Name  string `validate:"required,min=3,max=50"`
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=1,max=150"`
	}

	obj := &TestStruct{
		Name:  "John Doe",
		Email: "john@example.com",
		Age:   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ValidateStruct(obj)
	}
}

// BenchmarkValidate 测试 Validate 函数性能
func BenchmarkValidate(b *testing.B) {
	b.Run("Required", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Validate("test", "required")
		}
	})

	b.Run("Email", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Validate("test@example.com", "email")
		}
	})

	b.Run("Multiple-Rules", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = Validate("test", "required,min=3,max=50")
		}
	})
}

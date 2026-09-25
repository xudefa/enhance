// Package main demonstrates the Validator starter usage.
//
// This example shows how to use the Validator starter to:
// 1. Auto-configure validator
// 2. Validate structs
// 3. Use custom validators (phone, idcard)
//
// Run:
//
//	go run main.go
package main

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/xudefa/enhance/boot"
	"github.com/xudefa/enhance/core"

	_ "github.com/xudefa/enhance/starter/validator"
)

// User represents a user struct for validation
type User struct {
	Name     string `validate:"required,min=2,max=50"`
	Email    string `validate:"required,email"`
	Age      int    `validate:"gte=0,lte=130"`
	Phone    string `validate:"omitempty,phone"`
	IDCard   string `validate:"omitempty,idcard"`
	Password string `validate:"required,min=8"`
}

func main() {
	fmt.Println("=== Validator Starter Example ===")
	fmt.Println()

	app := newApp()
	defer app.Stop()

	if err := app.Start(); err != nil {
		fmt.Printf("Failed to start application: %v\n", err)
		return
	}

	validate, err := getValidator(app)
	if err != nil {
		return
	}

	runStructValidation(validate)
	runFieldValidation(validate)

	fmt.Println("\n=== Example completed successfully ===")
}

// newApp 创建应用实例，配置名称与激活的 Profile。
func newApp() *boot.Boot {
	app, err := boot.NewApplication(
		boot.WithAppName("validator-example"),
		boot.WithProfiles("default"),
	)
	if err != nil {
		panic(fmt.Sprintf("Failed to create application: %v", err))
	}
	return app
}

// getValidator 从容器的 Bean 中获取校验器实例。
func getValidator(app *boot.Boot) (*validator.Validate, error) {
	validate, err := core.GetByName[*validator.Validate](app.Container(), "")
	if err != nil {
		fmt.Printf("Failed to get validator: %v\n", err)
		return nil, fmt.Errorf("Failed to get validator: %w", err)
	}
	return validate, nil
}

// runStructValidation 演示对结构体的整体校验（Test 1-3）。
func runStructValidation(validate *validator.Validate) {
	// Test 1: Valid user
	fmt.Println("--- Test 1: Valid User ---")
	validUser := User{
		Name:     "John Doe",
		Email:    "john.doe@example.com",
		Age:      30,
		Phone:    "13800138000",
		IDCard:   "110101199001011234",
		Password: "securepassword",
	}

	if err := validate.Struct(validUser); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
	} else {
		fmt.Println("Validation passed!")
	}

	// Test 2: Invalid user (missing required fields)
	fmt.Println("\n--- Test 2: Invalid User (Missing Required Fields) ---")
	invalidUser := User{
		Name:  "",        // Required
		Email: "invalid", // Invalid email
		Age:   -5,        // Invalid age
	}

	if err := validate.Struct(invalidUser); err != nil {
		fmt.Printf("Validation failed (expected):\n%v\n", err)
	} else {
		fmt.Println("Validation passed (unexpected)")
	}

	// Test 3: Invalid user (password too short)
	fmt.Println("\n--- Test 3: Invalid User (Password Too Short) ---")
	shortPasswordUser := User{
		Name:     "Jane Doe",
		Email:    "jane.doe@example.com",
		Age:      25,
		Password: "123", // Too short
	}

	if err := validate.Struct(shortPasswordUser); err != nil {
		fmt.Printf("Validation failed (expected):\n%v\n", err)
	} else {
		fmt.Println("Validation passed (unexpected)")
	}
}

// runFieldValidation 演示字段级与跨字段校验（Test 4-5）。
func runFieldValidation(validate *validator.Validate) {
	// Test 4: Field-level validation
	fmt.Println("\n--- Test 4: Field-Level Validation ---")
	field := "email"
	value := "not-an-email"

	if err := validate.Var(value, field); err != nil {
		fmt.Printf("Field validation failed for %s: %v\n", field, err)
	} else {
		fmt.Printf("Field validation passed for %s\n", field)
	}

	// Test 5: Cross-field validation
	fmt.Println("\n--- Test 5: Cross-Field Validation ---")
	type PasswordConfirm struct {
		Password        string `validate:"required"`
		PasswordConfirm string `validate:"required,eqfield=Password"`
	}

	validPC := PasswordConfirm{
		Password:        "securepassword",
		PasswordConfirm: "securepassword",
	}

	if err := validate.Struct(validPC); err != nil {
		fmt.Printf("Cross-field validation failed: %v\n", err)
	} else {
		fmt.Println("Cross-field validation passed!")
	}

	invalidPC := PasswordConfirm{
		Password:        "securepassword",
		PasswordConfirm: "differentpassword",
	}

	if err := validate.Struct(invalidPC); err != nil {
		fmt.Printf("Cross-field validation failed (expected):\n%v\n", err)
	} else {
		fmt.Println("Cross-field validation passed (unexpected)")
	}
}

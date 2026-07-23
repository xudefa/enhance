package main

import (
	"fmt"

	"github.com/xudefa/enhance/validation"
)

type CreateUserRequest struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age" validate:"required,gt=0,lte=130"`
}

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=2,max=50"`
}

func main() {
	fmt.Println("=== Validation Demo ===")

	validator := validation.NewTagValidator()

	fmt.Println("\n1. Valid request:")
	validReq := CreateUserRequest{
		Name:  "Alice",
		Email: "alice@example.com",
		Age:   25,
	}
	if err := validator.Validate(validReq); err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   PASS: validation succeeded")
	}

	fmt.Println("\n2. Invalid request (missing name):")
	invalidReq1 := CreateUserRequest{
		Email: "bob@example.com",
		Age:   30,
	}
	if err := validator.Validate(invalidReq1); err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   PASS: validation succeeded")
	}

	fmt.Println("\n3. Invalid request (bad email):")
	invalidReq2 := CreateUserRequest{
		Name:  "Charlie",
		Email: "not-an-email",
		Age:   20,
	}
	if err := validator.Validate(invalidReq2); err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   PASS: validation succeeded")
	}

	fmt.Println("\n4. Invalid request (age out of range):")
	invalidReq3 := CreateUserRequest{
		Name:  "Dave",
		Email: "dave@example.com",
		Age:   -5,
	}
	if err := validator.Validate(invalidReq3); err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   PASS: validation succeeded")
	}

	fmt.Println("\n5. Name length validation:")
	shortName := UpdateUserRequest{Name: "X"}
	if err := validator.Validate(shortName); err != nil {
		fmt.Printf("   Short name ERROR: %v\n", err)
	}
	longName := UpdateUserRequest{Name: "ThisNameIsWayTooLongForThe50CharacterLimitAndShouldFail"}
	if err := validator.Validate(longName); err != nil {
		fmt.Printf("   Long name ERROR: %v\n", err)
	}
	goodName := UpdateUserRequest{Name: "GoodName"}
	if err := validator.Validate(goodName); err != nil {
		fmt.Printf("   ERROR: %v\n", err)
	} else {
		fmt.Println("   Good name: PASS")
	}

	fmt.Println("\n=== Validation Demo Complete ===")
}

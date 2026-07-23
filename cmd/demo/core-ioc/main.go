package main

import (
	"fmt"

	"github.com/xudefa/enhance/core"
)

type UserRepository struct {
	data map[string]string
}

func NewUserRepository() *UserRepository {
	return &UserRepository{
		data: map[string]string{
			"1": "Alice",
			"2": "Bob",
		},
	}
}

func (r *UserRepository) Find(id string) (string, bool) {
	name, ok := r.data[id]
	return name, ok
}

type UserService struct {
	repo *UserRepository
}

func NewUserService(repo *UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUser(id string) (string, bool) {
	return s.repo.Find(id)
}

func main() {
	fmt.Println("=== Core IoC Container Demo ===")

	c := core.NewContainer()

	fmt.Println("\n1. Registering UserRepository:")
	userRepo := NewUserRepository()
	_ = core.Register[*UserRepository](c,
		core.WithFactory[*UserRepository](func(_ ...any) (any, error) {
			return userRepo, nil
		}),
	)

	fmt.Println("\n2. Registering UserService with dependency:")
	userSvc := NewUserService(userRepo)
	_ = core.Register[*UserService](c,
		core.WithFactory[*UserService](func(_ ...any) (any, error) {
			return userSvc, nil
		}),
	)

	fmt.Println("\n3. Getting beans from container:")
	svc, err := core.GetByName[*UserService](c, "")
	if err != nil {
		fmt.Printf("   ERROR getting UserService: %v\n", err)
		return
	}
	fmt.Printf("   Found UserService: %v\n", svc != nil)

	if name, ok := svc.GetUser("1"); ok {
		fmt.Printf("   GetUser(1) = %s\n", name)
	}
	if name, ok := svc.GetUser("2"); ok {
		fmt.Printf("   GetUser(2) = %s\n", name)
	}
	if _, ok := svc.GetUser("999"); !ok {
		fmt.Println("   GetUser(999) = not found")
	}

	fmt.Println("\n4. Checking bean existence:")
	fmt.Printf("   Has[*UserRepository]: %v\n", core.Has[*UserRepository](c, ""))

	fmt.Println("\n=== IoC Container Demo Complete ===")
}

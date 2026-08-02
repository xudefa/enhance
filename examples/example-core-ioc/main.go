// Package main demonstrates the enhance core IoC container features:
// container creation, bean registration, dependency injection,
// generic API, lifecycle management, and concurrent access safety.
package main

import (
	"fmt"
	"sync"

	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

// ==================== Domain Types ====================

// Database represents a database connection.
type Database struct {
	DSN string
}

// UserService depends on Database.
type UserService struct {
	DB  *Database
	Name string
}

// OrderService depends on both Database and UserService.
type OrderService struct {
	DB      *Database
	UserSvc *UserService
}

// RequestBean is a prototype-scope bean (new instance each time).
type RequestBean struct {
	ID int
}

func main() {
	fmt.Println("=== enhance IoC Container Example ===")
	fmt.Println()

	c := core.NewContainer()

	// ---- 1. Register a singleton Database via factory ----
	_ = core.Register[*Database](c,
		core.WithName[*Database]("db"),
		core.WithFactory[*Database](func(c ...any) (any, error) {
			fmt.Println("  [factory] Creating Database singleton")
			return &Database{DSN: "localhost:3306/mydb"}, nil
		}),
		core.WithInit[*Database](func(bean any) error {
			db := bean.(*Database)
			fmt.Printf("  [init] Database connected: %s\n", db.DSN)
			return nil
		}),
		core.WithDestroy[*Database](func(bean any) error {
			fmt.Println("  [destroy] Database connection closed")
			return nil
		}),
	)

	// ---- 2. Register UserService with dependency on Database ----
	_ = core.Register[*UserService](c,
		core.WithName[*UserService]("userService"),
		core.WithFactory[*UserService](func(c ...any) (any, error) {
			fmt.Println("  [factory] Creating UserService")
			return &UserService{Name: "UserService"}, nil
		}),
		core.WithInit[*UserService](func(bean any) error {
			fmt.Println("  [init] UserService started")
			return nil
		}),
		core.WithDestroy[*UserService](func(bean any) error {
			fmt.Println("  [destroy] UserService stopped")
			return nil
		}),
	)

	// ---- 3. Register OrderService with multiple dependencies ----
	_ = core.Register[*OrderService](c,
		core.WithName[*OrderService]("orderService"),
		core.WithFactory[*OrderService](func(c ...any) (any, error) {
			fmt.Println("  [factory] Creating OrderService")
			return &OrderService{}, nil
		}),
	)

	// ---- 4. Register a prototype bean ----
	_ = core.Register[*RequestBean](c,
		core.WithName[*RequestBean]("requestBean"),
		core.WithScope[*RequestBean](registry.Prototype),
		core.WithFactory[*RequestBean](func(c ...any) (any, error) {
			return &RequestBean{ID: 1}, nil
		}),
	)

	// ---- Initialize container (triggers Init callbacks) ----
	fmt.Println("--- Initializing container ---")
	if err := c.Initialize(); err != nil {
		panic(err)
	}

	// ---- 5. Retrieve and use beans via generic API ----
	fmt.Println()
	fmt.Println("--- Retrieving beans ---")

	db := core.MustGet[*Database](c, "db")
	fmt.Printf("  Database DSN: %s\n", db.DSN)

	// Manually wire dependencies (DI simulation)
	userSvc := core.MustGet[*UserService](c, "userService")
	userSvc.DB = db
	fmt.Printf("  UserService: %s (DB: %s)\n", userSvc.Name, userSvc.DB.DSN)

	orderSvc := core.MustGet[*OrderService](c, "orderService")
	orderSvc.DB = db
	orderSvc.UserSvc = userSvc
	fmt.Printf("  OrderService -> DB=%s, UserSvc=%s\n", orderSvc.DB.DSN, orderSvc.UserSvc.Name)

	// ---- 6. Prototype scope: each Get creates a new instance ----
	fmt.Println()
	fmt.Println("--- Prototype scope ---")
	req1 := core.MustGet[*RequestBean](c, "requestBean")
	req2 := core.MustGet[*RequestBean](c, "requestBean")
	fmt.Printf("  req1 == req2: %v (should be false)\n", req1 == req2)

	// ---- 7. Check bean existence ----
	fmt.Println()
	fmt.Println("--- Bean existence ---")
	fmt.Printf("  Has[*Database]: %v\n", core.Has[*Database](c, "db"))
	fmt.Printf("  Has[*UserService]: %v\n", core.Has[*UserService](c, "userService"))

	// ---- 8. Concurrent access safety test ----
	fmt.Println()
	fmt.Println("--- Concurrent access test ---")
	var wg sync.WaitGroup
	errs := make(chan error, 20)

	for i := 0; i < 10; i++ {
		wg.Add(2)
		go func() {
			defer wg.Done()
			_, err := core.GetByName[*Database](c, "db")
			if err != nil {
				errs <- err
			}
		}()
		go func() {
			defer wg.Done()
			_, err := core.GetByName[*UserService](c, "userService")
			if err != nil {
				errs <- err
			}
		}()
	}
	wg.Wait()
	close(errs)

	concurrentErrors := 0
	for err := range errs {
		concurrentErrors++
		fmt.Printf("  Error: %v\n", err)
	}
	if concurrentErrors == 0 {
		fmt.Println("  All 20 concurrent operations succeeded!")
	}

	// ---- Destroy container ----
	fmt.Println()
	fmt.Println("--- Destroying container ---")
	if err := c.Destroy(); err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("=== Example completed successfully ===")
}

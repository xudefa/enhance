# example-core-ioc

Demonstrates the enhance core IoC container features.

## Features Demonstrated

- **Container creation and configuration** — creating a new IoC container
- **Bean registration** — factory, singleton, and prototype scope beans
- **Dependency injection** — wiring dependencies between beans
- **Generic API** — type-safe `Register[T]`, `GetByName[T]`, `MustGet[T]`, `Has[T]`
- **Lifecycle management** — init and destroy callbacks
- **Concurrent access safety** — goroutine-safe bean retrieval

## Run

```bash
go run .
```

## Expected Output

```
=== enhance IoC Container Example ===

--- Initializing container ---
  [factory] Creating Database singleton
  [init] Database connected: localhost:3306/mydb
  [factory] Creating UserService
  [init] UserService started
  [factory] Creating OrderService

--- Retrieving beans ---
  Database DSN: localhost:3306/mydb
  [factory] Creating UserService
  UserService: UserService (DB: localhost:3306/mydb)
  OrderService -> DB=localhost:3306/mydb, UserSvc=UserService

--- Prototype scope ---
  req1 == req2: false (should be false)

--- Bean existence ---
  Has[*Database]: true
  Has[*UserService]: true

--- Concurrent access test ---
  All 20 concurrent operations succeeded!

--- Destroying container ---
  [destroy] UserService stopped
  [destroy] Database connection closed

=== Example completed successfully ===
```

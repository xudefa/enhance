# example-web-rest-api

Demonstrates the enhance web framework REST API capabilities.

## Features Demonstrated

- **HTTP server setup** — stdlib-based server with configuration
- **Route registration** — GET, POST with path parameters
- **Middleware** — logging and recovery middleware
- **Request handling** — JSON binding, path params, query params
- **Error handling** — proper HTTP status codes
- **Graceful shutdown** — signal-based shutdown

## Run

```bash
go run .
```

## Expected Output

```
=== enhance Web REST API Example ===

  Server starting on http://127.0.0.1:9999

--- Test Requests ---
  [GET /] -> 200 OK
  [GET /users] -> 200 OK
  [GET /users/1] -> 200 OK
  [POST /users] -> 201 OK
  [GET /health] -> 200 OK
  [GET /error] -> 500 ERROR

--- Graceful Shutdown ---
  Shutting down server...
  Server stopped gracefully

=== Example completed successfully ===
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | Welcome message |
| GET | `/users` | List all users |
| GET | `/users/:id` | Get user by ID |
| POST | `/users` | Create a new user |
| GET | `/health` | Health check |
| GET | `/error` | Error handling demo |

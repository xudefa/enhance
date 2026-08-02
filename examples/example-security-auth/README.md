# example-security-auth

Demonstrates the enhance security framework authentication and authorization features.

## Features Demonstrated

- **UserDetailsService** — in-memory user management
- **PasswordEncoder** — NoOp encoder for demo (use BCrypt in production)
- **AuthenticationManager** — DaoAuthenticationProvider for username/password auth
- **Role-based authorization** — WebExpressionVoter with role checks
- **Access decision managers** — AffirmativeBased and UnanimousBased strategies
- **SecurityBuilder** — fluent API for security configuration

## Run

```bash
go run .
```

## Expected Output

```
=== enhance Security Auth Example ===

--- 1. Creating Users ---
  Created 3 users

--- 2. Setting up Authentication Manager ---

--- 3. Authenticating Users ---
  admin authenticated: true, authorities: [ROLE_ADMIN ROLE_USER]
  user authenticated: true, authorities: [ROLE_USER]
  admin wrong password: bad credentials (expected)
  unknown user: bad credentials (expected)

--- 4. Role-Based Authorization ---
  Resource: /api/admin
    admin -> /api/admin [hasRole('ADMIN')]: GRANTED
    user -> /api/admin [hasRole('ADMIN')]: DENIED (access denied)
    manager -> /api/admin [hasRole('ADMIN')]: DENIED (access denied)
  Resource: /api/users
    admin -> /api/users [authenticated]: GRANTED
    user -> /api/users [authenticated]: GRANTED
  Resource: /api/reports (any of ADMIN, MANAGER)
    admin -> /api/reports [hasAnyRole('ADMIN','MANAGER')]: GRANTED
    user -> /api/reports [hasAnyRole('ADMIN','MANAGER')]: DENIED (access denied)
    manager -> /api/reports [hasAnyRole('ADMIN','MANAGER')]: GRANTED
  Resource: /api/secret
    admin -> /api/secret [denyAll]: DENIED (access denied)
  Resource: /public
    user -> /public [permitAll]: GRANTED

--- 5. SecurityBuilder Configuration ---
  Security config built: *security.builtSecurityConfig

--- 6. Unanimous-Based Decision Manager ---
  Admin (has ROLE_ADMIN):
    admin -> /resource [hasRole('ADMIN')]: GRANTED
  User (has ROLE_USER only):
    user -> /resource [hasRole('ADMIN')]: DENIED (access denied)
  Manager (has ROLE_MANAGER):
    manager -> /resource [hasRole('ADMIN')]: DENIED (access denied)

=== Example completed successfully ===
```

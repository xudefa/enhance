# enhance AI Agent Rules (Single Source of Truth)

> This file is the sole source for `.cursorrules`, `.windsurfrules`, `.codex`, `.github/copilot-instructions.md`.
> Edit this file then run `make skill-sync` to regenerate. Do not edit platform files directly.

## [Project Overview]
enhance is a Go enterprise framework inspired by Spring Framework/Spring Boot, following idiomatic Go conventions. 526+ Go files, 30+ top-level packages. Three-layer architecture: Boot Layer -> Core Layer -> Infrastructure Layer.

## [Zero External Dependencies]
- Core framework packages (core/boot/context/condition/lifecycle/config/event/web/security/cache/schedule/log/metrics/actuator/tracing/observability/mq/resilience/validation/exception/tenant/email/async/audit/retry/i18n/spel/testing/webtest/openapi/metadata/devtools) must NOT import any third-party packages.
- Only use Go standard library + enhance internal packages.
- Third-party integrations go in starter/<name>/ with independent go.mod files.
- Starter packages must NOT depend on each other; no circular dependencies.
- Examples with third-party dependencies go in examples/<name>/.
- Core framework examples (no third-party) go in cmd/demo/.

## [doc.go Facade]
- All public interfaces, types, and aliases are defined in doc.go.
- doc.go must NOT contain any implementation logic.
- Constructor functions must return interfaces, not concrete struct types.
- Interface names use er suffix (e.g., Reader, Logger, Cache), never I prefix.
- Interface methods should be minimal (1-5 methods).

## [Dependency Direction]
- Boot Layer -> Core Layer -> Infrastructure Layer (one-way).
- No circular dependencies allowed between packages.
- Starter packages must NOT depend on each other.

## [Naming]
- Package names: lowercase, no underscores or mixed case.
- Exported identifiers: CamelCase.
- Error variables: Err prefix (e.g., ErrNotFound).
- Interface names: er suffix (e.g., Reader, Writer, Cache).
- Never use I prefix for interfaces.

## [Testing]
- Use table-driven tests with t.Parallel().
- Both test functions and sub-tests must use t.Parallel().
- Test coverage must be >= 80%.
- Cover both normal paths and error paths.

## [Code Style]
- Use early returns; avoid unnecessary else branches.
- Use functional options pattern for configuration.
- context.Context as first parameter.
- Error wrapping with fmt.Errorf("...: %w", err).
- Use errors.Is/As for error comparison, never direct comparison.
- No naked goroutines; use errgroup or WaitGroup.
- Files: production <= 500 lines, tests <= 1000 lines.
- Functions: <= 80 lines (excluding comments and blank lines).
- Algorithm nesting depth: <= 17 levels.
- Algorithm complexity: time and space cannot both be >= O(n^2).

## [Documentation]
- Read docs/AI_INDEX.md first for navigation.
- Read AGENTS.md for global rules.
- Read docs/AI_TASK_HANDBOOK.md for SOPs.
- Read docs/AI_PITFALLS.md for common pitfalls.
- Read ARCHITECTURE.md and CODING_STYLE.md for design details.

## [Skill]
- This project uses opencode skill: .opencode/skills/enhance-dev/SKILL.md.
- AI should load that skill for the complete iteration workflow.

## [Impact Analysis]
- Changing core/boot/config/condition/log affects many downstream packages; always read docs/AI_CHANGE_IMPACT.md first.
- core affects: actuator, boot, config, context, schedule, security, testing, tracing, web.
- boot affects: actuator, metrics, schedule, security, testing, tracing, web, cmd.
- config affects: actuator, boot, condition, context, schedule, security, testing, tracing.
- condition affects: actuator, boot, metrics, schedule, security, tracing, web, cmd.
- Prefer adding new interface methods over changing existing signatures.
- When modifying interfaces, update docs/API_CHANGELOG.md.

## [Common Pitfalls]
- Never define interfaces in implementation files; use doc.go (ADR-005).
- Never return concrete types from constructors; return interfaces (ADR-006).
- Never import third-party packages in core packages; use starter/ (ADR-004).
- Never compare errors with ==; use errors.Is/As.
- Never let starter packages depend on each other.
- Never forget t.Parallel() in tests (both test functions and sub-tests).
- Event routing uses Type() string; typos cause silent routing failure (ADR-009).
- Cache.Get miss returns ErrNotFound; do not use zero value to check.
- Cron expressions are 6-field Spring style: sec min hour day month weekday.
- Functions must not exceed 80 lines (excluding comments/blank lines); extract sub-functions.
- Algorithm nesting depth must not exceed 17 levels; use early returns/guard clauses.
- Time and space complexity cannot both be >= O(n^2); optimize or trade off.

## [Auto-Config Priority]
- OrderPriorityInfrastructure: -3000 (log, config, condition)
- OrderPriorityDataLayer: -2000 (cache, gorm, redis)
- OrderPriorityServiceDiscovery: -1800 (consul, nacos)
- OrderPriorityAuthentication: -1500 (jwt, oauth2)
- OrderPriorityAuthorization: -1200 (casbin, rbac)
- OrderPriorityTaskLayer: -500 (cron, asynq)
- OrderPrioritySecurityCore: -100 (security)
- OrderPriorityWebLayer: 0 (web)
- OrderPriorityBusinessLayer: 1000 (event, schedule)
- OrderPriorityMonitoring: 2000 (metrics, actuator)

## [Verification]
Before any code change, run:
1. make ai-verify (compile + test + format + static analysis + audit + docs sync + dependency direction)
2. make quality-score (score >= 80)
3. go build ./...
4. go test ./...
5. go test -race <affected-packages>
6. go fmt ./...
7. go vet ./...
# Changelog

All notable changes to the enhance project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- AI tool configuration files (.cursorrules, .windsurfrules, .github/copilot-instructions.md, .codex)
- Code template library (docs/templates/)
- AI maintainability checks (Makefile targets)
- Testing guidelines (docs/TESTING_GUIDE.md)
- Document synchronization mechanism
- Package dependency visualization tools
- API changelog system

### Changed
- Enhanced Makefile with AI maintainability checks
- Refactored `Plugin` interface to compose optional `PluginMeta` sub-interface
- Refactored `CasbinEnforcer` interface to compose `PolicyManager` sub-interface
- Refactored `Connection` interface to compose `AttributeStore` and `RoomParticipant` sub-interfaces

### Deprecated
- Nothing

### Removed
- Nothing

### Fixed
- Nothing

### Security
- Nothing

## [0.0.6] - 2026-09-16

### Added
- IoC container with generic API
- Auto-configuration with conditions
- Starter mechanism for third-party integrations
- Web framework with middleware support
- Security framework with authentication/authorization
- Cache abstraction with LRU implementation
- Scheduled task support
- Logging abstraction
- Metrics collection
- Actuator endpoints
- Distributed tracing

### Changed
- Improved doc.go facade pattern
- Enhanced error handling
- Better test coverage

### Fixed
- Dependency direction issues
- Circular dependency prevention

## [0.0.5] - 2026-09-01

### Added
- Initial framework structure
- Basic IoC container
- Configuration management
- Event-driven architecture

## [0.0.1] - 2026-08-01

### Added
- Project initialization
- Basic documentation

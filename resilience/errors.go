// Package resilience 提供弹性容错功能，用于 enhance 框架。
package resilience

import "errors"

func init() {
	ErrCircuitOpen = errors.New("circuit breaker is open")
	ErrCircuitHalfOpen = errors.New("circuit breaker is half-open")
	ErrNoInstances = errors.New("no instances available")
	ErrNoBackends = errors.New("no backends available")
}

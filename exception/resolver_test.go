package exception

import (
	"context"
	"errors"
	"testing"
)

func TestResolverChain_AddResolver(t *testing.T) {
	t.Parallel()
	chain := NewResolverChain()

	resolver1 := &mockResolver{order: 2}
	resolver2 := &mockResolver{order: 1}

	chain.AddResolver(resolver1)
	chain.AddResolver(resolver2)

	resolvers := chain.GetResolvers()
	if len(resolvers) != 2 {
		t.Errorf("Expected 2 resolvers, got %d", len(resolvers))
	}

	if resolvers[0].Order() != 1 {
		t.Errorf("Expected first resolver order 1, got %d", resolvers[0].Order())
	}
	if resolvers[1].Order() != 2 {
		t.Errorf("Expected second resolver order 2, got %d", resolvers[1].Order())
	}
}

func TestResolverChain_Resolve(t *testing.T) {
	t.Parallel()
	chain := NewResolverChain()

	resolver1 := &mockResolver{
		order:    1,
		supports: func(err error) bool { return false },
	}
	resolver2 := &mockResolver{
		order:    2,
		supports: func(err error) bool { return errors.Is(err, ErrNotFound) },
		resolve: func(ctx context.Context, err error) *ErrorResponse {
			return NewErrorResponse(404, "Custom not found", "", "", nil)
		},
	}

	chain.AddResolver(resolver1)
	chain.AddResolver(resolver2)

	resp := chain.Resolve(context.Background(), ErrNotFound)
	if resp == nil || resp.Code != 404 || resp.Message != "Custom not found" {
		t.Fatalf("Expected code 404 and message 'Custom not found', got code=%v, message=%v", resp != nil, resp != nil)
	}
}

type mockResolver struct {
	order    int
	supports func(err error) bool
	resolve  func(ctx context.Context, err error) *ErrorResponse
}

func (m *mockResolver) Resolve(ctx context.Context, err error) *ErrorResponse {
	if m.resolve != nil {
		return m.resolve(ctx, err)
	}
	return nil
}

func (m *mockResolver) Supports(err error) bool {
	if m.supports != nil {
		return m.supports(err)
	}
	return false
}

func (m *mockResolver) Order() int {
	return m.order
}

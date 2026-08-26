package testing

import (
	"testing"
)

func TestTestRunner_Run_WithPropertiesAndMockBeans(t *testing.T) {
	t.Parallel()
	mock := &mockService{}

	runner := NewTestRunner(t,
		WithProperty("feature.enabled", true),
		WithMockBean("mockService", mock),
	)

	runner.Run(func(ctx TestContext) {
		if runner.GetContext() == nil {
			t.Fatal("expected context to be initialized after Run")
		}
		got, err := GetByType[*mockService](ctx)
		if err != nil {
			t.Fatalf("expected mock bean to be registered: %v", err)
		}
		if got != mock {
			t.Error("expected retrieved bean to be the injected mock instance")
		}
	})
}

func TestTestRunner_RegisterMockBeans_EarlyReturns(t *testing.T) {
	t.Parallel()

	t.Run("nil context returns early", func(t *testing.T) {
		t.Parallel()
		runner := NewTestRunner(t, WithMockBean("mockService", &mockService{}))
		runner.registerMockBeans()
	})

	t.Run("non-container backend returns early", func(t *testing.T) {
		t.Parallel()
		runner := NewTestRunner(t, WithMockBean("mockService", &mockService{}))
		ctx := NewTestContext(t)
		impl := ctx.(*testContextImpl)
		impl.container = nil
		runner.context = ctx

		runner.registerMockBeans()
	})

	t.Run("registers into fresh context", func(t *testing.T) {
		t.Parallel()
		mock := &mockService{}
		runner := NewTestRunner(t, WithMockBean("mockService", mock))
		runner.context = NewTestContext(t)

		runner.registerMockBeans()

		got, err := GetByType[*mockService](runner.context)
		if err != nil {
			t.Fatalf("expected mock bean registered: %v", err)
		}
		if got != mock {
			t.Error("expected retrieved bean to be the injected mock instance")
		}
	})
}

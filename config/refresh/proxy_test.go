package refresh

import (
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
)

func TestRefreshProxy_GetTarget_NoRefresh(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	scopeManager := &RefreshScopeManager{
		config:           DefaultRefreshConfig(),
		refreshFlags:     make(map[string]bool),
		beanVersions:     make(map[string]*atomic.Int64),
		activeProxies:    make(map[string]*RefreshProxy),
		refreshableBeans: make(map[string]RefreshableBean),
		metrics:          NewRefreshMetrics(),
		logger:           logger,
	}

	initial := "initial"
	proxy := NewRefreshProxy("testBean", initial, scopeManager, logger)

	got := proxy.GetTarget()
	if got != initial {
		t.Errorf("GetTarget() = %v, want %v", got, initial)
	}
}

func TestRefreshProxy_GetTarget_WithManager(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	creator := &simpleBeanCreator{value: "new-value"}
	scopeManager := NewRefreshScopeManager(creator, logger)

	proxy := NewRefreshProxy("testBean", "old-value", scopeManager, logger)

	got := proxy.GetTarget()
	if got != "old-value" {
		t.Errorf("expected old-value, got %v", got)
	}

	proxy.MarkForRefresh()
	got = proxy.GetTarget()
	if got != "new-value" {
		t.Errorf("expected new-value after refresh, got %v", got)
	}
}

func TestRefreshProxy_GetTarget_RefreshFailureFallsBack(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	creator := &failingBeanCreator{}
	scopeManager := NewRefreshScopeManager(creator, logger)

	proxy := NewRefreshProxy("testBean", "fallback", scopeManager, logger)
	proxy.MarkForRefresh()

	got := proxy.GetTarget()
	if got != "fallback" {
		t.Errorf("expected fallback value, got %v", got)
	}
}

func TestRefreshProxy_DoubleCheckLocking(t *testing.T) {
	t.Parallel()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	creator := &simpleBeanCreator{value: "refreshed"}
	scopeManager := NewRefreshScopeManager(creator, logger)

	proxy := NewRefreshProxy("testBean", "initial", scopeManager, logger)
	proxy.MarkForRefresh()

	got := proxy.GetTarget()
	if got != "refreshed" {
		t.Errorf("expected refreshed, got %v", got)
	}

	if proxy.needsRefresh.Load() {
		t.Error("needsRefresh should be false after successful refresh")
	}
}

type simpleBeanCreator struct {
	value any
}

func (c *simpleBeanCreator) CreateBean(beanID string) (any, error) {
	return c.value, nil
}

type failingBeanCreator struct{}

func (c *failingBeanCreator) CreateBean(beanID string) (any, error) {
	return nil, &testError{"creation failed"}
}

type testError struct {
	msg string
}

func (e *testError) Error() string { return e.msg }

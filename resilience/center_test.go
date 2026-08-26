package resilience

import (
	"context"
	"testing"
)

func TestNewInMemoryRegistry(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	if reg == nil {
		t.Fatal("expected non-nil registry")
	}
}

func TestInMemoryRegistry_Register(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	info := InstanceInfo{
		ID:          "inst1",
		ServiceName: "test-service",
		Host:        "localhost",
		Port:        8080,
	}

	err := reg.Register(ctx, info)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
}

func TestInMemoryRegistry_Deregister(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	info := InstanceInfo{
		ID:          "inst1",
		ServiceName: "test-service",
		Host:        "localhost",
		Port:        8080,
	}

	_ = reg.Register(ctx, info)

	err := reg.Deregister(ctx, info)
	if err != nil {
		t.Fatalf("Deregister failed: %v", err)
	}

	instances, err := reg.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected 0 instances after deregister, got %d", len(instances))
	}
}

func TestInMemoryRegistry_Deregister_NotFound(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	info := InstanceInfo{
		ID:          "inst1",
		ServiceName: "test-service",
	}

	err := reg.Deregister(ctx, info)
	if err != nil {
		t.Fatalf("Deregister for non-existent instance failed: %v", err)
	}
}

func TestInMemoryRegistry_Discover(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	info1 := InstanceInfo{
		ID:          "inst1",
		ServiceName: "test-service",
		Host:        "localhost",
		Port:        8080,
	}
	info2 := InstanceInfo{
		ID:          "inst2",
		ServiceName: "test-service",
		Host:        "localhost",
		Port:        8081,
	}

	_ = reg.Register(ctx, info1)
	_ = reg.Register(ctx, info2)

	instances, err := reg.Discover(ctx, "test-service")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestInMemoryRegistry_Discover_NotFound(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	instances, err := reg.Discover(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}
	if len(instances) != 0 {
		t.Errorf("expected 0 instances, got %d", len(instances))
	}
}

func TestInMemoryRegistry_Watch(t *testing.T) {
	t.Parallel()
	reg := NewInMemoryRegistry()
	ctx := context.Background()

	info := InstanceInfo{
		ID:          "inst1",
		ServiceName: "test-service",
		Host:        "localhost",
		Port:        8080,
	}
	_ = reg.Register(ctx, info)

	ch, err := reg.Watch(ctx, "test-service")
	if err != nil {
		t.Fatalf("Watch failed: %v", err)
	}

	select {
	case instances := <-ch:
		if len(instances) != 1 {
			t.Errorf("expected 1 instance, got %d", len(instances))
		}
	default:
		t.Error("expected to receive initial instance list")
	}
}

package context

import (
	"reflect"
	"testing"

	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/core"
	"github.com/xudefa/enhance/core/registry"
)

func TestCreateApplicationContext(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	env := environment.NewEnvironment()

	ctx, err := CreateApplicationContext(
		WithContainer(container),
		WithEnvironment(env),
	)

	if err != nil {
		t.Fatalf("CreateApplicationContext should succeed: %v", err)
	}
	if ctx.Container() != container {
		t.Error("container not set correctly")
	}
	if ctx.Environment() != env {
		t.Error("environment not set correctly")
	}
}

func TestWithContainer(t *testing.T) {
	t.Parallel()
	container := core.NewContainer()
	opt := WithContainer(container)

	builder := NewApplicationContextBuilder()
	opt(builder)

	if builder.container != container {
		t.Error("WithContainer should set container")
	}
}

func TestWithEnvironment(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	opt := WithEnvironment(env)

	builder := NewApplicationContextBuilder()
	opt(builder)

	if builder.env != env {
		t.Error("WithEnvironment should set environment")
	}
}

func TestWithBean(t *testing.T) {
	t.Parallel()
	type testBean struct{}
	tt := reflect.TypeFor[testBean]()
	opt := WithBean(tt, core.WithScope[any](registry.Singleton))

	builder := NewApplicationContextBuilder()
	opt(builder)

	if len(builder.beans[tt]) != 1 {
		t.Errorf("WithBean should add bean option, got %d", len(builder.beans[tt]))
	}
}

func TestProfile(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()
	builder.Profile("dev")

	if len(builder.profiles) != 1 || builder.profiles[0] != "dev" {
		t.Errorf("Profile should add profile, got %v", builder.profiles)
	}
}

func TestProfiles(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()
	builder.Profiles("dev", "test")

	if len(builder.profiles) != 2 {
		t.Errorf("Profiles should add 2 profiles, got %d", len(builder.profiles))
	}
}

func TestWithProfile(t *testing.T) {
	t.Parallel()
	opt := WithProfile("dev")

	builder := NewApplicationContextBuilder()
	opt(builder)

	if len(builder.profiles) != 1 || builder.profiles[0] != "dev" {
		t.Errorf("WithProfile should add profile, got %v", builder.profiles)
	}
}

func TestWithProfiles(t *testing.T) {
	t.Parallel()
	opt := WithProfiles("dev", "test")

	builder := NewApplicationContextBuilder()
	opt(builder)

	if len(builder.profiles) != 2 {
		t.Errorf("WithProfiles should add 2 profiles, got %d", len(builder.profiles))
	}
}

func TestBuilder_Build_WithProfiles(t *testing.T) {
	t.Parallel()
	builder := NewApplicationContextBuilder()
	builder.Profiles("dev", "test")

	ctx, err := builder.Build()

	if err != nil {
		t.Fatalf("Build should succeed: %v", err)
	}

	profiles := ctx.Environment().GetActiveProfiles()
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}

	if profiles[0] != "dev" || profiles[1] != "test" {
		t.Errorf("expected profiles [dev, test], got %v", profiles)
	}
}

package casbinxorm

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/casbin/casbin/v2"
)

const testModelText = `
[request_definition]
r = sub, obj, act

[policy_definition]
p = sub, obj, act

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
`

func writeModelFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	modelPath := filepath.Join(dir, "model.conf")
	if err := os.WriteFile(modelPath, []byte(strings.TrimSpace(content)), 0644); err != nil {
		t.Fatalf("failed to write model file: %v", err)
	}
	return modelPath
}

func writePolicyFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.csv")
	if err := os.WriteFile(policyPath, []byte(strings.TrimSpace(content)), 0644); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}
	return policyPath
}

func newTestEnforcer(t *testing.T) *XormCasbinEnforcer {
	t.Helper()
	modelPath := writeModelFile(t, testModelText)
	policyPath := writePolicyFile(t, "")
	e, err := casbin.NewEnforcer(modelPath, policyPath)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}
	return &XormCasbinEnforcer{
		Enforcer: e,
		adapter:  nil,
	}
}

func TestXormCasbinEnforcer_Enforce(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	ctx := context.Background()

	_, err := e.Enforcer.AddPolicy("admin", "/data1", "read")
	if err != nil {
		t.Fatalf("AddPolicy error = %v", err)
	}
	_, err = e.Enforcer.AddRoleForUser("alice", "admin")
	if err != nil {
		t.Fatalf("AddRoleForUser error = %v", err)
	}

	tests := []struct {
		name    string
		sub     string
		obj     string
		act     string
		want    bool
		wantErr bool
	}{
		{"admin read allowed", "alice", "/data1", "read", true, false},
		{"admin read denied", "alice", "/data2", "read", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := e.Enforce(ctx, tt.sub, tt.obj, tt.act)
			if (err != nil) != tt.wantErr {
				t.Errorf("Enforce() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Enforce(%s, %s, %s) = %v, want %v", tt.sub, tt.obj, tt.act, got, tt.want)
			}
		})
	}
}

func TestXormCasbinEnforcer_AddPolicy_RemovePolicy(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	ctx := context.Background()

	if err := e.AddPolicy(ctx, "user", "/profile", "read"); err != nil {
		t.Fatalf("AddPolicy() error = %v", err)
	}

	allowed, err := e.Enforce(ctx, "user", "/profile", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if !allowed {
		t.Error("expected user to have read access to /profile")
	}

	if err := e.RemovePolicy(ctx, "user", "/profile", "read"); err != nil {
		t.Fatalf("RemovePolicy() error = %v", err)
	}

	allowed, err = e.Enforce(ctx, "user", "/profile", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if allowed {
		t.Error("expected user to not have read access to /profile after removing")
	}
}

func TestXormCasbinEnforcer_GetPolicy(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	ctx := context.Background()

	_, err := e.Enforcer.AddPolicy("guest", "/public", "read")
	if err != nil {
		t.Fatalf("AddPolicy error = %v", err)
	}

	policies, err := e.GetPolicy(ctx)
	if err != nil {
		t.Fatalf("GetPolicy() error = %v", err)
	}
	if len(policies) == 0 {
		t.Error("expected at least one policy")
	}
}

func TestXormCasbinEnforcer_LoadPolicy(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	ctx := context.Background()

	if err := e.LoadPolicy(ctx); err != nil {
		t.Fatalf("LoadPolicy() error = %v", err)
	}
}

func TestXormCasbinEnforcer_SavePolicy(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	ctx := context.Background()

	_, err := e.Enforcer.AddPolicy("manager", "/finance", "read")
	if err != nil {
		t.Fatalf("AddPolicy error = %v", err)
	}

	if err := e.SavePolicy(ctx); err != nil {
		t.Fatalf("SavePolicy() error = %v", err)
	}
}

func TestXormCasbinEnforcer_InterfaceCompliance(t *testing.T) {
	t.Parallel()
	e := newTestEnforcer(t)
	var _ interface {
		Enforce(ctx context.Context, subject, object, action string) (bool, error)
		AddPolicy(ctx context.Context, sub, obj, act string) error
		RemovePolicy(ctx context.Context, sub, obj, act string) error
		GetPolicy(ctx context.Context) ([][]string, error)
		LoadPolicy(ctx context.Context) error
		SavePolicy(ctx context.Context) error
	} = e
}

package casbin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xudefa/enhance/log"
)

func TestGetSubject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule []string
		want string
	}{
		{"normal rule", []string{"alice", "/data1", "read"}, "alice"},
		{"single element", []string{"bob"}, "bob"},
		{"empty rule", []string{}, ""},
		{"nil rule", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetSubject(tt.rule)
			if got != tt.want {
				t.Errorf("GetSubject(%v) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

func TestGetObject(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule []string
		want string
	}{
		{"normal rule", []string{"alice", "/data1", "read"}, "/data1"},
		{"one element", []string{"alice"}, ""},
		{"two elements", []string{"alice", "/data1"}, "/data1"},
		{"empty rule", []string{}, ""},
		{"nil rule", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetObject(tt.rule)
			if got != tt.want {
				t.Errorf("GetObject(%v) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

func TestGetAction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule []string
		want string
	}{
		{"normal rule", []string{"alice", "/data1", "read"}, "read"},
		{"two elements", []string{"alice", "/data1"}, ""},
		{"empty rule", []string{}, ""},
		{"nil rule", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := GetAction(tt.rule)
			if got != tt.want {
				t.Errorf("GetAction(%v) = %q, want %q", tt.rule, got, tt.want)
			}
		})
	}
}

func TestIsRoleRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule []string
		want bool
	}{
		{"group rule", []string{"g", "alice", "admin"}, true},
		{"policy rule", []string{"p", "/data", "read"}, false},
		{"empty", []string{}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsRoleRule(tt.rule)
			if got != tt.want {
				t.Errorf("IsRoleRule(%v) = %v, want %v", tt.rule, got, tt.want)
			}
		})
	}
}

func TestIsPolicyRule(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		rule []string
		want bool
	}{
		{"policy rule", []string{"p", "/data", "read"}, true},
		{"group rule", []string{"g", "alice", "admin"}, false},
		{"empty", []string{}, false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := IsPolicyRule(tt.rule)
			if got != tt.want {
				t.Errorf("IsPolicyRule(%v) = %v, want %v", tt.rule, got, tt.want)
			}
		})
	}
}

func newTestLogger(t *testing.T) log.Logger {
	t.Helper()
	return log.Build()
}

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

const testPolicyText = `
p, admin, /data1, read
p, admin, /data1, write
p, user, /data2, read
g, alice, admin
g, bob, user
`

func writeTempFiles(t *testing.T, modelText, policyText string) (modelPath, policyPath string) {
	t.Helper()

	dir := t.TempDir()
	modelPath = filepath.Join(dir, "model.conf")
	policyPath = filepath.Join(dir, "policy.csv")

	if err := os.WriteFile(modelPath, []byte(strings.TrimSpace(modelText)), 0644); err != nil {
		t.Fatalf("failed to write model file: %v", err)
	}
	if err := os.WriteFile(policyPath, []byte(strings.TrimSpace(policyText)), 0644); err != nil {
		t.Fatalf("failed to write policy file: %v", err)
	}
	return modelPath, policyPath
}

func TestNewCasbinEnforcer_Success(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}
	if enforcer == nil {
		t.Fatal("NewCasbinEnforcer() returned nil enforcer")
	}
	if enforcer.enforcer == nil {
		t.Fatal("inner enforcer is nil")
	}
}

func TestNewCasbinEnforcer_InvalidModelPath(t *testing.T) {
	t.Parallel()
	_, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), "/nonexistent/model.conf", "/nonexistent/policy.csv")
	if err == nil {
		t.Error("expected error for invalid model path, got nil")
	}
}

func TestDefaultCasbinEnforcer_Enforce(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	tests := []struct {
		name    string
		sub     string
		obj     string
		act     string
		want    bool
		wantErr bool
	}{
		{"admin read /data1", "alice", "/data1", "read", true, false},
		{"admin write /data1", "alice", "/data1", "write", true, false},
		{"user read /data2", "bob", "/data2", "read", true, false},
		{"admin read /data2 denied", "alice", "/data2", "read", false, false},
		{"user write /data1 denied", "bob", "/data1", "write", false, false},
		{"unknown user", "unknown", "/data1", "read", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := enforcer.Enforce(context.Background(), tt.sub, tt.obj, tt.act)
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

func TestDefaultCasbinEnforcer_AddPolicy_RemovePolicy(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	if err := enforcer.AddPolicy(ctx, "guest", "/public", "read"); err != nil {
		t.Fatalf("AddPolicy() error = %v", err)
	}

	allowed, err := enforcer.Enforce(ctx, "guest", "/public", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if !allowed {
		t.Error("expected guest to have read access to /public after adding policy")
	}

	if err := enforcer.RemovePolicy(ctx, "guest", "/public", "read"); err != nil {
		t.Fatalf("RemovePolicy() error = %v", err)
	}

	allowed, err = enforcer.Enforce(ctx, "guest", "/public", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if allowed {
		t.Error("expected guest to not have read access to /public after removing policy")
	}
}

func TestDefaultCasbinEnforcer_GetPolicy(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	policies, err := enforcer.GetPolicy(context.Background())
	if err != nil {
		t.Fatalf("GetPolicy() error = %v", err)
	}
	if len(policies) == 0 {
		t.Error("expected at least one policy, got none")
	}
}

func TestDefaultCasbinEnforcer_RoleMethods(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	roles, err := enforcer.GetRolesForUser(ctx, "alice")
	if err != nil {
		t.Fatalf("GetRolesForUser() error = %v", err)
	}
	if len(roles) != 1 || roles[0] != "admin" {
		t.Errorf("expected [admin], got %v", roles)
	}

	has, err := enforcer.HasRoleForUser(ctx, "alice", "admin")
	if err != nil {
		t.Fatalf("HasRoleForUser() error = %v", err)
	}
	if !has {
		t.Error("expected alice to have admin role")
	}

	users, err := enforcer.GetUsersForRole(ctx, "admin")
	if err != nil {
		t.Fatalf("GetUsersForRole() error = %v", err)
	}
	if len(users) == 0 {
		t.Error("expected at least one user for admin role")
	}

	ok, err := enforcer.AddRoleForUser(ctx, "charlie", "user")
	if err != nil {
		t.Fatalf("AddRoleForUser() error = %v", err)
	}
	if !ok {
		t.Error("expected true from AddRoleForUser")
	}

	has, err = enforcer.HasRoleForUser(ctx, "charlie", "user")
	if err != nil {
		t.Fatalf("HasRoleForUser() error = %v", err)
	}
	if !has {
		t.Error("expected charlie to have user role")
	}

	ok, err = enforcer.DeleteRoleForUser(ctx, "alice", "admin")
	if err != nil {
		t.Fatalf("DeleteRoleForUser() error = %v", err)
	}
	if !ok {
		t.Error("expected true from DeleteRoleForUser")
	}

	roles, err = enforcer.GetRolesForUser(ctx, "alice")
	if err != nil {
		t.Fatalf("GetRolesForUser() error = %v", err)
	}
	if len(roles) != 0 {
		t.Errorf("expected no roles for alice, got %v", roles)
	}
}

func TestDefaultCasbinEnforcer_DeleteUser_DeleteRole(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	ok, err := enforcer.DeleteUser(ctx, "bob")
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}
	if !ok {
		t.Error("expected true from DeleteUser")
	}

	ok, err = enforcer.DeleteRole(ctx, "user")
	if err != nil {
		t.Fatalf("DeleteRole() error = %v", err)
	}
	if !ok {
		t.Error("expected true from DeleteRole")
	}
}

func TestDefaultCasbinEnforcer_AddPolicies_RemovePolicies(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	newPolicies := [][]string{
		{"manager", "/finance", "read"},
		{"manager", "/finance", "write"},
	}

	ok, err := enforcer.AddPolicies(ctx, newPolicies)
	if err != nil {
		t.Fatalf("AddPolicies() error = %v", err)
	}
	if !ok {
		t.Error("expected true from AddPolicies")
	}

	allowed, err := enforcer.Enforce(ctx, "manager", "/finance", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if !allowed {
		t.Error("expected manager to have read access to /finance")
	}

	ok, err = enforcer.RemovePolicies(ctx, newPolicies)
	if err != nil {
		t.Fatalf("RemovePolicies() error = %v", err)
	}
	if !ok {
		t.Error("expected true from RemovePolicies")
	}
}

func TestDefaultCasbinEnforcer_PermissionsForUser(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	perms := enforcer.GetPermissionsForUser(ctx, "admin")
	if len(perms) == 0 {
		t.Error("expected at least one permission for admin")
	}

	has := enforcer.HasPermissionForUser(ctx, "admin", "/data1", "read")
	if !has {
		t.Error("expected admin to have read permission on /data1")
	}

	has = enforcer.HasPermissionForUser(ctx, "admin", "/data99", "read")
	if has {
		t.Error("expected admin to not have read permission on /data99")
	}
}

func TestDefaultCasbinEnforcer_ClearPolicy_EnableSettings(t *testing.T) {
	t.Parallel()
	modelPath, policyPath := writeTempFiles(t, testModelText, testPolicyText)

	enforcer, err := NewCasbinEnforcer(context.Background(), newTestLogger(t), modelPath, policyPath)
	if err != nil {
		t.Fatalf("NewCasbinEnforcer() error = %v", err)
	}

	ctx := context.Background()

	enforcer.EnableAutoSave(true)
	enforcer.EnableAutoSave(false)
	enforcer.EnableEnforce(true)
	enforcer.EnableEnforce(false)
	enforcer.EnableEnforce(true)

	if err := enforcer.ClearPolicy(ctx); err != nil {
		t.Fatalf("ClearPolicy() error = %v", err)
	}

	allowed, err := enforcer.Enforce(ctx, "alice", "/data1", "read")
	if err != nil {
		t.Fatalf("Enforce() error = %v", err)
	}
	if allowed {
		t.Error("expected no access after clearing policy")
	}
}

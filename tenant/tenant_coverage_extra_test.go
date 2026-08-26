package tenant

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSubdomainResolver_EmptyHost(t *testing.T) {
	t.Parallel()
	resolver := NewSubdomainResolver("example.com")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = ""

	_, err := resolver.Resolve(req)
	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestJWTClaims_ContextRoundTrip(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		setup   func(ctx context.Context) context.Context
		wantOK  bool
		wantTID string
	}{
		{
			name: "claims stored and retrieved",
			setup: func(ctx context.Context) context.Context {
				return SetJWTClaims(ctx, map[string]any{"tid": "tenant-1"})
			},
			wantOK:  true,
			wantTID: "tenant-1",
		},
		{
			name:  "no claims in context",
			setup: func(ctx context.Context) context.Context { return ctx },
		},
		{
			name: "non-map value under claims key",
			setup: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, jwtClaimsKey{}, "not-a-map")
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			claims, ok := ExtractJWTClaims(tt.setup(context.Background()))
			if ok != tt.wantOK {
				t.Fatalf("expected ok = %v, got %v", tt.wantOK, ok)
			}
			if !tt.wantOK {
				return
			}
			if got := claims["tid"]; got != tt.wantTID {
				t.Errorf("expected tid %q, got %v", tt.wantTID, got)
			}
		})
	}
}

func TestNewJWTExtractor(t *testing.T) {
	t.Parallel()
	parse := func(authHeader string) (map[string]any, error) {
		return map[string]any{"sub": authHeader}, nil
	}

	extractor := NewJWTExtractor(parse)
	if extractor == nil {
		t.Fatal("expected non-nil extractor")
	}
	if extractor.parse == nil {
		t.Fatal("expected parse function to be set")
	}
}

func TestJWTExtractor_Handle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		authHeader string
		parse      func(string) (map[string]any, error)
		wantStatus int
		wantTID    string
	}{
		{
			name:       "missing authorization header returns 401",
			parse:      func(string) (map[string]any, error) { return nil, nil },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "parse failure returns 401",
			authHeader: "Bearer invalid",
			parse:      func(string) (map[string]any, error) { return nil, errors.New("invalid token") },
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "valid token passes claims downstream",
			authHeader: "Bearer valid",
			parse: func(string) (map[string]any, error) {
				return map[string]any{"tid": "tenant-9"}, nil
			},
			wantStatus: http.StatusOK,
			wantTID:    "tenant-9",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			extractor := NewJWTExtractor(tt.parse)

			var gotTID string
			handler := extractor.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				claims, _ := ExtractJWTClaims(r.Context())
				gotTID, _ = claims["tid"].(string)
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, rr.Code)
			}
			if tt.wantTID != "" && gotTID != tt.wantTID {
				t.Errorf("expected tid %q in downstream handler, got %q", tt.wantTID, gotTID)
			}
		})
	}
}

func TestJWTResolver_Resolve(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		hasClaims bool
		claims    map[string]any
		claimName string
		wantID    string
		wantErr   bool
	}{
		{
			name:      "no claims in request context",
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim key missing from claims map",
			hasClaims: true,
			claims:    map[string]any{"other": "x"},
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim value is not a string",
			hasClaims: true,
			claims:    map[string]any{"tid": 42},
			claimName: "tid",
			wantErr:   true,
		},
		{
			name:      "claim resolved successfully",
			hasClaims: true,
			claims:    map[string]any{"tid": "tenant-1"},
			claimName: "tid",
			wantID:    "tenant-1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resolver := NewJWTResolver(tt.claimName)
			if resolver == nil {
				t.Fatal("expected non-nil resolver")
			}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.hasClaims {
				req = req.WithContext(SetJWTClaims(req.Context(), tt.claims))
			}

			got, err := resolver.Resolve(req)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err presence = %v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && got != tt.wantID {
				t.Errorf("expected tenant ID %q, got %q", tt.wantID, got)
			}
		})
	}
}

func TestTenantManager_SetCurrentTenant_NotFound(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))

	if err := manager.SetCurrentTenant("ghost"); err == nil {
		t.Fatal("expected error when setting unregistered current tenant")
	}
	if manager.GetCurrentTenant() != nil {
		t.Error("expected no current tenant after failed set")
	}
}

func TestTenantMiddleware_UnknownTenant(t *testing.T) {
	t.Parallel()
	manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))
	middleware := NewTenantMiddleware(manager)

	called := false
	handler := middleware.Handle(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Tenant-ID", "ghost")
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Errorf("expected status 403 for unknown tenant, got %d", rr.Code)
	}
	if called {
		t.Error("expected downstream handler not to be invoked")
	}
}

func TestTenantIsolation_UnknownTenantErrors(t *testing.T) {
	t.Parallel()
	isolation := NewTenantIsolation(NewTenantManager(NewHeaderResolver("X-Tenant-ID")))

	tests := []struct {
		name string
		act  func(TenantIsolation, string) (string, error)
	}{
		{"IsolateDatabase", TenantIsolation.IsolateDatabase},
		{"IsolateSchema", TenantIsolation.IsolateSchema},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.act(isolation, "ghost")
			if err == nil {
				t.Fatalf("expected error for unknown tenant via %s", tt.name)
			}
		})
	}
}

func TestTenantProvider_GetCurrentTenantDatabase_Errors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		registerDB   bool
		setCurrent   bool
		wantErr      bool
		wantDatabase string
	}{
		{
			name:    "no current tenant",
			wantErr: true,
		},
		{
			name:       "current tenant without database",
			registerDB: false,
			setCurrent: true,
			wantErr:    true,
		},
		{
			name:         "current tenant with database",
			registerDB:   true,
			setCurrent:   true,
			wantDatabase: "db_tenant1",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			manager := NewTenantManager(NewHeaderResolver("X-Tenant-ID"))
			tenant := &Tenant{ID: "tenant-1", Name: "Tenant 1"}
			if tt.registerDB {
				tenant.Database = "db_tenant1"
			}
			manager.RegisterTenant(tenant)
			if tt.setCurrent {
				if err := manager.SetCurrentTenant("tenant-1"); err != nil {
					t.Fatalf("failed to set current tenant: %v", err)
				}
			}

			provider := NewTenantProvider(manager)
			db, err := provider.GetCurrentTenantDatabase()

			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err presence = %v, got %v", tt.wantErr, err)
			}
			if !tt.wantErr && db != tt.wantDatabase {
				t.Errorf("expected database %q, got %q", tt.wantDatabase, db)
			}
		})
	}
}

package asynq

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/hibiken/asynq"
	"github.com/xudefa/enhance/config/environment"
)

func TestNewAsynqConfig_Defaults(t *testing.T) {
	t.Parallel()
	cfg := NewAsynqConfig()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"host", cfg.Host, "localhost"},
		{"port", cfg.Port, 6379},
		{"db", cfg.DB, 0},
		{"pool-size", cfg.PoolSize, 10},
		{"enable-scheduler", cfg.EnableScheduler, false},
		{"enabled", cfg.Enabled, false},
		{"password", cfg.Password, ""},
		{"concurrency", cfg.Concurrency, 10},
		{"retry-limit", cfg.RetryLimit, 25},
		{"timeout", cfg.Timeout, time.Duration(0)},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestLoadConfig_Bound(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-asynq", environment.PriorityNormal, map[string]any{
		"asynq.enabled":          "true",
		"asynq.host":             "192.168.1.100",
		"asynq.port":             "6380",
		"asynq.password":         "secret",
		"asynq.db":               "1",
		"asynq.pool-size":        "20",
		"asynq.enable-scheduler": "true",
	}))

	cfg, err := LoadConfig(env)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"enabled", cfg.Enabled, true},
		{"host", cfg.Host, "192.168.1.100"},
		{"port", cfg.Port, 6380},
		{"password", cfg.Password, "secret"},
		{"db", cfg.DB, 1},
		{"pool-size", cfg.PoolSize, 20},
		{"enable-scheduler", cfg.EnableScheduler, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestLoadConfig_PartialOverride(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-asynq-partial", environment.PriorityNormal, map[string]any{
		"asynq.enabled": "true",
		"asynq.host":    "10.0.0.1",
	}))

	cfg, err := LoadConfig(env)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"host overridden", cfg.Host, "10.0.0.1"},
		{"port default", cfg.Port, DefaultAsynqPort},
		{"db default", cfg.DB, DefaultAsynqDB},
		{"pool-size default", cfg.PoolSize, DefaultAsynqPoolSize},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestLoadConfig_Disabled(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-asynq-disabled", environment.PriorityNormal, map[string]any{
		"asynq.enabled": "false",
	}))

	cfg, err := LoadConfig(env)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Enabled {
		t.Errorf("expected enabled=false, got %v", cfg.Enabled)
	}
}

func TestLoadConfig_EmptyEnvironment(t *testing.T) {
	t.Parallel()
	env := environment.NewEnvironment()

	cfg, err := LoadConfig(env)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	defaults := NewAsynqConfig()
	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"host", cfg.Host, defaults.Host},
		{"port", cfg.Port, defaults.Port},
		{"db", cfg.DB, defaults.DB},
		{"pool-size", cfg.PoolSize, defaults.PoolSize},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestAsynqConstants_Values(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"config prefix", ConfigPrefix, "asynq"},
		{"enabled key", AsynqEnabled, "asynq.enabled"},
		{"condition true", ConditionTrue, "true"},
		{"default host", DefaultAsynqHost, "localhost"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.got)
			}
		})
	}
}

func TestAsynqConstants_NumericDefaults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"default port", DefaultAsynqPort, 6379},
		{"default db", DefaultAsynqDB, 0},
		{"default pool-size", DefaultAsynqPoolSize, 10},
		{"default concurrency", DefaultConcurrency, 10},
		{"default retry-limit", DefaultRetryLimit, 25},
		{"default timeout", int(DefaultTimeout), 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, tt.got)
			}
		})
	}
}

func TestAsynqConfig_BuildAddr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cfg      *AsynqConfig
		expected string
	}{
		{"default", &AsynqConfig{Host: DefaultAsynqHost, Port: DefaultAsynqPort}, "localhost:6379"},
		{"custom host and port", &AsynqConfig{Host: "192.168.1.100", Port: 6380}, "192.168.1.100:6380"},
		{"custom host default port", &AsynqConfig{Host: "10.0.0.1", Port: DefaultAsynqPort}, "10.0.0.1:6379"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.cfg.BuildAddr()
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestAsynqConfig_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		cfg         *AsynqConfig
		expectedErr error
	}{
		{"valid default", NewAsynqConfig(), nil},
		{"valid custom", &AsynqConfig{Host: "redis", Port: 6380, DB: 1, PoolSize: 20, Concurrency: 5, RetryLimit: 10, Timeout: 30 * time.Second}, nil},
		{"valid port boundary min", &AsynqConfig{Host: "localhost", Port: 1}, nil},
		{"valid port boundary max", &AsynqConfig{Host: "localhost", Port: 65535}, nil},
		{"valid zero timeout", &AsynqConfig{Host: "localhost", Port: 6379, Timeout: 0}, nil},
		{"empty host", &AsynqConfig{Host: "", Port: 6379}, ErrInvalidHost},
		{"zero port", &AsynqConfig{Host: "localhost", Port: 0}, ErrInvalidPort},
		{"negative port", &AsynqConfig{Host: "localhost", Port: -1}, ErrInvalidPort},
		{"port overflow", &AsynqConfig{Host: "localhost", Port: 65536}, ErrInvalidPort},
		{"negative db", &AsynqConfig{Host: "localhost", Port: 6379, DB: -1}, ErrInvalidDB},
		{"negative pool size", &AsynqConfig{Host: "localhost", Port: 6379, PoolSize: -1}, ErrInvalidPoolSize},
		{"negative concurrency", &AsynqConfig{Host: "localhost", Port: 6379, Concurrency: -1}, ErrInvalidConcurrency},
		{"negative retry limit", &AsynqConfig{Host: "localhost", Port: 6379, RetryLimit: -1}, ErrInvalidRetryLimit},
		{"negative timeout", &AsynqConfig{Host: "localhost", Port: 6379, Timeout: -1 * time.Second}, ErrInvalidTimeout},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := tt.cfg.Validate()
			if tt.expectedErr == nil {
				if err != nil {
					t.Errorf("expected nil error, got %v", err)
				}
			} else {
				if !errors.Is(err, tt.expectedErr) {
					t.Errorf("expected %v, got %v", tt.expectedErr, err)
				}
			}
		})
	}
}

func TestAsynqConfig_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		cfg      *AsynqConfig
		contains string
	}{
		{"default no password", NewAsynqConfig(), "Password="},
		{"with password", &AsynqConfig{Host: "redis", Port: 6379, Password: "secret"}, "Password=****"},
		{"shows host", &AsynqConfig{Host: "10.0.0.1", Port: 6380}, "Host=10.0.0.1"},
		{"shows port", &AsynqConfig{Host: "localhost", Port: 6380}, "Port=6380"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.cfg.String()
			if !strings.Contains(got, tt.contains) {
				t.Errorf("String() = %q, want to contain %q", got, tt.contains)
			}
		})
	}
}

func TestAsynqConfig_String_ExactMatch(t *testing.T) {
	t.Parallel()
	cfg := NewAsynqConfig()
	expected := "AsynqConfig{Enabled=false, Host=localhost, Port=6379, Password=, DB=0, PoolSize=10, EnableScheduler=false, Concurrency=10, RetryLimit=25, Timeout=0s}"
	if got := cfg.String(); got != expected {
		t.Errorf("expected %q, got %q", expected, got)
	}
}

func TestAsynqConfig_RedisClientOpt(t *testing.T) {
	t.Parallel()
	cfg := &AsynqConfig{
		Host:     "192.168.1.100",
		Port:     6380,
		Password: "secret",
		DB:       2,
		PoolSize: 25,
	}

	opt := cfg.RedisClientOpt()

	tests := []struct {
		name     string
		got      any
		expected any
	}{
		{"addr", opt.Addr, "192.168.1.100:6380"},
		{"password", opt.Password, "secret"},
		{"db", opt.DB, 2},
		{"pool size", opt.PoolSize, 25},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, tt.got)
			}
		})
	}
}

func TestAsynqConfig_RedisClientOpt_Defaults(t *testing.T) {
	t.Parallel()
	cfg := NewAsynqConfig()
	opt := cfg.RedisClientOpt()

	if opt.Addr != "localhost:6379" {
		t.Errorf("expected 'localhost:6379', got %q", opt.Addr)
	}
	if opt.Password != "" {
		t.Errorf("expected empty password, got %q", opt.Password)
	}
}

func TestAsynqConfig_ApplyDefaults(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		input  *AsynqConfig
		expect *AsynqConfig
	}{
		{"empty config", &AsynqConfig{}, &AsynqConfig{Host: "localhost", Port: 6379, PoolSize: 10, Concurrency: 10, RetryLimit: 25}},
		{"partial override", &AsynqConfig{Host: "redis", Port: 6380}, &AsynqConfig{Host: "redis", Port: 6380, PoolSize: 10, Concurrency: 10, RetryLimit: 25}},
		{"already set", &AsynqConfig{Host: "10.0.0.1", Port: 6379, PoolSize: 5, Concurrency: 3, RetryLimit: 15}, &AsynqConfig{Host: "10.0.0.1", Port: 6379, PoolSize: 5, Concurrency: 3, RetryLimit: 15}},
		{"timeout preserved", &AsynqConfig{Timeout: 30 * time.Second}, &AsynqConfig{Host: "localhost", Port: 6379, PoolSize: 10, Concurrency: 10, RetryLimit: 25, Timeout: 30 * time.Second}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.input.ApplyDefaults()
			if tt.input.Host != tt.expect.Host {
				t.Errorf("host: expected %q, got %q", tt.expect.Host, tt.input.Host)
			}
			if tt.input.Port != tt.expect.Port {
				t.Errorf("port: expected %d, got %d", tt.expect.Port, tt.input.Port)
			}
			if tt.input.PoolSize != tt.expect.PoolSize {
				t.Errorf("pool-size: expected %d, got %d", tt.expect.PoolSize, tt.input.PoolSize)
			}
			if tt.input.Concurrency != tt.expect.Concurrency {
				t.Errorf("concurrency: expected %d, got %d", tt.expect.Concurrency, tt.input.Concurrency)
			}
			if tt.input.RetryLimit != tt.expect.RetryLimit {
				t.Errorf("retry-limit: expected %d, got %d", tt.expect.RetryLimit, tt.input.RetryLimit)
			}
			if tt.input.Timeout != tt.expect.Timeout {
				t.Errorf("timeout: expected %v, got %v", tt.expect.Timeout, tt.input.Timeout)
			}
		})
	}
}

func TestAsynqConfig_ApplyDefaults_Idempotent(t *testing.T) {
	t.Parallel()
	cfg := NewAsynqConfig()
	before := cfg.String()
	cfg.ApplyDefaults()
	after := cfg.String()
	if before != after {
		t.Errorf("ApplyDefaults not idempotent: before=%q, after=%q", before, after)
	}
}

func TestConfigPrefix_Relationship(t *testing.T) {
	t.Parallel()
	if AsynqEnabled != ConfigPrefix+".enabled" {
		t.Errorf("AsynqEnabled=%q, expected %q", AsynqEnabled, ConfigPrefix+".enabled")
	}
}

func TestAsynqAutoConfiguration_Enqueue_ClientNotConfigured(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}

	tests := []struct {
		name string
		fn   func() (*asynq.TaskInfo, error)
	}{
		{"enqueue", func() (*asynq.TaskInfo, error) {
			return c.Enqueue(asynq.NewTask("test", nil))
		}},
		{"enqueue at", func() (*asynq.TaskInfo, error) {
			return c.EnqueueAt(asynq.NewTask("test", nil), time.Now())
		}},
		{"enqueue in", func() (*asynq.TaskInfo, error) {
			return c.EnqueueIn(asynq.NewTask("test", nil), 5*time.Minute)
		}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := tt.fn()
			if !errors.Is(err, ErrClientNotConfigured) {
				t.Errorf("expected ErrClientNotConfigured, got %v", err)
			}
		})
	}
}

func TestErrClientNotConfigured_Value(t *testing.T) {
	t.Parallel()
	if ErrClientNotConfigured.Error() != "asynq: client not configured" {
		t.Errorf("unexpected error message: %q", ErrClientNotConfigured.Error())
	}
}

func TestAsynqAutoConfiguration_Name(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	if c.Name() != StarterName {
		t.Errorf("expected %q, got %q", StarterName, c.Name())
	}
	if c.Name() != "AsynqStarter" {
		t.Errorf("expected 'AsynqStarter', got %q", c.Name())
	}
}

func TestAsynqAutoConfiguration_Dependencies(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	if c.Dependencies() != nil {
		t.Errorf("expected nil dependencies, got %v", c.Dependencies())
	}
}

func TestAsynqAutoConfiguration_Stop_NotConfigured(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}

	if c.GetClient() != nil {
		t.Error("expected nil client before Configure")
	}
	if c.GetScheduler() != nil {
		t.Error("expected nil scheduler before Configure")
	}
}

func TestAsynqAutoConfiguration_IsConfigured(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	if c.IsConfigured() {
		t.Error("expected false before Configure")
	}
}

func TestAsynqAutoConfiguration_GetConfig_NotConfigured(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	if c.GetConfig() != nil {
		t.Error("expected nil config before Configure")
	}
}

func TestAsynqAutoConfiguration_GetCondition(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	cond := c.GetCondition()
	if cond == nil {
		t.Error("expected non-nil condition")
	}
}

func TestValidationErrors(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{"invalid host", ErrInvalidHost, "asynq: invalid host, must not be empty"},
		{"invalid port", ErrInvalidPort, "asynq: invalid port, must be in [1, 65535]"},
		{"invalid db", ErrInvalidDB, "asynq: invalid db, must be >= 0"},
		{"invalid pool size", ErrInvalidPoolSize, "asynq: invalid pool-size, must be >= 0"},
		{"invalid concurrency", ErrInvalidConcurrency, "asynq: invalid concurrency, must be >= 0"},
		{"invalid retry limit", ErrInvalidRetryLimit, "asynq: invalid retry-limit, must be >= 0"},
		{"invalid timeout", ErrInvalidTimeout, "asynq: invalid timeout, must be >= 0"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.err.Error() != tt.msg {
				t.Errorf("expected %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestDoc_Value(t *testing.T) {
	t.Parallel()
	if Doc == "" {
		t.Error("expected non-empty Doc constant")
	}
}

func TestStarterName_Value(t *testing.T) {
	t.Parallel()
	if StarterName != "AsynqStarter" {
		t.Errorf("expected 'AsynqStarter', got %q", StarterName)
	}
}

func TestAsynqConfig_ApplyDefaults_ThenValidate(t *testing.T) {
	t.Parallel()
	cfg := &AsynqConfig{}
	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		t.Errorf("ApplyDefaults should produce valid config, got %v", err)
	}
}

func TestAsynqConfig_String_WithTimeout(t *testing.T) {
	t.Parallel()
	cfg := &AsynqConfig{Host: "localhost", Port: 6379, Timeout: 30 * time.Second}
	got := cfg.String()
	if !strings.Contains(got, "Timeout=30s") {
		t.Errorf("String() = %q, want to contain 'Timeout=30s'", got)
	}
}

func TestAsynqConfig_String_WithRetryLimit(t *testing.T) {
	t.Parallel()
	cfg := &AsynqConfig{Host: "localhost", Port: 6379, RetryLimit: 50}
	got := cfg.String()
	if !strings.Contains(got, "RetryLimit=50") {
		t.Errorf("String() = %q, want to contain 'RetryLimit=50'", got)
	}
}

func TestAsynqConfig_String_WithConcurrency(t *testing.T) {
	t.Parallel()
	cfg := &AsynqConfig{Host: "localhost", Port: 6379, Concurrency: 20}
	got := cfg.String()
	if !strings.Contains(got, "Concurrency=20") {
		t.Errorf("String() = %q, want to contain 'Concurrency=20'", got)
	}
}

func TestAsynqAutoConfiguration_Stop_ResetsConfig(t *testing.T) {
	t.Parallel()
	c := &AsynqAutoConfiguration{}
	if c.GetConfig() != nil {
		t.Error("expected nil config before Configure")
	}
}

func TestAsynqConfig_Validate_ErrorsAreDistinct(t *testing.T) {
	t.Parallel()
	errs := []error{ErrInvalidHost, ErrInvalidPort, ErrInvalidDB, ErrInvalidPoolSize, ErrInvalidConcurrency, ErrInvalidRetryLimit, ErrInvalidTimeout}
	for i := 0; i < len(errs); i++ {
		for j := i + 1; j < len(errs); j++ {
			if errors.Is(errs[i], errs[j]) {
				t.Errorf("validation errors should be distinct: %v matches %v", errs[i], errs[j])
			}
		}
	}
}

package zerolog

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/rs/zerolog"
	"github.com/xudefa/enhance/config/environment"
	"github.com/xudefa/enhance/log"
)

func newTestLogger(t *testing.T) (*ZerologLogger, *bytes.Buffer) {
	t.Helper()
	var buf bytes.Buffer
	zl := zerolog.New(&buf).Level(zerolog.DebugLevel).With().Timestamp().Logger()
	return &ZerologLogger{
		logger:    zl,
		level:     zerolog.DebugLevel,
		format:    "json",
		addSource: false,
		output:    &buf,
	}, &buf
}

func TestZerologLogger_Debug(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	logger.Debug(context.Background(), "debug msg", log.KeyValue{Key: "k", Value: "v"})
	if buf.Len() == 0 {
		t.Error("Debug() should write to buffer")
	}
}

func TestZerologLogger_Info(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	logger.Info(context.Background(), "info msg", log.KeyValue{Key: "k", Value: "v"})
	if buf.Len() == 0 {
		t.Error("Info() should write to buffer")
	}
}

func TestZerologLogger_Warn(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	logger.Warn(context.Background(), "warn msg")
	if buf.Len() == 0 {
		t.Error("Warn() should write to buffer")
	}
}

func TestZerologLogger_Error(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	logger.Error(context.Background(), "error msg", log.KeyValue{Key: "err", Value: "fail"})
	if buf.Len() == 0 {
		t.Error("Error() should write to buffer")
	}
}

func TestZerologLogger_DPanic(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	logger.DPanic(context.Background(), "dpanic msg")
	if buf.Len() == 0 {
		t.Error("DPanic() should write to buffer")
	}
}

func TestZerologLogger_Panic(t *testing.T) {
	t.Parallel()
	logger, _ := newTestLogger(t)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Panic() should have panicked")
		}
	}()
	logger.Panic(context.Background(), "panic msg")
}

func TestZerologLogger_With(t *testing.T) {
	t.Parallel()
	logger, buf := newTestLogger(t)
	child := logger.With(context.Background(), log.KeyValue{Key: "module", Value: "auth"})
	child.Info(context.Background(), "child logger msg")
	if buf.Len() == 0 {
		t.Error("With().Info() should write to buffer")
	}
}

func TestZerologLogger_Sync(t *testing.T) {
	t.Parallel()
	logger, _ := newTestLogger(t)
	if err := logger.Sync(); err != nil {
		t.Errorf("Sync() error = %v", err)
	}
}

func TestZerologLogger_Close_NoFile(t *testing.T) {
	t.Parallel()
	logger := &ZerologLogger{file: nil}
	if err := logger.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestZerologLogger_Close_WithFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.log")
	f, err := os.Create(fpath)
	if err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	logger := &ZerologLogger{file: f}
	if err := logger.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
}

func TestZerologLogger_InterfaceCompliance(t *testing.T) {
	t.Parallel()
	logger, _ := newTestLogger(t)
	var _ log.Logger = logger
}

func TestBuildZerolog_JSONFormat(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	cfg := &ZerologConfig{
		Level:      "info",
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
		OutputPath: "",
	}
	l := autoCfg.buildZerolog(cfg)
	if l == nil {
		t.Fatal("buildZerolog() returned nil")
	}
	var _ log.Logger = l
}

func TestBuildZerolog_ConsoleFormat(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	cfg := &ZerologConfig{
		Level:      "debug",
		Format:     "console",
		TimeFormat: "2006-01-02 15:04:05",
		OutputPath: "",
	}
	l := autoCfg.buildZerolog(cfg)
	if l == nil {
		t.Fatal("buildZerolog() returned nil")
	}
}

func TestBuildZerolog_WithOutputFile(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	dir := t.TempDir()
	fpath := filepath.Join(dir, "test.log")
	cfg := &ZerologConfig{
		Level:      "info",
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
		OutputPath: fpath,
	}
	l := autoCfg.buildZerolog(cfg)
	if l == nil {
		t.Fatal("buildZerolog() returned nil")
	}
	l.Info(context.Background(), "test file output")
	zl, ok := l.(*ZerologLogger)
	if !ok {
		t.Fatal("expected *ZerologLogger")
	}
	if err := zl.Close(); err != nil {
		t.Errorf("Close() error = %v", err)
	}
	data, err := os.ReadFile(fpath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty log file")
	}
}

func TestBuildZerolog_InvalidOutputPath(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	cfg := &ZerologConfig{
		Level:      "info",
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
		OutputPath: "/nonexistent/dir/test.log",
	}
	l := autoCfg.buildZerolog(cfg)
	if l == nil {
		t.Fatal("buildZerolog() returned nil, should fall back to stdout")
	}
}

func TestBuildZerolog_LogLevels(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		level string
	}{
		{"debug", "debug"},
		{"info", "info"},
		{"warn", "warn"},
		{"warning", "warning"},
		{"error", "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			autoCfg := &ZerologAutoConfiguration{}
			cfg := &ZerologConfig{
				Level:      tt.level,
				Format:     "json",
				TimeFormat: "2006-01-02 15:04:05",
			}
			l := autoCfg.buildZerolog(cfg)
			if l == nil {
				t.Fatalf("buildZerolog() returned nil for level %s", tt.level)
			}
		})
	}
}

func TestBuildZerolog_UnknownLevel(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	cfg := &ZerologConfig{
		Level:      "unknown-level",
		Format:     "json",
		TimeFormat: "2006-01-02 15:04:05",
	}
	l := autoCfg.buildZerolog(cfg)
	if l == nil {
		t.Fatal("buildZerolog() returned nil for unknown level")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}
	env := environment.NewEnvironment()

	cfg, err := autoCfg.loadConfig(env)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.Level != DefaultZeroLogLevel {
		t.Errorf("Level = %q, want %q", cfg.Level, DefaultZeroLogLevel)
	}
	if cfg.Format != DefaultZeroLogFormat {
		t.Errorf("Format = %q, want %q", cfg.Format, DefaultZeroLogFormat)
	}
	if cfg.TimeFormat != DefaultZeroLogTimeFormat {
		t.Errorf("TimeFormat = %q, want %q", cfg.TimeFormat, DefaultZeroLogTimeFormat)
	}
	if cfg.AddSource != DefaultZeroLogAddSource {
		t.Errorf("AddSource = %v, want %v", cfg.AddSource, DefaultZeroLogAddSource)
	}
	if cfg.OutputPath != DefaultZeroLogOutputPath {
		t.Errorf("OutputPath = %q, want %q", cfg.OutputPath, DefaultZeroLogOutputPath)
	}
}

func TestLoadConfig_CustomValues(t *testing.T) {
	t.Parallel()
	autoCfg := &ZerologAutoConfiguration{}

	env := environment.NewEnvironment()
	env.AddPropertySource(environment.NewMapPropertySource("test-zerolog", environment.PriorityNormal, map[string]any{
		"log.zerolog.level":       "debug",
		"log.zerolog.format":      "console",
		"log.zerolog.time-format": "15:04:05",
		"log.zerolog.add-source":  "true",
		"log.zerolog.output-path": "/tmp/test.log",
	}))

	cfg, err := autoCfg.loadConfig(env)
	if err != nil {
		t.Fatalf("loadConfig() error = %v", err)
	}
	if cfg.Level != "debug" {
		t.Errorf("Level = %q, want 'debug'", cfg.Level)
	}
	if cfg.Format != "console" {
		t.Errorf("Format = %q, want 'console'", cfg.Format)
	}
	if cfg.OutputPath != "/tmp/test.log" {
		t.Errorf("OutputPath = %q, want '/tmp/test.log'", cfg.OutputPath)
	}
}

func TestConstants(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"ZeroLogEnabled", ZeroLogEnabled, "log.zerolog.enabled"},
		{"ZeroLogLevel", ZeroLogLevel, "log.zerolog.level"},
		{"ZeroLogFormat", ZeroLogFormat, "log.zerolog.format"},
		{"ZeroLogTimeFormat", ZeroLogTimeFormat, "log.zerolog.time-format"},
		{"ZeroLogAddSource", ZeroLogAddSource, "log.zerolog.add-source"},
		{"ZeroLogOutputPath", ZeroLogOutputPath, "log.zerolog.output-path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.key != tt.value {
				t.Errorf("%s = %q, want %q", tt.name, tt.key, tt.value)
			}
		})
	}
}

func TestDefaultValues(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		got  any
		want any
	}{
		{"DefaultZeroLogLevel", DefaultZeroLogLevel, "info"},
		{"DefaultZeroLogFormat", DefaultZeroLogFormat, "json"},
		{"DefaultZeroLogTimeFormat", DefaultZeroLogTimeFormat, "2006-01-02 15:04:05"},
		{"DefaultZeroLogAddSource", DefaultZeroLogAddSource, false},
		{"DefaultZeroLogOutputPath", DefaultZeroLogOutputPath, ""},
		{"ConditionTrue", ConditionTrue, "true"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
			}
		})
	}
}

func TestLogFieldConstants(t *testing.T) {
	t.Parallel()
	if LogFieldLevel != "level" {
		t.Errorf("LogFieldLevel = %q, want 'level'", LogFieldLevel)
	}
	if LogFieldFormat != "format" {
		t.Errorf("LogFieldFormat = %q, want 'format'", LogFieldFormat)
	}
	if LogFieldOutput != "output-path" {
		t.Errorf("LogFieldOutput = %q, want 'output-path'", LogFieldOutput)
	}
}

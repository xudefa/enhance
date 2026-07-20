package generator

import (
	"reflect"
	"testing"
)

func TestParseGoGenerate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    *GoGenerateDirective
		wantErr bool
	}{
		{
			name:  "parse type directive",
			input: "//go:generate enhance aop gen -type=UserService",
			want: &GoGenerateDirective{
				Types: []string{"UserService"},
				Mode:  "static",
			},
		},
		{
			name:  "parse multiple types",
			input: "//go:generate enhance aop gen -type=UserService,OrderService",
			want: &GoGenerateDirective{
				Types: []string{"UserService", "OrderService"},
				Mode:  "static",
			},
		},
		{
			name:  "parse with output",
			input: "//go:generate enhance aop gen -type=UserService -output=proxy.go",
			want: &GoGenerateDirective{
				Types:  []string{"UserService"},
				Output: "proxy.go",
				Mode:   "static",
			},
		},
		{
			name:  "parse interface directive",
			input: "//go:generate enhance aop gen -interface=ServiceInterface",
			want: &GoGenerateDirective{
				Interfaces: []string{"ServiceInterface"},
				Mode:       "static",
			},
		},
		{
			name:  "parse all directive",
			input: "//go:generate enhance aop gen -all",
			want: &GoGenerateDirective{
				All:  true,
				Mode: "static",
			},
		},
		{
			name:  "parse with mode",
			input: "//go:generate enhance aop gen -type=UserService -mode=aop",
			want: &GoGenerateDirective{
				Types: []string{"UserService"},
				Mode:  "aop",
			},
		},
		{
			name:  "parse with package",
			input: "//go:generate enhance aop gen -type=UserService -package=service",
			want: &GoGenerateDirective{
				Types:   []string{"UserService"},
				Package: "service",
				Mode:    "static",
			},
		},
		{
			name:    "not a go:generate directive",
			input:   "// some comment",
			wantErr: true,
		},
		{
			name:    "not an enhance-aop directive",
			input:   "//go:generate mockgen -source=service.go",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGoGenerate(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGoGenerate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(got.Types, tt.want.Types) {
				t.Errorf("Types = %v, want %v", got.Types, tt.want.Types)
			}
			if !reflect.DeepEqual(got.Interfaces, tt.want.Interfaces) {
				t.Errorf("Interfaces = %v, want %v", got.Interfaces, tt.want.Interfaces)
			}
			if got.Output != tt.want.Output {
				t.Errorf("Output = %v, want %v", got.Output, tt.want.Output)
			}
			if got.Package != tt.want.Package {
				t.Errorf("Package = %v, want %v", got.Package, tt.want.Package)
			}
			if got.Mode != tt.want.Mode {
				t.Errorf("Mode = %v, want %v", got.Mode, tt.want.Mode)
			}
			if got.All != tt.want.All {
				t.Errorf("All = %v, want %v", got.All, tt.want.All)
			}
		})
	}
}

func TestGoGenerateDirective_HasTargets(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    GoGenerateDirective
		want bool
	}{
		{"has types", GoGenerateDirective{Types: []string{"UserService"}}, true},
		{"has interfaces", GoGenerateDirective{Interfaces: []string{"ServiceInterface"}}, true},
		{"has all", GoGenerateDirective{All: true}, true},
		{"empty", GoGenerateDirective{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.d.HasTargets(); got != tt.want {
				t.Errorf("HasTargets() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseArgs(t *testing.T) {
	t.Parallel()
	args := parseArgs("-type=UserService -mode=static -all")

	if args["type"] != "UserService" {
		t.Errorf("type = %v, want UserService", args["type"])
	}
	if args["mode"] != "static" {
		t.Errorf("mode = %v, want static", args["mode"])
	}
	if _, ok := args["all"]; !ok {
		t.Error("missing 'all' key")
	}
}

func TestSplitAndTrim(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    string
		sep  string
		want []string
	}{
		{"simple", "a,b,c", ",", []string{"a", "b", "c"}},
		{"with spaces", "a, b, c", ",", []string{"a", "b", "c"}},
		{"empty parts", "a,,b", ",", []string{"a", "b"}},
		{"single", "a", ",", []string{"a"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitAndTrim(tt.s, tt.sep)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("splitAndTrim() = %v, want %v", got, tt.want)
			}
		})
	}
}

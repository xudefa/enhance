package spel

import (
	"math"
	"testing"
)

func TestToFloat64_AllTypes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		input  any
		want   float64
		wantOK bool
	}{
		{"int", int(42), 42.0, true},
		{"int8", int8(8), 8.0, true},
		{"int16", int16(16), 16.0, true},
		{"int32", int32(32), 32.0, true},
		{"int64", int64(64), 64.0, true},
		{"uint", uint(10), 10.0, true},
		{"uint8", uint8(8), 8.0, true},
		{"uint16", uint16(16), 16.0, true},
		{"uint32", uint32(32), 32.0, true},
		{"uint64", uint64(64), 64.0, true},
		{"float32", float32(1.5), 1.5, true},
		{"float64", float64(2.5), 2.5, true},
		{"string", "abc", 0, false},
		{"bool", true, 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tt.input)
			if ok != tt.wantOK {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.wantOK)
			}
			if ok && got != tt.want {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestArithmetic_AllOps(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		left    any
		right   any
		op      string
		want    float64
		wantErr bool
	}{
		{"add", 3.0, 2.0, "+", 5.0, false},
		{"subtract", 5.0, 2.0, "-", 3.0, false},
		{"multiply", 3.0, 4.0, "*", 12.0, false},
		{"divide", 10.0, 2.0, "/", 5.0, false},
		{"divide by zero", 10.0, 0.0, "/", 0, true},
		{"unsupported op", 1.0, 2.0, "%", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := arithmetic(tt.left, tt.right, tt.op)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			gotF, _ := toFloat64(got)
			if math.Abs(gotF-tt.want) > 1e-9 {
				t.Errorf("arithmetic(%v, %v, %q) = %v, want %v", tt.left, tt.right, tt.op, got, tt.want)
			}
		})
	}
}

func TestArithmetic_NonNumericOperands(t *testing.T) {
	t.Parallel()

	_, err := arithmetic("a", "b", "+")
	if err == nil {
		t.Error("expected error for non-numeric operands")
	}
}

func TestSplitArgsRespectingQuotes_Var(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"single", "a", []string{"a"}},
		{"multiple", "a,b,c", []string{"a", "b", "c"}},
		{"quoted comma", "'a,b',c", []string{"'a,b'", "c"}},
		{"double quoted", `"'a,b'",c`, []string{`"'a,b'"`, "c"}},
		{"spaces", " a , b ", []string{" a ", " b "}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := splitArgsRespectingQuotes(tt.input)
			if len(got) != len(tt.expected) {
				t.Fatalf("got %d args, want %d: %v", len(got), len(tt.expected), got)
			}
			for i, v := range got {
				if v != tt.expected[i] {
					t.Errorf("got[%d] = %q, want %q", i, v, tt.expected[i])
				}
			}
		})
	}
}

func TestIsTruthy_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"nil", nil, false},
		{"true", true, true},
		{"false", false, false},
		{"zero int", 0, false},
		{"non-zero int", 42, true},
		{"zero float", 0.0, false},
		{"non-zero float", 1.5, true},
		{"empty string", "", false},
		{"non-empty string", "hello", true},
		{"zero uint", uint(0), false},
		{"non-zero uint", uint(1), true},
		{"struct", struct{}{}, true},
		{"slice", []int{}, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := isTruthy(tt.input); got != tt.expected {
				t.Errorf("isTruthy(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsNumber_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected bool
	}{
		{"42", true},
		{"3.14", true},
		{"-10", true},
		{"-3.5", true},
		{"abc", false},
		{"", false},
		{"12abc", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isNumber(tt.input); got != tt.expected {
				t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseInt_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected int64
		wantErr  bool
	}{
		{"42", 42, false},
		{"-10", -10, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
		{"12abc", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			got, err := parseInt(tt.input, 10, 64)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("parseInt(%q) = %d, want %d", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseFloat_Extended(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{"valid_3_14", "3.14", 3.14, false},
		{"valid_neg_2_5", "-2.5", -2.5, false},
		{"valid_int_str", "42", 42.0, false},
		{"invalid_abc", "abc", 0, true},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseFloat(tt.input, 64)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("parseFloat(%q) = %f, want %f", tt.input, got, tt.expected)
			}
		})
	}
}

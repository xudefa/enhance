package spel

import (
	"testing"
)

func TestSplitArgsRespectingQuotes(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string
		expected []string
	}{
		{"empty", "", []string{}},
		{"single", "a", []string{"a"}},
		{"multiple", "a,b,c", []string{"a", "b", "c"}},
		{"quoted comma", "'a,b',c", []string{"'a,b'", "c"}},
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

func TestIsTruthyAdvanced(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected bool
	}{
		{"uint positive", uint(1), true},
		{"uint zero", uint(0), false},
		{"float positive", 1.5, true},
		{"float zero", 0.0, false},
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

func TestIsSimplePropertyEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"123", false},
		{"abc", true},
		{"_abc", true},
		{"a1b2", true},
		{"abc_def", true},
		{"abc.def", false},
		{"abc-def", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isSimpleProperty(tt.input); got != tt.expected {
				t.Errorf("isSimpleProperty(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestIsLiteralEdgeCases(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected bool
	}{
		{"", false},
		{"true", true},
		{"false", true},
		{"null", true},
		{"TRUE", true},
		{"123", true},
		{"abc", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.input, func(t *testing.T) {
			t.Parallel()
			if got := isLiteral(tt.input); got != tt.expected {
				t.Errorf("isLiteral(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToInt64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected int64
	}{
		{int(42), 42},
		{int8(8), 8},
		{int16(16), 16},
		{int32(32), 32},
		{int64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toInt64(tt.input); got != tt.expected {
				t.Errorf("toInt64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToUint64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected uint64
	}{
		{uint(42), 42},
		{uint8(8), 8},
		{uint16(16), 16},
		{uint32(32), 32},
		{uint64(64), 64},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toUint64(tt.input); got != tt.expected {
				t.Errorf("toUint64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64Value(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    any
		expected float64
	}{
		{int(42), 42.0},
		{int8(8), 8.0},
		{int16(16), 16.0},
		{int32(32), 32.0},
		{int64(64), 64.0},
		{uint(10), 10.0},
		{uint8(8), 8.0},
		{uint16(16), 16.0},
		{uint32(32), 32.0},
		{uint64(64), 64.0},
		{float32(1.5), 1.5},
		{float64(2.5), 2.5},
		{"not a number", 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(t *testing.T) {
			t.Parallel()
			if got := toFloat64Value(tt.input); got != tt.expected {
				t.Errorf("toFloat64Value(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestToFloat64(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    any
		expected float64
		ok       bool
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
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, ok := toFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && got != tt.expected {
				t.Errorf("toFloat64(%v) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestEqualsCrossTypeComparisons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		left     any
		right    any
		expected bool
	}{
		{"int vs int8", int(8), int8(8), true},
		{"int vs int16", int(16), int16(16), true},
		{"int vs int32", int(32), int32(32), true},
		{"int vs int64", int(64), int64(64), true},
		{"int vs uint8", int(8), uint8(8), true},
		{"int vs uint16", int(16), uint16(16), true},
		{"int vs uint32", int(32), uint32(32), true},
		{"int vs float32", int(1), float32(1), true},
		{"int vs float64", int(1), float64(1), true},
		{"int64 vs int", int64(42), int(42), true},
		{"int64 vs int8", int64(8), int8(8), true},
		{"int64 vs int16", int64(16), int16(16), true},
		{"int64 vs int32", int64(32), int32(32), true},
		{"int64 vs uint8", int64(8), uint8(8), true},
		{"int64 vs uint16", int64(16), uint16(16), true},
		{"int64 vs uint32", int64(32), uint32(32), true},
		{"int64 vs float32", int64(1), float32(1), true},
		{"int64 vs float64", int64(1), float64(1), true},
		{"uint vs uint", uint(42), uint(42), true},
		{"uint vs uint8", uint(8), uint8(8), true},
		{"uint vs uint16", uint(16), uint16(16), true},
		{"uint vs uint32", uint(32), uint32(32), true},
		{"uint vs uint64", uint(64), uint64(64), true},
		{"uint vs float32", uint(1), float32(1), true},
		{"uint vs float64", uint(1), float64(1), true},
		{"uint64 vs int8", uint64(8), int8(8), true},
		{"uint64 vs int16", uint64(16), int16(16), true},
		{"uint64 vs int32", uint64(32), int32(32), true},
		{"uint64 vs uint", uint64(42), uint(42), true},
		{"uint64 vs uint8", uint64(8), uint8(8), true},
		{"uint64 vs uint16", uint64(16), uint16(16), true},
		{"uint64 vs uint32", uint64(32), uint32(32), true},
		{"uint64 vs uint64", uint64(64), uint64(64), true},
		{"uint64 vs float32", uint64(1), float32(1), true},
		{"uint64 vs float64", uint64(1), float64(1), true},
		{"float32 vs int", float32(42), int(42), true},
		{"float32 vs int8", float32(8), int8(8), true},
		{"float32 vs int16", float32(16), int16(16), true},
		{"float32 vs int32", float32(32), int32(32), true},
		{"float32 vs int64", float32(64), int64(64), true},
		{"float32 vs uint", float32(42), uint(42), true},
		{"float32 vs uint8", float32(8), uint8(8), true},
		{"float32 vs uint16", float32(16), uint16(16), true},
		{"float32 vs uint32", float32(32), uint32(32), true},
		{"float32 vs uint64", float32(64), uint64(64), true},
		{"float32 vs float64", float32(1.5), float64(1.5), true},
		{"float64 vs int", float64(42), int(42), true},
		{"float64 vs int8", float64(8), int8(8), true},
		{"float64 vs int16", float64(16), int16(16), true},
		{"float64 vs int32", float64(32), int32(32), true},
		{"float64 vs int64", float64(64), int64(64), true},
		{"float64 vs uint", float64(42), uint(42), true},
		{"float64 vs uint8", float64(8), uint8(8), true},
		{"float64 vs uint16", float64(16), uint16(16), true},
		{"float64 vs uint32", float64(32), uint32(32), true},
		{"float64 vs uint64", float64(64), uint64(64), true},
		{"float64 vs float32", float64(1.5), float32(1.5), true},
		{"int8 vs int", int8(42), int(42), true},
		{"int16 vs int", int16(42), int(42), true},
		{"int32 vs int", int32(42), int(42), true},
		{"uint8 vs uint", uint8(42), uint(42), true},
		{"uint16 vs uint", uint16(42), uint(42), true},
		{"uint32 vs uint", uint32(42), uint(42), true},
		{"float64 vs string", float64(1), "a", false},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := equals(tt.left, tt.right); got != tt.expected {
				t.Errorf("equals(%v, %v) = %v, want %v", tt.left, tt.right, got, tt.expected)
			}
		})
	}
}

func TestCompareValuesUnsupported(t *testing.T) {
	t.Parallel()
	_, err := compareValues("a", "b", "~")
	if err == nil {
		t.Error("expected error for unsupported operator")
	}
}

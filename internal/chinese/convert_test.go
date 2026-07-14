package chinese

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"零", 0},
		{"一", 1},
		{"五", 5},
		{"九", 9},
		{"一二三", 123},
		{"四五六", 456},
		{"九八七", 987},
		{"零零零", 0},
	}
	for _, tt := range tests {
		got, err := Parse(tt.input)
		if err != nil {
			t.Errorf("Parse(%q) returned error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("Parse(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestParseError(t *testing.T) {
	_, err := Parse("十二") // multi-char not supported
	if err == nil {
		t.Error("Parse(\"十二\") expected error, got nil")
	}
	_, err = Parse("abc")
	if err == nil {
		t.Error("Parse(\"abc\") expected error, got nil")
	}
	_, err = Parse("")
	if err != nil {
		// empty string should parse as 0 (loop doesn't execute)
		t.Logf("Parse(\"\") returned error (acceptable): %v", err)
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "零"},
		{1, "一"},
		{5, "五"},
		{9, "九"},
		{123, "一二三"},
		{456, "四五六"},
		{987, "九八七"},
		{-1, "负一"},
		{-123, "负一二三"},
		{100, "一零零"},
	}
	for _, tt := range tests {
		got := Format(tt.input)
		if got != tt.want {
			t.Errorf("Format(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "零"},
		{1.0, "一"},
		{3.5, "三点五"},
		{10.1, "一零点一"},
		{0.5, "零点五"},
		{-3.5, "负三点五"},
		{math.NaN(), "NaN"},
		{math.Inf(1), "Inf"},
		{math.Inf(-1), "负Inf"},
	}
	for _, tt := range tests {
		got := FormatFloat(tt.input)
		if got != tt.want {
			t.Errorf("FormatFloat(%v) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestIsWhole(t *testing.T) {
	tests := []struct {
		input float64
		want  bool
	}{
		{0, true},
		{1, true},
		{1.5, false},
		{-3, true},
		{-3.14, false},
		{1 << 53, true},     // 9007199254740992 — целое
		{1<<53 + 2.0, true}, // 9007199254740994 — чётное, представимо
		{1<<53 + 1.0, true}, // float64 не отличает от 2^53, так что это 2^53 — целое
	}
	for _, tt := range tests {
		got := IsWhole(tt.input)
		if got != tt.want {
			t.Errorf("IsWhole(%v) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestRoundtrip(t *testing.T) {
	// Only non-negative values roundtrip (Parse doesn't handle "负" prefix)
	values := []int{0, 1, 9, 10, 99, 100, 123456789}
	for _, v := range values {
		s := Format(v)
		parsed, err := Parse(s)
		if err != nil {
			t.Errorf("Parse(Format(%d)) error: %v", v, err)
			continue
		}
		if parsed != v {
			t.Errorf("roundtrip Format(%d) = %q, Parse() = %d", v, s, parsed)
		}
	}
}

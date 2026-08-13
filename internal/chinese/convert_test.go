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
		{"一二三", 123},
		{"零零零", 0},
		{"负一二三", -123},
	}
	for _, test := range tests {
		got, err := Parse(test.input)
		if err != nil {
			t.Errorf("Parse(%q) 返回错误：%v", test.input, err)
			continue
		}
		if got != test.want {
			t.Errorf("Parse(%q) = %d，期望 %d", test.input, got, test.want)
		}
	}
}

func TestParseError(t *testing.T) {
	for _, input := range []string{"十二", "abc", "负"} {
		if _, err := Parse(input); err == nil {
			t.Errorf("Parse(%q) 应返回错误", input)
		}
	}
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"三点五", 3.5},
		{"负零点二五", -0.25},
		{"一二三", 123},
	}
	for _, test := range tests {
		got, err := ParseFloat(test.input)
		if err != nil || got != test.want {
			t.Errorf("ParseFloat(%q) = %v, %v；期望 %v", test.input, got, err, test.want)
		}
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "零"},
		{123, "一二三"},
		{-123, "负一二三"},
		{100, "一零零"},
	}
	for _, test := range tests {
		if got := Format(test.input); got != test.want {
			t.Errorf("Format(%d) = %q，期望 %q", test.input, got, test.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	tests := []struct {
		input float64
		want  string
	}{
		{0, "零"},
		{3.5, "三点五"},
		{0.25, "零点二五"},
		{-3.5, "负三点五"},
		{math.NaN(), "NaN"},
		{math.Inf(1), "Inf"},
		{math.Inf(-1), "负Inf"},
	}
	for _, test := range tests {
		if got := FormatFloat(test.input); got != test.want {
			t.Errorf("FormatFloat(%v) = %q，期望 %q", test.input, got, test.want)
		}
	}
}

func TestRoundtrip(t *testing.T) {
	values := []float64{0, 1, -9, 10.25, -0.125, 123456.75}
	for _, value := range values {
		formatted := FormatFloat(value)
		parsed, err := ParseFloat(formatted)
		if err != nil || parsed != value {
			t.Errorf("往返转换失败：%v → %q → %v（%v）", value, formatted, parsed, err)
		}
	}
}

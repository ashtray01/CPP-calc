// Package chinese implements single-digit Chinese numeral conversion.
//
// Each Chinese character maps to exactly one decimal digit (0–9).
// Multi-digit Chinese numbers like 十二 (12) or 三百 (300) are NOT supported.
//
// Maps:
//
//	零→0, 一→1, 二→2, 三→3, 四→4, 五→5, 六→6, 七→7, 八→8, 九→9
package chinese

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// DigitMaps — bidirectional mapping between Chinese and Arabic digits.
var (
	ToArabic = map[rune]int{
		'零': 0, '一': 1, '二': 2, '三': 3, '四': 4,
		'五': 5, '六': 6, '七': 7, '八': 8, '九': 9,
	}
	ToChinese = map[int]rune{
		0: '零', 1: '一', 2: '二', 3: '三', 4: '四',
		5: '五', 6: '六', 7: '七', 8: '八', 9: '九',
	}
)

// ChineseDigits is the ordered list of Chinese digit runes.
var ChineseDigits = []rune{'零', '一', '二', '三', '四', '五', '六', '七', '八', '九'}

// Parse converts a Chinese numeral string to an int.
// Each character in s must be a Chinese digit (零…九).
// Returns an error if any character is not a valid Chinese digit.
func Parse(s string) (int, error) {
	result := 0
	for _, ch := range s {
		if v, ok := ToArabic[ch]; ok {
			result = result*10 + v
		} else {
			return 0, fmt.Errorf("invalid Chinese digit character: %c", ch)
		}
	}
	return result, nil
}

// Format converts an int to a Chinese numeral string.
// Supports negative numbers (prefix "负").
func Format(n int) string {
	if n == 0 {
		return "零"
	}
	prefix := ""
	if n < 0 {
		prefix = "负"
		n = -n
	}
	var buf strings.Builder
	for _, ch := range strconv.Itoa(n) {
		buf.WriteRune(ToChinese[int(ch-'0')])
	}
	return prefix + buf.String()
}

// FormatFloat converts a float64 to a Chinese numeral string with decimal point.
// If the value is a whole number, delegates to Format.
// Supports negative numbers (prefix "负").
// NaN and Inf are returned as literal strings.
func FormatFloat(f float64) string {
	if math.IsNaN(f) {
		return "NaN"
	}
	if math.IsInf(f, 1) {
		return "Inf"
	}
	if math.IsInf(f, -1) {
		return "负Inf"
	}
	if f == 0 {
		return "零"
	}
	if f == math.Trunc(f) {
		return Format(int(f))
	}
	prefix := ""
	if f < 0 {
		prefix = "负"
		f = -f
	}
	intPart := int(math.Floor(f))
	fracPart := f - math.Floor(f)

	result := Format(intPart) + "点"
	fracStr := fmt.Sprintf("%.10f", fracPart)
	fracStr = strings.TrimRight(fracStr, "0")
	fracStr = strings.TrimLeft(fracStr, "0.")
	for _, ch := range fracStr {
		result += string(ToChinese[int(ch-'0')])
	}
	return prefix + result
}

// IsWhole reports whether f is a whole number representable as int
// without precision loss. Uses a safe epsilon-based comparison for
// values within the int64 range, and falls back to Trunc comparison
// for larger values (which may have floating-point precision limits).
func IsWhole(f float64) bool {
	const maxSafeInt = 1 << 53 // 2^53 — точный диапазон float64 для целых
	if f > maxSafeInt || f < -maxSafeInt {
		return f == math.Trunc(f)
	}
	return f == float64(int(f))
}

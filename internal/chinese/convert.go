// Package chinese 提供逐位中文数字与阿拉伯数字之间的转换。
package chinese

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// ToArabic 和 ToChinese 提供双向数字映射。
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

// ChineseDigits 按从零到九的顺序保存中文数字。
var ChineseDigits = []rune{'零', '一', '二', '三', '四', '五', '六', '七', '八', '九'}

// Parse 将不带小数点的逐位中文数字转换为整数。
func Parse(value string) (int, error) {
	if value == "" {
		return 0, nil
	}
	sign := 1
	value = strings.TrimSpace(value)
	if strings.HasPrefix(value, "负") {
		sign = -1
		value = strings.TrimPrefix(value, "负")
	}
	if value == "" {
		return 0, fmt.Errorf("缺少数字")
	}
	result := 0
	for _, character := range value {
		digit, ok := ToArabic[character]
		if !ok {
			return 0, fmt.Errorf("无效的中文数字：%c", character)
		}
		result = result*10 + digit
	}
	return sign * result, nil
}

// ParseFloat 将包含“负”和“点”的逐位中文数字转换为浮点数。
func ParseFloat(value string) (float64, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, fmt.Errorf("缺少数字")
	}
	var builder strings.Builder
	for index, character := range []rune(value) {
		switch {
		case character == '负' && index == 0:
			builder.WriteByte('-')
		case character == '点':
			builder.WriteByte('.')
		default:
			digit, ok := ToArabic[character]
			if !ok {
				return 0, fmt.Errorf("无效的中文数字：%c", character)
			}
			builder.WriteByte(byte('0' + digit))
		}
	}
	result, err := strconv.ParseFloat(builder.String(), 64)
	if err != nil {
		return 0, fmt.Errorf("无法解析数字：%w", err)
	}
	return result, nil
}

// Format 将整数转换为逐位中文数字，并保留负号。
func Format(number int) string {
	if number == 0 {
		return "零"
	}
	prefix := ""
	if number < 0 {
		prefix = "负"
		number = -number
	}
	var builder strings.Builder
	for _, character := range strconv.Itoa(number) {
		builder.WriteRune(ToChinese[int(character-'0')])
	}
	return prefix + builder.String()
}

// FormatFloat 将浮点数转换为最多十位小数的中文显示文本。
func FormatFloat(number float64) string {
	if math.IsNaN(number) {
		return "NaN"
	}
	if math.IsInf(number, 1) {
		return "Inf"
	}
	if math.IsInf(number, -1) {
		return "负Inf"
	}
	if number == 0 {
		return "零"
	}
	if IsWhole(number) && number <= float64(math.MaxInt) && number >= float64(math.MinInt) {
		return Format(int(number))
	}

	prefix := ""
	if number < 0 {
		prefix = "负"
		number = -number
	}
	decimal := strconv.FormatFloat(number, 'f', 10, 64)
	decimal = strings.TrimRight(strings.TrimRight(decimal, "0"), ".")
	var builder strings.Builder
	for _, character := range decimal {
		if character == '.' {
			builder.WriteRune('点')
			continue
		}
		builder.WriteRune(ToChinese[int(character-'0')])
	}
	return prefix + builder.String()
}

// IsWhole 判断浮点数是否为有限整数。
func IsWhole(number float64) bool {
	return !math.IsNaN(number) && !math.IsInf(number, 0) && number == math.Trunc(number)
}

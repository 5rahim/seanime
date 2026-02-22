package util

import (
	"math"
	"strconv"
	"strings"
)

func StringToInt(str string) (int, bool) {
	dotIndex := strings.IndexByte(str, '.')
	if dotIndex != -1 {
		str = str[:dotIndex]
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0, false
	}
	return i, true
}

func StringToIntMust(str string) int {
	dotIndex := strings.IndexByte(str, '.')
	if dotIndex != -1 {
		str = str[:dotIndex]
	}
	i, err := strconv.Atoi(str)
	if err != nil {
		return 0
	}
	return i
}

func IntegerToRoman(number int) string {
	maxRomanNumber := 3999
	if number > maxRomanNumber {
		return strconv.Itoa(number)
	}

	conversions := []struct {
		value int
		digit string
	}{
		{1000, "M"},
		{900, "CM"},
		{500, "D"},
		{400, "CD"},
		{100, "C"},
		{90, "XC"},
		{50, "L"},
		{40, "XL"},
		{10, "X"},
		{9, "IX"},
		{5, "V"},
		{4, "IV"},
		{1, "I"},
	}

	var roman strings.Builder
	for _, conversion := range conversions {
		for number >= conversion.value {
			roman.WriteString(conversion.digit)
			number -= conversion.value
		}
	}

	return roman.String()
}

func getOrdinalSuffix(number int) string {
	absNumber := int(math.Abs(float64(number)))

	i := absNumber % 100
	if i == 11 || i == 12 || i == 13 {
		return "th"
	}

	switch absNumber % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func IntegerToOrdinal(number int) string {
	return strconv.Itoa(number) + getOrdinalSuffix(number)
}

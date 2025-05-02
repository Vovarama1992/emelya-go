package utils

import (
	"strings"
	"unicode"
)

func NormalizePhone(phone string) string {
	var digits strings.Builder

	for _, r := range phone {
		if unicode.IsDigit(r) {
			digits.WriteRune(r)
		}
	}

	result := digits.String()

	if strings.HasPrefix(result, "8") && len(result) == 11 {
		result = "7" + result[1:]
	} else if strings.HasPrefix(result, "9") && len(result) == 10 {
		result = "7" + result
	}

	return "+7" + result[1:]
}

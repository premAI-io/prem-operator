package utils

import (
	"strings"
	"unicode"
)

func IsAlphanumeric(str string) bool {
	for _, char := range str {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func ToHostName(name string) string {
	s := strings.ReplaceAll(name, "_", "-")
	s = strings.ReplaceAll(s, ".", "-")

	return strings.ToLower(s)
}

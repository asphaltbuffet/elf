// Package utilities contains helper functions used by the application.
package utilities

import (
	"unicode"
)

// CamelToTitle converts a camelCase string to a title case string.
func CamelToTitle(x string) string {
	var out string

	for i, char := range x {
		switch {
		case i == 0:
			out += string(unicode.ToUpper(char))
		case unicode.IsUpper(char) && unicode.IsLetter(rune(x[i-1])):
			out += " " + string(char)
		default:
			out += string(char)
		}
	}

	return out
}

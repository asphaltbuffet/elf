// Package utilities contains helper functions used by the application.
package utilities

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
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

func ToCamel(s string) string {
	// Remove all characters that are not alphanumeric or spaces or underscores
	s = regexp.MustCompile("[^a-zA-Z0-9'-_ ]+").ReplaceAllString(s, "")

	// Replace all underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")

	// Title case s
	s = cases.Title(language.AmericanEnglish, cases.NoLower).String(s)

	// Remove all spaces
	s = strings.ReplaceAll(s, " ", "")

	// Lowercase the first letter
	if len(s) > 0 {
		s = strings.ToLower(s[:1]) + s[1:]
	}

	return s
}

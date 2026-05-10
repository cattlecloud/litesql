package litesql

import (
	"strings"
	"unicode"
)

// SanitizeFTS5 removes SQLite FTS5 control characters (*, :, ^, ", etc.)
// by retaining only alphanumeric characters and basic whitespace.
func SanitizeFTS5(query string) string {
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsSpace(r) {
			return r
		}
		return -1 // drop
	}, query)

	// clean up duplicate spaces and trim padding
	return strings.Join(strings.Fields(sanitized), " ")
}

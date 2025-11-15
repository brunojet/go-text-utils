package textutil

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// RemoveConnectors removes common Portuguese connector words (de, do, da, etc.)
// and returns the cleaned, uppercased string (words separated by single spaces).
func RemoveConnectors(s string) string {
	connectors := map[string]bool{
		"DE": true, "DO": true, "DA": true, "DAS": true, "DOS": true, "E": true,
	}
	parts := strings.Fields(strings.ToUpper(s))
	var clean []string
	for _, p := range parts {
		if !connectors[p] {
			clean = append(clean, p)
		}
	}
	return strings.Join(clean, " ")
}

// RemoveAccents uppercases and removes diacritics (combining marks) and preserves digits.
// It returns the cleaned string without spaces (same behavior as the previous implementation
// which concatenated words after removing accents).
func RemoveAccents(s string) string {
	// First normalize NFD to decompose characters and separate diacritics
	t := transform.Chain(norm.NFD)
	decomposed, _, err := transform.String(t, s)
	if err != nil {
		// fallback to original on error
		decomposed = s
	}

	// Remove combining marks and keep only letters and digits, uppercase
	var b strings.Builder
	upper := strings.ToUpper(decomposed)
	for len(upper) > 0 {
		r, size := utf8.DecodeRuneInString(upper)
		upper = upper[size:]
		// skip marks (Mn)
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			// letters are already uppercased via upper variable
			b.WriteRune(r)
		}
	}
	return b.String()
}

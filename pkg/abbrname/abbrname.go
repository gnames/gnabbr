// Package abbrname recognizes names with an abbreviated genus, such as
// `C. elegans` or `Cae. elegans`.
package abbrname

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// AbbrName is a name with an abbreviated genus.
type AbbrName struct {
	// Prefix is the abbreviation without its period: `C`, `Cae`.
	Prefix string

	// Epithet is the specific epithet: `elegans`.
	Epithet string
}

// Parse returns the abbreviated genus and the epithet of a name. The
// abbreviation is a capital letter followed by up to two lowercase
// letters and a period; the epithet starts with a lowercase letter.
// It returns false for any other name.
func Parse(name string) (AbbrName, bool) {
	ws := strings.Fields(name)
	if len(ws) < 2 {
		return AbbrName{}, false
	}
	prefix, ok := strings.CutSuffix(ws[0], ".")
	if !ok || !isPrefix(prefix) || !isEpithet(ws[1]) {
		return AbbrName{}, false
	}
	return AbbrName{Prefix: prefix, Epithet: ws[1]}, true
}

// Genus returns the genus of a full name (`Carex` for `Carex alba L.`),
// or an empty string when the name does not start with a genus.
func Genus(name string) string {
	ws := strings.Fields(name)
	if len(ws) == 0 {
		return ""
	}
	g := ws[0]
	for i, r := range g {
		if (i == 0 && !unicode.IsUpper(r)) || (i > 0 && !unicode.IsLower(r)) {
			return ""
		}
	}
	if utf8.RuneCountInString(g) < 2 {
		return ""
	}
	return g
}

func isPrefix(s string) bool {
	n := utf8.RuneCountInString(s)
	if n == 0 || n > 3 {
		return false
	}
	for i, r := range s {
		if (i == 0 && !unicode.IsUpper(r)) || (i > 0 && !unicode.IsLower(r)) {
			return false
		}
	}
	return true
}

func isEpithet(s string) bool {
	r, _ := utf8.DecodeRuneInString(s)
	return unicode.IsLower(r)
}

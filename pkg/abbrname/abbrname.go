// Package abbrname recognizes names with an abbreviated genus, such as
// `C. elegans` or `Cae. elegans`.
package abbrname

import (
	"slices"
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

// Readings returns possible readings of an abbreviated name whose
// abbreviation has two capitals in the text (`CS. ambigua`, normalized to
// `Cs. ambigua`), or nil for any other name. Such an abbreviation is
// usually an OCR error: a stray character before the real one (`(` read
// as `C`, `J`, `I`), a doubled letter (`LL.`), or a two-letter
// abbreviation with a wrong case (`PL.`, `CI.` for `Cl.`). The readings
// are, in order: without the first letter (`S. ambigua`), the first
// letter only (`C. ambigua`), `l` for `I` as the second letter
// (`Cl. ambigua` for `CI.`), and the name as it is.
func Readings(verbatim, name string) []string {
	v := strings.TrimLeftFunc(verbatim, func(r rune) bool { return !unicode.IsLetter(r) })
	ws := strings.Fields(v)
	if len(ws) < 2 {
		return nil
	}
	prefix, ok := strings.CutSuffix(ws[0], ".")
	rs := []rune(prefix)
	if !ok || len(rs) < 2 || len(rs) > 3 ||
		!unicode.IsUpper(rs[0]) || !unicode.IsUpper(rs[1]) ||
		(len(rs) == 3 && !unicode.IsLower(rs[2])) {
		return nil
	}
	a, ok := Parse(name)
	if !ok {
		return nil
	}
	rest := strings.Join(strings.Fields(name)[1:], " ")
	prefixes := []string{string(rs[1:]), string(rs[0])}
	if rs[1] == 'I' {
		prefixes = append(prefixes, string(rs[0])+"l"+string(rs[2:]))
	}
	prefixes = append(prefixes, a.Prefix)

	var res []string
	for _, p := range prefixes {
		if !isPrefix(p) {
			continue
		}
		n := p + ". " + rest
		if !slices.Contains(res, n) {
			res = append(res, n)
		}
	}
	return res
}

// Expand replaces the abbreviated genus of a name with genus and keeps
// the rest as written: `Carex alba var. nigra` for `C. alba var. nigra`.
func Expand(name, genus string) string {
	ws := strings.Fields(name)
	if len(ws) == 0 {
		return genus
	}
	return strings.Join(append([]string{genus}, ws[1:]...), " ")
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

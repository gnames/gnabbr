// Package hints finds genera that an abbreviation may stand for: genera
// of full names mentioned before it nearby in the same scope.
//
// Names carry no rank, so a uninomial hint may be a higher taxon
// (`Coleoptera` before `C. auratus`); gnames abbr_names sorts out which
// hints are genera.
package hints

import (
	"slices"
	"strings"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/doc"
)

// New returns genus hints for each abbreviated name of a scope, by its
// index in names. Scope holds indexes of the scope's names, sorted by
// Start. Hints are distinct genera that start with the abbreviation,
// taken from full names (multinomials and uninomials) earlier in the
// scope and at most window runes back, nearest first, at most limit of
// them.
func New(names []doc.Name, scope []int, window, limit int) map[int][]string {
	type mention struct {
		genus string
		start int
	}
	res := make(map[int][]string)
	var seen []mention
	for _, i := range scope {
		n := names[i]
		if a, ok := abbrname.Parse(n.Name); ok {
			var hs []string
			for j := len(seen) - 1; j >= 0 && len(hs) < limit; j-- {
				m := seen[j]
				if n.Start-m.start > window {
					break
				}
				if strings.HasPrefix(m.genus, a.Prefix) && !slices.Contains(hs, m.genus) {
					hs = append(hs, m.genus)
				}
			}
			if len(hs) > 0 {
				res[i] = hs
			}
			continue
		}
		if g := abbrname.Genus(n.Name); g != "" {
			seen = append(seen, mention{genus: g, start: n.Start})
		}
	}
	return res
}

// Merge joins lists of hints, nearest first: a genus takes its best
// position in any list, ties keep the order of the lists. It returns at
// most limit hints.
func Merge(limit int, lists ...[]string) []string {
	type hint struct {
		genus     string
		pos, list int
	}
	var hs []hint
	idx := make(map[string]int)
	for l, list := range lists {
		for p, g := range list {
			if k, ok := idx[g]; ok {
				if p < hs[k].pos {
					hs[k].pos = p
				}
				continue
			}
			idx[g] = len(hs)
			hs = append(hs, hint{genus: g, pos: p, list: l})
		}
	}
	slices.SortStableFunc(hs, func(a, b hint) int {
		if a.pos != b.pos {
			return a.pos - b.pos
		}
		return a.list - b.list
	})
	var res []string
	for _, h := range hs[:min(limit, len(hs))] {
		res = append(res, h.genus)
	}
	return res
}

// Extend appends hints of other names of a group to the hints of one
// name, so a hint before the first name of a group helps all its names.
// It returns at most limit hints.
func Extend(own, group []string, limit int) []string {
	res := slices.Clone(own)
	for _, g := range group {
		if len(res) >= limit {
			break
		}
		if !slices.Contains(res, g) {
			res = append(res, g)
		}
	}
	return res[:min(limit, len(res))]
}

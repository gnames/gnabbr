// Package hints finds names that hint at the genus an abbreviation stands
// for: full names mentioned before it nearby in the same scope.
//
// A hint is a full name (`Carex nigra`), which shows that the text
// mentions the genus, or a genus (`Carex`) for a uninomial. gnames
// abbr_names counts hint positions by distinct genera and weighs a
// name more than a genus, and a name with the abbreviation's epithet
// most (`Carex alba` for `C. alba`).
//
// Names carry no rank, so a uninomial hint may be a higher taxon
// (`Coleoptera` before `C. auratus`); abbr_names sorts out which
// one-word hints are genera.
package hints

import (
	"slices"
	"strings"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/doc"
)

// New returns hints for each abbreviated name of a scope, by its index in
// names. Scope holds indexes of the scope's names, sorted by Start.
// Hints come from full names (multinomials and uninomials) earlier in
// the scope and at most window runes back, with genera that start with
// the abbreviation, nearest first. Each of the nearest limit genera
// gives its nearest name, and its names with the epithet of the
// abbreviated name.
func New(names []doc.Name, scope []int, window, limit int) map[int][]string {
	type mention struct {
		genus, hint, epithet string
		start                int
	}
	res := make(map[int][]string)
	var seen []mention
	for _, i := range scope {
		n := names[i]
		if a, ok := abbrname.Parse(n.Name); ok {
			var hs, genera []string
			for j := len(seen) - 1; j >= 0; j-- {
				m := seen[j]
				if n.Start-m.start > window {
					break
				}
				if !strings.HasPrefix(m.genus, a.Prefix) || slices.Contains(hs, m.hint) {
					continue
				}
				switch {
				case !slices.Contains(genera, m.genus):
					if len(genera) == limit {
						continue
					}
					genera = append(genera, m.genus)
					hs = append(hs, m.hint)
				case m.epithet == a.Epithet:
					hs = append(hs, m.hint)
				}
			}
			if len(hs) > 0 {
				res[i] = hs
			}
			continue
		}
		if g := abbrname.Genus(n.Name); g != "" {
			m := mention{genus: g, hint: g, start: n.Start}
			if ws := strings.Fields(n.Name); len(ws) > 1 {
				m.hint, m.epithet = n.Name, ws[1]
			}
			seen = append(seen, m)
		}
	}
	return res
}

// Genus returns the genus of a hint: its first word.
func Genus(hint string) string {
	g, _, _ := strings.Cut(strings.TrimSpace(hint), " ")
	return g
}

// Genera returns distinct genera of hints, in their order.
func Genera(hints []string) []string {
	var res []string
	for _, h := range hints {
		if g := Genus(h); !slices.Contains(res, g) {
			res = append(res, g)
		}
	}
	return res
}

// Merge joins lists of hints, nearest first: a hint takes its best
// position in any list, ties keep the order of the lists. It keeps hints
// of at most limit genera.
func Merge(limit int, lists ...[]string) []string {
	type hint struct {
		name      string
		pos, list int
	}
	var hs []hint
	idx := make(map[string]int)
	for l, list := range lists {
		for p, h := range list {
			if k, ok := idx[h]; ok {
				if p < hs[k].pos {
					hs[k].pos = p
				}
				continue
			}
			idx[h] = len(hs)
			hs = append(hs, hint{name: h, pos: p, list: l})
		}
	}
	slices.SortStableFunc(hs, func(a, b hint) int {
		if a.pos != b.pos {
			return a.pos - b.pos
		}
		return a.list - b.list
	})
	res := make([]string, 0, len(hs))
	for _, h := range hs {
		res = append(res, h.name)
	}
	return limitGenera(res, limit)
}

// Extend appends hints of other names of a group to the hints of one
// name, so a hint before the first name of a group helps all its names.
// It keeps hints of at most limit genera.
func Extend(own, group []string, limit int) []string {
	res := slices.Clone(own)
	for _, h := range group {
		if !slices.Contains(res, h) {
			res = append(res, h)
		}
	}
	return limitGenera(res, limit)
}

// limitGenera keeps hints of the first limit distinct genera.
func limitGenera(hs []string, limit int) []string {
	var genera []string
	res := hs[:0]
	for _, h := range hs {
		g := Genus(h)
		if !slices.Contains(genera, g) {
			if len(genera) == limit {
				continue
			}
			genera = append(genera, g)
		}
		res = append(res, h)
	}
	if len(res) == 0 {
		return nil
	}
	return res
}

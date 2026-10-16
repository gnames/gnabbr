// Package hints finds genera that an abbreviation may stand for: genera
// of full names mentioned before it nearby in the document.
package hints

import (
	"cmp"
	"slices"
	"strings"

	"github.com/gnames/gnabbr/pkg/ent/abbrname"
	"github.com/gnames/gnabbr/pkg/ent/doc"
)

// New returns genus hints for each abbreviated occurrence, by its index
// in occs. Hints are distinct genera that start with the abbreviation,
// taken from full names earlier in the same part and at most window
// pages back, nearest first, at most limit of them.
func New(occs []doc.Occurrence, window, limit int) map[int][]string {
	order := make([]int, len(occs))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		oa, ob := occs[a], occs[b]
		return cmp.Or(
			cmp.Compare(oa.Part, ob.Part),
			cmp.Compare(oa.Page, ob.Page),
			cmp.Compare(oa.Offset, ob.Offset),
		)
	})

	type mention struct {
		genus string
		page  int
	}
	res := make(map[int][]string)
	var seen []mention
	part := ""
	for n, i := range order {
		o := occs[i]
		if n == 0 || o.Part != part {
			part, seen = o.Part, seen[:0]
		}
		if a, ok := abbrname.Parse(o.Name); ok {
			var hs []string
			for j := len(seen) - 1; j >= 0 && len(hs) < limit; j-- {
				m := seen[j]
				if o.Page-m.page > window {
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
		if g := abbrname.Genus(o.Name); g != "" {
			seen = append(seen, mention{genus: g, page: o.Page})
		}
	}
	return res
}

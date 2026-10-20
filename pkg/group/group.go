// Package group splits abbreviated names of a document into local
// groups: names with the same abbreviation, in the same part, on pages
// close to each other. A group is what gnames abbr_names resolves at
// once, so names in a group are expected to share a genus.
package group

import (
	"cmp"
	"slices"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/doc"
)

// Group is a set of abbreviated names that are resolved together.
type Group struct {
	// Part of the document the group is in.
	Part string

	// Prefix is the abbreviation without its period.
	Prefix string

	// StartPage and EndPage are the first and last pages of the group.
	StartPage, EndPage int

	// Members are indexes of the group's occurrences in the document,
	// in reading order.
	Members []int
}

// New makes groups from the abbreviated names among occurrences. A group
// ends where the next name with its abbreviation is more than gapPages
// pages away, or where the group would span more than spanPages pages.
func New(occs []doc.Occurrence, gapPages, spanPages int) []Group {
	type item struct {
		idx    int
		prefix string
	}
	var items []item
	for i, o := range occs {
		if a, ok := abbrname.Parse(o.Name); ok {
			items = append(items, item{idx: i, prefix: a.Prefix})
		}
	}
	slices.SortStableFunc(items, func(a, b item) int {
		oa, ob := occs[a.idx], occs[b.idx]
		return cmp.Or(
			cmp.Compare(oa.Part, ob.Part),
			cmp.Compare(a.prefix, b.prefix),
			cmp.Compare(oa.Page, ob.Page),
			cmp.Compare(oa.Offset, ob.Offset),
		)
	})

	var res []Group
	var cur *Group
	for _, it := range items {
		o := occs[it.idx]
		if cur == nil || cur.Part != o.Part || cur.Prefix != it.prefix ||
			o.Page-cur.EndPage > gapPages || o.Page-cur.StartPage > spanPages {
			res = append(res, Group{
				Part: o.Part, Prefix: it.prefix,
				StartPage: o.Page, EndPage: o.Page,
			})
			cur = &res[len(res)-1]
		}
		cur.EndPage = o.Page
		cur.Members = append(cur.Members, it.idx)
	}
	return res
}

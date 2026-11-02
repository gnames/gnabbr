// Package group splits abbreviated names of one scope (a section or the
// whole document) into local groups: names with the same abbreviation
// close to each other. A group is what gnames abbr_names reconciles at
// once, so names in a group are expected to share a genus.
package group

import (
	"cmp"
	"slices"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/doc"
)

// Group is a set of abbreviated names that are reconciled together.
type Group struct {
	// Section is the index of the group's section in Document.Sections,
	// -1 for the document scope.
	Section int

	// Prefix is the abbreviation without its period.
	Prefix string

	// Start and End are the offsets of the start of the first and the last name.
	Start, End int

	// Members are indexes of the group's names in Document.Names, by
	// Start.
	Members []int
}

// New makes groups from the abbreviated names of a scope. Scope holds
// indexes of the scope's names in names, sorted by Start. A group ends
// where the next name with its abbreviation is more than gapChars runes
// away, or where the group would span more than spanChars runes.
func New(
	names []doc.Name,
	scope []int,
	section, gapChars, spanChars int,
) []Group {
	type item struct {
		idx    int
		prefix string
	}
	var items []item
	for _, i := range scope {
		if a, ok := abbrname.Parse(names[i].Name); ok {
			items = append(items, item{idx: i, prefix: a.Prefix})
		}
	}
	// Stable: names with the same prefix stay sorted by Start.
	slices.SortStableFunc(items, func(a, b item) int {
		return cmp.Compare(a.prefix, b.prefix)
	})

	var res []Group
	var cur *Group
	for _, it := range items {
		start := names[it.idx].Start
		if cur == nil || cur.Prefix != it.prefix ||
			start-cur.End > gapChars || start-cur.Start > spanChars {
			res = append(res, Group{
				Section: section, Prefix: it.prefix, Start: start, End: start,
			})
			cur = &res[len(res)-1]
		}
		cur.End = start
		cur.Members = append(cur.Members, it.idx)
	}
	return res
}

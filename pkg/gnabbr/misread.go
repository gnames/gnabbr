package gnabbr

import (
	"context"
	"fmt"
	"slices"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/hints"
	"github.com/gnames/gnlib/ent/abbr"
)

// reread replaces names whose abbreviation has two capitals in the text
// (`CS. ambigua`) with their best reading (see abbrname.Readings), so
// they are grouped with the names of the abbreviation they stand for.
//
// Each reading is sent to abbr_names alone, with the hints it gets in
// its scope, and the reading whose best genus has the highest score
// wins; an earlier reading wins a tie. A name in several scopes takes
// the best reading of all of them. A name no reading of which gets a
// genus stays as it is.
func (g *gnabbr) reread(
	ctx context.Context,
	d doc.Document,
	names []doc.Name,
	scs []scope,
) error {
	readings := make(map[int][]string)
	var most int
	for i, n := range names {
		if rs := abbrname.Readings(n.Verbatim, n.Name); len(rs) > 1 {
			readings[i] = rs
			most = max(most, len(rs))
		}
	}
	if len(readings) == 0 {
		return nil
	}

	// variants[k] has the k-th reading of every misread name that has
	// one, so hints of a reading come from the scope as for any name.
	variants := make([][]doc.Name, most)
	for k := range variants {
		variants[k] = slices.Clone(names)
		for i, rs := range readings {
			if k < len(rs) {
				variants[k][i].Name = rs[k]
			}
		}
	}

	type best struct {
		name  string
		score float64
	}
	bests := make(map[int]best)
	for _, sc := range scs {
		var misread []int
		for _, i := range sc.names {
			if _, ok := readings[i]; ok {
				misread = append(misread, i)
			}
		}
		if len(misread) == 0 {
			continue
		}
		st := d.TaxonStats
		if sc.section >= 0 && d.Sections[sc.section].TaxonStats != nil {
			st = *d.Sections[sc.section].TaxonStats
		}
		for k, vs := range variants {
			hs := hints.New(vs, sc.names, g.cfg.HintWindow, g.cfg.HintsLimit)
			for _, i := range misread {
				if k >= len(readings[i]) {
					continue
				}
				inp := abbr.Input{
					Names: []abbr.Name{{
						ID: "0", Name: vs[i].Name, GenusHints: hs[i],
					}},
					MainTaxon:     st.MainTaxon.Name,
					DominantClass: st.Class.Name,
					GeneraLimit:   1,
				}
				out, err := g.client.AbbrNames(ctx, inp)
				if err != nil {
					return err
				}
				if out.Error != "" {
					return fmt.Errorf("abbr_names for %q: %s", vs[i].Name, out.Error)
				}
				if len(out.Genera) == 0 {
					continue
				}
				s := out.Genera[0].Score.Total
				if b, ok := bests[i]; !ok || s > b.score {
					bests[i] = best{name: vs[i].Name, score: s}
				}
			}
		}
	}
	for i, b := range bests {
		names[i].Name = b.name
	}
	return nil
}

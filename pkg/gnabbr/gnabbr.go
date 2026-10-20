// Package gnabbr reconciles abbreviated genera of scientific names
// (`C. elegans` -> Caenorhabditis elegans) found in documents, using
// genus hints, taxonomic statistics of the document and gnames
// abbr_names.
package gnabbr

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/group"
	"github.com/gnames/gnabbr/pkg/hints"
	"github.com/gnames/gnlib/ent/abbr"
	"github.com/gnames/gnstats/ent/stats"
)

type gnabbr struct {
	cfg    *config.Config
	client AbbrNamer
}

// New creates GNabbr with a given abbr_names implementation. Callers
// use github.com/gnames/gnabbr.New, which gives it the gnames client.
func New(cfg *config.Config, client AbbrNamer) GNabbr {
	return &gnabbr{cfg: cfg, client: client}
}

// candidate is a group with the answer of abbr_names for it.
type candidate struct {
	index    int
	group    group.Group
	stats    TaxonStats
	hints    []string
	details  Details
	genus    *abbr.Genus
	accepted bool
}

// Resolve implements GNabbr.
func (g *gnabbr) Resolve(ctx context.Context, d doc.Document) (Output, error) {
	start := time.Now()
	names := slices.Clone(d.Names)
	for i := range names {
		if names[i].ID == "" {
			names[i].ID = strconv.Itoa(i)
		}
	}

	var cands []candidate
	for _, sc := range scopes(names, d.Sections) {
		sec, scope := sc.section, sc.names
		st := d.TaxonStats
		if sec >= 0 && d.Sections[sec].TaxonStats != nil {
			st = *d.Sections[sec].TaxonStats
		}
		hs := hints.New(names, scope, g.cfg.HintWindow, g.cfg.HintsLimit)
		gs := group.New(names, scope, sec, g.cfg.GapChars, g.cfg.SpanChars)
		for _, gr := range gs {
			c, err := g.reconcile(ctx, names, gr, hs, st)
			if err != nil {
				return Output{}, err
			}
			c.index = len(cands)
			cands = append(cands, c)
		}
	}

	res := g.output(d, names, cands)
	res.TotalSec = float32(time.Since(start).Seconds())
	return res, nil
}

// scope holds indexes of names of a section, sorted by Start.
type scope struct {
	// section is the index of the section, -1 for names outside of all
	// sections (all names, if there are no sections).
	section int
	names   []int
}

// scopes returns the scopes of the sections, in their order, and the
// scope of the document last. Sections may overlap: a name is in each
// section that covers it.
func scopes(names []doc.Name, secs []doc.Section) []scope {
	order := make([]int, len(names))
	for i := range order {
		order[i] = i
	}
	slices.SortStableFunc(order, func(a, b int) int {
		return cmp.Compare(names[a].Start, names[b].Start)
	})

	res := make([]scope, len(secs)+1)
	for s := range secs {
		res[s].section = s
	}
	res[len(secs)].section = -1
	for _, i := range order {
		in := false
		for s, sec := range secs {
			if names[i].Start >= sec.Start && names[i].Start < sec.End {
				res[s].names = append(res[s].names, i)
				in = true
			}
		}
		if !in {
			res[len(secs)].names = append(res[len(secs)].names, i)
		}
	}
	return res
}

// reconcile sends a group to abbr_names. Each distinct name string goes
// once, with hints of all its occurrences and of the rest of the group.
func (g *gnabbr) reconcile(
	ctx context.Context,
	names []doc.Name,
	gr group.Group,
	hs map[int][]string,
	st stats.Stats,
) (candidate, error) {
	limit := g.cfg.HintsLimit
	var uniq []string
	own := make(map[string][][]string)
	var all [][]string
	for _, i := range gr.Members {
		n := names[i].Name
		if _, ok := own[n]; !ok {
			uniq = append(uniq, n)
			own[n] = nil
		}
		own[n] = append(own[n], hs[i])
		all = append(all, hs[i])
	}
	groupHints := hints.Merge(limit, all...)

	res := candidate{
		group: gr,
		stats: TaxonStats{
			MainTaxon:     st.MainTaxon.Name,
			DominantClass: st.Class.Name,
		},
		hints: groupHints,
	}
	inp := abbr.Input{
		Names:         make([]abbr.Name, len(uniq)),
		MainTaxon:     res.stats.MainTaxon,
		DominantClass: res.stats.DominantClass,
		GeneraLimit:   g.cfg.GeneraLimit,
		WithFeatures:  g.cfg.WithFeatures,
	}
	for j, n := range uniq {
		inp.Names[j] = abbr.Name{
			ID:         strconv.Itoa(j),
			Name:       n,
			GenusHints: hints.Extend(hints.Merge(limit, own[n]...), groupHints, limit),
		}
	}

	out, err := g.client.AbbrNames(ctx, inp)
	if err != nil {
		return res, err
	}
	if out.Error != "" {
		return res, fmt.Errorf("abbr_names for %q: %s", gr.Prefix, out.Error)
	}
	res.details = Details{Input: inp, Output: out}
	// Genera come best first.
	if len(out.Genera) > 0 {
		res.genus = &out.Genera[0]
		res.accepted = res.genus.Score.Total >= g.cfg.Threshold
	}
	return res, nil
}

// output makes reconciled and failed groups. A name in several sections
// is in several groups: it is expanded by the accepted group with the
// best score and is left out of the others.
func (g *gnabbr) output(d doc.Document, names []doc.Name, cands []candidate) Output {
	best := make(map[int]int)
	for ci, c := range cands {
		if !c.accepted {
			continue
		}
		for _, i := range c.group.Members {
			if b, ok := best[i]; !ok || c.genus.Score.Total > cands[b].genus.Score.Total {
				best[i] = ci
			}
		}
	}

	res := Output{
		Meta: Meta{
			Version:    Version,
			DocumentID: d.ID,
			Settings: Settings{
				GapChars:    g.cfg.GapChars,
				SpanChars:   g.cfg.SpanChars,
				HintWindow:  g.cfg.HintWindow,
				HintsLimit:  g.cfg.HintsLimit,
				GeneraLimit: g.cfg.GeneraLimit,
				Threshold:   g.cfg.Threshold,
			},
			NamesNum: len(names),
		},
		Reconciled: []Group{},
		Failed:     []Group{},
	}
	abbrs := make(map[int]struct{})
	for ci, c := range cands {
		gr := g.group(d, c)
		for _, i := range c.group.Members {
			abbrs[i] = struct{}{}
			n := names[i]
			b, ok := best[i]
			switch {
			case c.accepted && ok && b == ci:
				gr.Names = append(gr.Names, Name{
					ID:       n.ID,
					Verbatim: n.Verbatim,
					Name:     abbrname.Expand(n.Name, c.genus.Name),
					AbbrName: n.Name,
					Start:    n.Start,
				})
			case !c.accepted && !ok:
				v := cmp.Or(n.Verbatim, n.Name)
				if !slices.Contains(gr.Verbatims, v) {
					gr.Verbatims = append(gr.Verbatims, v)
				}
			}
		}
		if len(gr.Names) > 0 {
			res.Reconciled = append(res.Reconciled, gr)
			res.ReconciledNum += len(gr.Names)
		} else if len(gr.Verbatims) > 0 {
			res.Failed = append(res.Failed, gr)
		}
	}
	res.AbbrNamesNum = len(abbrs)
	return res
}

func (g *gnabbr) group(d doc.Document, c candidate) Group {
	res := Group{
		Index:      c.index,
		Abbr:       c.group.Prefix,
		Start:      c.group.Start,
		End:        c.group.End,
		TaxonStats: c.stats,
		Hints:      c.hints,
		Genus:      c.genus,
	}
	if s := c.group.Section; s >= 0 {
		res.Section = cmp.Or(d.Sections[s].ID, strconv.Itoa(s))
	}
	if gs := c.details.Output.Genera; len(gs) > 1 {
		res.Candidates = gs[1:]
	}
	if g.cfg.WithDetails {
		det := c.details
		res.Details = &det
	}
	return res
}

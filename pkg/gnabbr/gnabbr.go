// Package gnabbr resolves abbreviated genera of scientific names
// (`C. elegans` -> Caenorhabditis) found in documents, using the
// context of the document and gnames abbr_names.
package gnabbr

import (
	"context"

	"github.com/gnames/gnabbr/internal/io/abbrrest"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/ent/doc"
	"github.com/gnames/gnabbr/pkg/ent/group"
	"github.com/gnames/gnabbr/pkg/ent/hints"
	"github.com/gnames/gnlib/ent/abbr"
)

type gnabbr struct {
	cfg    *config.Config
	client AbbrNamer
}

// New creates GNabbr that calls abbr_names at cfg.GnamesURL.
func New(cfg *config.Config) GNabbr {
	return NewWithClient(cfg, abbrrest.New(cfg.GnamesURL))
}

// NewWithClient creates GNabbr with a given abbr_names implementation.
func NewWithClient(cfg *config.Config, client AbbrNamer) GNabbr {
	return &gnabbr{cfg: cfg, client: client}
}

// Resolve implements GNabbr.
func (g *gnabbr) Resolve(ctx context.Context, d doc.Document) (Output, error) {
	occs := d.Occurrences
	gs := group.New(occs, g.cfg.GapPages, g.cfg.SpanPages)
	hs := hints.New(occs, g.cfg.HintWindow, g.cfg.HintsLimit)

	res := Output{DocumentID: d.ID, Groups: make([]GroupOutput, 0, len(gs))}
	byID := make(map[string]*Result)
	results := make(map[int]*Result)
	for gi, gr := range gs {
		inp := g.input(d, gr, hs)
		out, err := g.client.AbbrNames(ctx, inp)
		if err != nil {
			return res, err
		}
		res.Groups = append(res.Groups, GroupOutput{Group: gr, Input: inp, Output: out})

		clear(byID)
		for _, i := range gr.Members {
			r := &Result{OccurrenceID: occs[i].ID, Name: occs[i].Name, Group: gi}
			results[i] = r
			byID[r.OccurrenceID] = r
		}
		// Genera come best first: the first genus matching a name wins.
		for _, gen := range out.Genera {
			for _, m := range gen.MatchedNames {
				if r, ok := byID[m.ID]; ok && r.Genus == "" {
					r.Genus = gen.Name
					r.Score = gen.Score.Total
					r.Accepted = gen.Score.Total >= g.cfg.Threshold
				}
			}
		}
	}

	for i := range occs {
		if r, ok := results[i]; ok {
			res.Results = append(res.Results, *r)
		}
	}
	return res, nil
}

func (g *gnabbr) input(d doc.Document, gr group.Group, hs map[int][]string) abbr.Input {
	ctx, ok := d.Contexts[gr.Part]
	if !ok {
		ctx = d.Contexts[""]
	}
	names := make([]abbr.Name, len(gr.Members))
	for j, i := range gr.Members {
		o := d.Occurrences[i]
		names[j] = abbr.Name{ID: o.ID, Name: o.Name, GenusHints: hs[i]}
	}
	return abbr.Input{
		Names:         names,
		MainTaxon:     ctx.MainTaxon,
		DominantClass: ctx.DominantClass,
		GeneraLimit:   g.cfg.GeneraLimit,
		WithFeatures:  g.cfg.WithFeatures,
	}
}

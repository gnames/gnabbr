package gnabbr_test

import (
	"context"
	"testing"

	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/gnabbr"
	"github.com/gnames/gnlib/ent/abbr"
	"github.com/gnames/gnstats/ent/stats"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNamer answers with the genus of the first hint of the group's
// first name, scored by the number of names, or with no genera if there
// are no hints.
type fakeNamer struct {
	inputs []abbr.Input
}

func (f *fakeNamer) AbbrNames(_ context.Context, inp abbr.Input) (abbr.Output, error) {
	f.inputs = append(f.inputs, inp)
	hs := inp.Names[0].GenusHints
	if len(hs) == 0 {
		return abbr.Output{}, nil
	}
	gen := abbr.Genus{Name: hs[0], Score: abbr.Score{Total: float64(len(inp.Names))}}
	for _, n := range inp.Names {
		gen.MatchedNames = append(gen.MatchedNames, abbr.MatchedName{ID: n.ID, Name: n.Name})
	}
	return abbr.Output{Genera: []abbr.Genus{gen, {Name: "Cyperus"}}}, nil
}

func taxonStats(mainTaxon, class string) stats.Stats {
	return stats.Stats{
		MainTaxon: stats.Taxon{Name: mainTaxon},
		Class:     stats.Taxon{Name: class},
	}
}

func TestResolve(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	cfg.Update([]config.Option{config.OptThreshold(2)})
	f := &fakeNamer{}
	g := gnabbr.New(cfg, f)

	p2 := taxonStats("Rhabditida", "")
	d := doc.Document{
		ID: "item-1",
		Names: []doc.Name{
			{Name: "C. elegans", Verbatim: "C. elegans,", Start: 20},
			{Name: "Caenorhabditis elegans", Start: 0},
			{Name: "C. briggsae", Start: 40},
			{Name: "C. elegans", Start: 60},
			{Name: "C. remanei", Start: 1020},
			{Name: "Ca. alba", Start: 1040},
		},
		Sections:   []doc.Section{{ID: "p2", Start: 1000, End: 2000, TaxonStats: &p2}},
		TaxonStats: taxonStats("Nematoda", "Chromadorea"),
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)
	assert.Equal("item-1", out.DocumentID)
	assert.Equal(6, out.NamesNum)
	assert.Equal(5, out.AbbrNamesNum)
	assert.Equal(3, out.ReconciledNum)

	// The section comes first: two groups without hints fail.
	require.Len(t, f.inputs, 3)
	assert.Equal("Rhabditida", f.inputs[0].MainTaxon)
	assert.Equal("Chromadorea", f.inputs[2].DominantClass)

	// A repeated name goes to abbr_names once.
	in := f.inputs[2]
	require.Len(t, in.Names, 2)
	assert.Equal("C. elegans", in.Names[0].Name)
	assert.Equal("C. briggsae", in.Names[1].Name)
	assert.Equal([]string{"Caenorhabditis"}, in.Names[1].GenusHints)

	require.Len(t, out.Reconciled, 1)
	r := out.Reconciled[0]
	assert.Empty(r.Section)
	assert.Equal("C", r.Abbr)
	assert.Equal("Caenorhabditis", r.Genus.Name)
	assert.Len(r.Candidates, 1)
	assert.Nil(r.Details)
	require.Len(t, r.Names, 3)
	assert.Equal(gnabbr.Name{
		ID: "0", Verbatim: "C. elegans,", Name: "Caenorhabditis elegans",
		AbbrName: "C. elegans", Start: 20,
	}, r.Names[0])
	assert.Equal("Caenorhabditis briggsae", r.Names[1].Name)
	assert.Equal("3", r.Names[2].ID)

	require.Len(t, out.Failed, 2)
	assert.Equal("p2", out.Failed[0].Section)
	assert.Nil(out.Failed[0].Genus)
	assert.Equal([]string{"C. remanei"}, out.Failed[0].Verbatims)
	assert.Equal([]string{"Ca. alba"}, out.Failed[1].Verbatims)
}

// A name in two overlapping sections is expanded by one group only: the
// one with the better score, the first one on ties (`C. nigra`).
func TestResolveOverlap(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	cfg.Update([]config.Option{config.OptWithDetails(true)})
	g := gnabbr.New(cfg, &fakeNamer{})

	d := doc.Document{
		Names: []doc.Name{
			{Name: "Carex alba", Start: 0},
			{Name: "C. alba", Start: 100},
			{Name: "C. nigra", Start: 200},
			{Name: "Cyperus fuscus", Start: 250},
			{Name: "C. fuscus", Start: 300},
		},
		Sections: []doc.Section{
			{ID: "a", Start: 0, End: 250},
			{ID: "b", Start: 200, End: 400},
		},
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)
	require.Len(t, out.Reconciled, 2)
	assert.Empty(out.Failed)

	a, b := out.Reconciled[0], out.Reconciled[1]
	assert.Equal("a", a.Section)
	assert.Equal("Carex", a.Genus.Name)
	require.Len(t, a.Names, 2)
	assert.Equal("Carex nigra", a.Names[1].Name)
	assert.NotNil(a.Details)

	assert.Equal("b", b.Section)
	assert.Equal("Cyperus", b.Genus.Name)
	require.Len(t, b.Names, 1)
	assert.Equal("Cyperus fuscus", b.Names[0].Name)
}

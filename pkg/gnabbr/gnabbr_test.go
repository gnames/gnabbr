package gnabbr_test

import (
	"context"
	"strings"
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

// mixedNamer matches a name with the genus of its first hint; names with
// `qqq` match nothing. With the default tolerance it returns a genus
// only if it matches all names; with tolerance 100 it returns every
// genus with its names.
type mixedNamer struct {
	inputs []abbr.Input
}

func (f *mixedNamer) AbbrNames(_ context.Context, inp abbr.Input) (abbr.Output, error) {
	f.inputs = append(f.inputs, inp)
	var order []string
	matched := make(map[string][]abbr.MatchedName)
	for _, n := range inp.Names {
		if len(n.GenusHints) == 0 || strings.Contains(n.Name, "qqq") {
			continue
		}
		g := n.GenusHints[0]
		if _, ok := matched[g]; !ok {
			order = append(order, g)
		}
		matched[g] = append(matched[g], abbr.MatchedName{ID: n.ID, Name: n.Name})
	}
	var res abbr.Output
	for _, g := range order {
		ms := matched[g]
		if inp.Tolerance == nil && len(ms) < len(inp.Names) {
			continue
		}
		res.Genera = append(res.Genera, abbr.Genus{
			Name: g, Score: abbr.Score{Total: float64(len(ms))}, MatchedNames: ms,
		})
	}
	return res, nil
}

func TestResolveMixed(t *testing.T) {
	assert := assert.New(t)
	f := &mixedNamer{}
	g := gnabbr.New(config.New(), f)

	d := doc.Document{
		Names: []doc.Name{
			{Name: "Gastrophilus intestinalis", Start: 0},
			{Name: "G. nasalis", Start: 100},
			{Name: "G. pecorum", Start: 200},
			{Name: "Glossina palpalis", Start: 300},
			{Name: "G. morsitans", Start: 400},
			{Name: "G. qqq", Verbatim: "G. qqq,", Start: 450},
			{Name: "G. palpalis", Start: 500},
		},
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)

	// The group, the group with full tolerance, and two runs.
	require.Len(t, f.inputs, 4)
	assert.Nil(f.inputs[0].Tolerance)
	assert.Equal(100, *f.inputs[1].Tolerance)

	require.Len(t, out.Reconciled, 2)
	a, b := out.Reconciled[0], out.Reconciled[1]
	assert.True(a.Split)
	assert.Equal("Gastrophilus", a.Genus.Name)
	assert.Equal(100, a.Start)
	assert.Equal(200, a.End)
	require.Len(t, a.Names, 2)
	assert.Equal("Gastrophilus nasalis", a.Names[0].Name)
	assert.Equal("Glossina", b.Genus.Name)
	require.Len(t, b.Names, 2)
	assert.Equal("Glossina morsitans", b.Names[0].Name)
	assert.Equal(4, out.ReconciledNum)

	// The name left out of the runs stays in the failed group.
	require.Len(t, out.Failed, 1)
	assert.False(out.Failed[0].Split)
	assert.Equal([]string{"G. qqq,"}, out.Failed[0].Verbatims)
}

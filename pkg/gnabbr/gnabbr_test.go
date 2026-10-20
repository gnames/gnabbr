package gnabbr_test

import (
	"cmp"
	"context"
	"slices"
	"strings"
	"testing"

	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/gnabbr"
	"github.com/gnames/gnabbr/pkg/hints"
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
	gen := abbr.Genus{Name: hints.Genus(hs[0]), Score: abbr.Score{Total: float64(len(inp.Names))}}
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

	// The section comes first: two groups without hints fail. The last
	// group has two distinct names, so it is also checked against hints
	// with full tolerance.
	require.Len(t, f.inputs, 4)
	assert.Equal(100, *f.inputs[3].Tolerance)
	assert.Equal("Rhabditida", f.inputs[0].MainTaxon)
	assert.Equal("Chromadorea", f.inputs[2].DominantClass)

	// A repeated name goes to abbr_names once.
	in := f.inputs[2]
	require.Len(t, in.Names, 2)
	assert.Equal("C. elegans", in.Names[0].Name)
	assert.Equal("C. briggsae", in.Names[1].Name)
	assert.Equal([]string{"Caenorhabditis elegans"}, in.Names[1].GenusHints)

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
		g := hints.Genus(n.GenusHints[0])
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

// tableNamer matches names with genera of a table by name. A genus
// scores the number of names it matched, plus 0.5/position for each
// matched name that has it among hints, as abbr_names weighs hints.
// With the default tolerance it returns only genera that matched all
// names.
type tableNamer struct {
	genera map[string][]string
}

func (f *tableNamer) AbbrNames(_ context.Context, inp abbr.Input) (abbr.Output, error) {
	var order []string
	matched := make(map[string][]abbr.MatchedName)
	score := make(map[string]float64)
	for _, n := range inp.Names {
		for _, g := range f.genera[n.Name] {
			if _, ok := matched[g]; !ok {
				order = append(order, g)
			}
			matched[g] = append(matched[g], abbr.MatchedName{ID: n.ID, Name: n.Name})
			score[g]++
			if pos := slices.Index(hints.Genera(n.GenusHints), g); pos >= 0 {
				score[g] += 0.5 / float64(pos+1)
			}
		}
	}
	var res abbr.Output
	for _, g := range order {
		ms := matched[g]
		if inp.Tolerance == nil && len(ms) < len(inp.Names) {
			continue
		}
		res.Genera = append(res.Genera, abbr.Genus{
			Name: g, Score: abbr.Score{Total: score[g]}, MatchedNames: ms,
		})
	}
	slices.SortStableFunc(res.Genera, func(a, b abbr.Genus) int {
		return cmp.Compare(b.Score.Total, a.Score.Total)
	})
	return res, nil
}

// A genus that matches all names by chance (Ficus) loses to the genus
// that is the nearest hint of a name (Formica for `F. rufa`). A name
// whose genus is not its nearest hint and does not hold for two names
// stays as it was.
func TestResolveUnsupportedGenus(t *testing.T) {
	assert := assert.New(t)
	f := &tableNamer{genera: map[string][]string{
		"F. rufa":     {"Ficus", "Formica"},
		"F. elastica": {"Ficus"},
	}}
	g := gnabbr.New(config.New(), f)

	d := doc.Document{
		Names: []doc.Name{
			{Name: "Ficus carica", Start: 0},
			{Name: "Formica rufa", Start: 50},
			{Name: "F. rufa", Start: 100},
			{Name: "F. elastica", Start: 150},
		},
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)

	require.Len(t, out.Reconciled, 1)
	r := out.Reconciled[0]
	assert.True(r.Split)
	assert.Equal("Formica", r.Genus.Name)
	require.Len(t, r.Names, 1)
	assert.Equal("Formica rufa", r.Names[0].Name)

	require.Len(t, out.Failed, 1)
	assert.Equal("Ficus", out.Failed[0].Genus.Name)
	assert.Equal([]string{"F. elastica"}, out.Failed[0].Verbatims)
}

// An accepted genus expands only the names it matched; the others stay
// as they were and are listed as verbatims of the group.
func TestResolveUnmatchedNames(t *testing.T) {
	assert := assert.New(t)
	f := &tableNamer{genera: map[string][]string{
		"C. alba":  {"Carex"},
		"C. nigra": {"Carex"},
	}}
	cfg := config.New()
	cfg.Update([]config.Option{config.OptThreshold(-1)})
	f2 := &partialNamer{f}
	g := gnabbr.New(cfg, f2)

	d := doc.Document{
		Names: []doc.Name{
			{Name: "Carex alba", Start: 0},
			{Name: "C. alba", Start: 100},
			{Name: "C. nigra", Start: 200},
			{Name: "C. qqq", Verbatim: "C. qqq.", Start: 300},
		},
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)
	require.Len(t, out.Reconciled, 1)
	r := out.Reconciled[0]
	assert.False(r.Split)
	assert.Equal("Carex", r.Genus.Name)
	assert.Len(r.Names, 2)
	assert.Equal([]string{"C. qqq."}, r.Verbatims)
	assert.Empty(out.Failed)
}

// partialNamer is a tableNamer that keeps a genus with one unmatched
// name, like the default tolerance of abbr_names for groups of 5+ names.
type partialNamer struct {
	*tableNamer
}

func (f *partialNamer) AbbrNames(ctx context.Context, inp abbr.Input) (abbr.Output, error) {
	all := 100
	if inp.Tolerance == nil {
		inp.Tolerance = &all
	}
	return f.tableNamer.AbbrNames(ctx, inp)
}

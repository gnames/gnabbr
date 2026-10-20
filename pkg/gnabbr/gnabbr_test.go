package gnabbr_test

import (
	"context"
	"testing"

	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/gnabbr"
	"github.com/gnames/gnlib/ent/abbr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeNamer answers with Caenorhabditis (score 5) for all names but the
// last of a group, and with Cyperus (score 1) for the last one.
type fakeNamer struct {
	inputs []abbr.Input
}

func (f *fakeNamer) AbbrNames(_ context.Context, inp abbr.Input) (abbr.Output, error) {
	f.inputs = append(f.inputs, inp)
	var first, last []abbr.MatchedName
	for i, n := range inp.Names {
		m := abbr.MatchedName{ID: n.ID, Name: n.Name}
		if i == len(inp.Names)-1 {
			last = append(last, m)
		} else {
			first = append(first, m)
		}
	}
	return abbr.Output{Genera: []abbr.Genus{
		{Name: "Caenorhabditis", Score: abbr.Score{Total: 5}, MatchedNames: first},
		{Name: "Cyperus", Score: abbr.Score{Total: 1}, MatchedNames: last},
	}}, nil
}

func TestResolve(t *testing.T) {
	assert := assert.New(t)
	cfg := config.New()
	cfg.Update([]config.Option{config.OptThreshold(2)})
	f := &fakeNamer{}
	g := gnabbr.New(cfg, f)

	d := doc.Document{
		ID: "item-1",
		Occurrences: []doc.Occurrence{
			{ID: "a", Name: "Caenorhabditis elegans", Page: 1},
			{ID: "b", Name: "C. elegans", Page: 2},
			{ID: "c", Name: "C. briggsae", Page: 2, Offset: 5},
			{ID: "d", Name: "C. remanei", Page: 3, Part: "p2"},
		},
		Contexts: map[string]doc.Context{
			"":   {DominantClass: "Chromadorea"},
			"p2": {MainTaxon: "Rhabditida"},
		},
	}
	out, err := g.Resolve(context.Background(), d)
	require.NoError(t, err)
	assert.Equal("item-1", out.DocumentID)
	require.Len(t, out.Groups, 2)
	require.Len(t, out.Results, 3)

	b, c, dd := out.Results[0], out.Results[1], out.Results[2]
	assert.Equal("Caenorhabditis", b.Genus)
	assert.True(b.Accepted)
	assert.Equal("Cyperus", c.Genus)
	assert.False(c.Accepted)
	assert.Equal(1, dd.Group)

	in := f.inputs[0]
	assert.Equal("Chromadorea", in.DominantClass)
	assert.Equal([]string{"Caenorhabditis"}, in.Names[0].GenusHints)
	assert.Equal("Rhabditida", f.inputs[1].MainTaxon)
	assert.Empty(f.inputs[1].Names[0].GenusHints)
}

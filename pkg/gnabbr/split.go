package gnabbr

import (
	"context"
	"slices"
	"strconv"

	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/group"
	"github.com/gnames/gnabbr/pkg/hints"
	"github.com/gnames/gnstats/ent/stats"
)

const (
	// splitGeneraLimit is the number of genera asked for to split a group.
	splitGeneraLimit = 10

	// minRun is the number of names in a row a new genus needs to start
	// a run, unless the genus is the nearest hint of the name.
	minRun = 2
)

// fullTolerance lets abbr_names return every genus that matched any
// name of a group.
var fullTolerance = 100

// run is a part of a group whose names share a genus.
type run struct {
	genus string
	group group.Group
}

// segment cuts a group into runs of names that share a genus. A text may
// use one abbreviation for several genera close to each other (`G.` for
// Gastrophilus, then for Glossina). Then no genus matches enough names
// of the whole group, or a genus that matches all names by chance wins
// over the genera the text names (Ficus over Formica for `F. rufa`).
//
// The group is asked with full tolerance, so every genus that matched a
// name comes back. A name gets the genus that matched it and is nearest
// among its hints; a name whose matching genera are not among its hints
// gets none (OCR errors matching some genus with the letter). In reading
// order, a run starts where the genus changes and the new genus holds
// for minRun names, or is the nearest hint of the name; other
// interruptions are left out. When genera alternate (checklists,
// indexes), few runs start and most names stay as they were.
func (g *gnabbr) segment(
	ctx context.Context,
	names []doc.Name,
	gr group.Group,
	hs map[int][]string,
	st stats.Stats,
) ([]run, error) {
	uniq, own, _ := distinct(names, gr, hs)
	p, err := g.ask(ctx, names, gr, hs, st, &fullTolerance, splitGeneraLimit)
	if err != nil {
		return nil, err
	}

	matched := make(map[string]map[string]bool)
	for _, gen := range p.details.Output.Genera {
		for _, m := range gen.MatchedNames {
			j, err := strconv.Atoi(m.ID)
			if err != nil || j < 0 || j >= len(uniq) {
				continue
			}
			if matched[uniq[j]] == nil {
				matched[uniq[j]] = make(map[string]bool)
			}
			matched[uniq[j]][gen.Name] = true
		}
	}
	type choice struct {
		genus  string
		strong bool
	}
	chosen := make(map[string]choice)
	for _, n := range uniq {
		for k, h := range hints.Merge(g.cfg.HintsLimit, own[n]...) {
			if matched[n][h] {
				chosen[n] = choice{genus: h, strong: k == 0}
				break
			}
		}
	}

	var members []int
	var genera []string
	var strong []bool
	for _, i := range gr.Members {
		if c, ok := chosen[names[i].Name]; ok {
			members = append(members, i)
			genera = append(genera, c.genus)
			strong = append(strong, c.strong)
		}
	}

	var res []run
	for _, r := range runs(genera, strong) {
		ms := make([]int, len(r))
		for k, pos := range r {
			ms[k] = members[pos]
		}
		res = append(res, run{
			genus: genera[r[0]],
			group: group.Group{
				Section: gr.Section,
				Prefix:  gr.Prefix,
				Start:   names[ms[0]].Start,
				End:     names[ms[len(ms)-1]].Start,
				Members: ms,
			},
		})
	}
	return res, nil
}

// runs cuts a sequence of genera into runs of the same genus, by
// positions. A new genus starts a run if it holds for minRun positions
// or is strong at its position (the nearest hint of the name); other
// interruptions are left out. If all genera are the same, the whole
// sequence is one run.
func runs(genera []string, strong []bool) [][]int {
	if len(genera) == 0 {
		return nil
	}
	if !slices.ContainsFunc(genera, func(g string) bool { return g != genera[0] }) {
		res := make([]int, len(genera))
		for k := range res {
			res[k] = k
		}
		return [][]int{res}
	}

	var res [][]int
	for k, gen := range genera {
		if n := len(res); n > 0 && genera[res[n-1][0]] == gen {
			res[n-1] = append(res[n-1], k)
			continue
		}
		holds := k+minRun <= len(genera) &&
			!slices.ContainsFunc(genera[k:k+minRun], func(g string) bool { return g != gen })
		if holds || (k < len(strong) && strong[k]) {
			res = append(res, []int{k})
		}
	}
	return res
}

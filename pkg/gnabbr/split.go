package gnabbr

import (
	"context"
	"slices"
	"strconv"

	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/gnames/gnabbr/pkg/group"
	"github.com/gnames/gnstats/ent/stats"
)

const (
	// splitGeneraLimit is the number of genera asked for to split a group.
	splitGeneraLimit = 10

	// minRun is the number of names in a row a new genus needs to start
	// a run; a shorter interruption is left out of the runs.
	minRun = 2
)

// fullTolerance lets abbr_names return every genus that matched any
// name of a group.
var fullTolerance = 100

// split cuts a group that got no genus into runs of names that share a
// genus, and reconciles each run as a group of its own. A text may use
// one abbreviation for several genera close to each other (`G.` for
// Gastrophilus, then for Glossina), and then no genus matches enough
// names of the whole group.
//
// The group is asked again with full tolerance, and each distinct name
// gets the best genus that matched it. Only genera supported by hints
// count: the genus is among the hints of one of its names; others are
// noise, such as OCR errors matching some genus. In reading order, the
// group is cut where the genus changes and the new genus holds for
// minRun names. When genera alternate (checklists, indexes), no run is
// long enough and the group stays failed.
func (g *gnabbr) split(
	ctx context.Context,
	names []doc.Name,
	gr group.Group,
	hs map[int][]string,
	st stats.Stats,
) ([]candidate, error) {
	uniq, own, _ := distinct(names, gr, hs)
	if len(uniq) < 2 {
		return nil, nil
	}
	p, err := g.ask(ctx, names, gr, hs, st, &fullTolerance, splitGeneraLimit)
	if err != nil {
		return nil, err
	}

	// Genera come best first, so the first genus of a name is its best.
	best := make(map[string]string)
	for _, gen := range p.details.Output.Genera {
		for _, m := range gen.MatchedNames {
			j, err := strconv.Atoi(m.ID)
			if err != nil || j < 0 || j >= len(uniq) {
				continue
			}
			if _, ok := best[uniq[j]]; !ok {
				best[uniq[j]] = gen.Name
			}
		}
	}
	supported := make(map[string]bool)
	for n, gen := range best {
		if slices.ContainsFunc(own[n], func(h []string) bool {
			return slices.Contains(h, gen)
		}) {
			supported[gen] = true
		}
	}

	var members []int
	var genera []string
	for _, i := range gr.Members {
		if gen := best[names[i].Name]; supported[gen] {
			members = append(members, i)
			genera = append(genera, gen)
		}
	}

	var res []candidate
	for _, run := range runs(genera) {
		ms := make([]int, len(run))
		for k, r := range run {
			ms[k] = members[r]
		}
		sub := group.Group{
			Section: gr.Section,
			Prefix:  gr.Prefix,
			Start:   names[ms[0]].Start,
			End:     names[ms[len(ms)-1]].Start,
			Members: ms,
		}
		c, err := g.ask(ctx, names, sub, hs, st, nil, g.cfg.GeneraLimit)
		if err != nil {
			return nil, err
		}
		c.split = true
		res = append(res, c)
	}
	return res, nil
}

// runs cuts a sequence of genera into runs of the same genus, by
// positions. A new genus starts a run only if it holds for minRun
// positions; shorter interruptions are left out. If all genera are the
// same, the whole sequence is one run.
func runs(genera []string) [][]int {
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
		if k+minRun <= len(genera) &&
			!slices.ContainsFunc(genera[k:k+minRun], func(g string) bool { return g != gen }) {
			res = append(res, []int{k})
		}
	}
	return res
}

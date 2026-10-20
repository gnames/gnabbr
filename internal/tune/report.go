package tune

import (
	"fmt"
	"io"
	"math"
	"slices"
	"text/tabwriter"
)

// Thresholds compared in the report.
var Thresholds = []float64{math.Inf(-1), 0, 2, 4, 5, 6, 7, 8, 9, 10, 12}

// Report writes agreement of expanded genera with labels, by label kind
// and by threshold. Shares are weighted by sample weights, so they
// estimate BHL as a whole; occ are plain counts.
func Report(w io.Writer, res []Outcome) {
	type sums struct {
		n                int
		all, exp, agreed float64
	}
	byKind := make(map[string]*sums)
	for _, o := range res {
		s, ok := byKind[o.Kind]
		if !ok {
			s = &sums{}
			byKind[o.Kind] = s
		}
		s.n++
		s.all += o.Weight
		if o.Genus != "" {
			s.exp += o.Weight
			if o.Genus == o.Label {
				s.agreed += o.Weight
			}
		}
	}
	kinds := make([]string, 0, len(byKind))
	for k := range byKind {
		kinds = append(kinds, k)
	}
	slices.Sort(kinds)

	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tw, "label\tocc\texpanded\tagree\t\n")
	for _, k := range kinds {
		s := byKind[k]
		agree := "-"
		if s.exp > 0 && k != "no label" {
			agree = pct(s.agreed / s.exp)
		}
		fmt.Fprintf(tw, "%s\t%d\t%s\t%s\t\n", k, s.n, pct(s.exp/s.all), agree)
	}
	tw.Flush()

	sweep(w, "antecedent high+medium", res, func(o Outcome) bool {
		return o.Kind == "antecedent high" || o.Kind == "antecedent medium"
	})
	sweep(w, "last_genus", res, func(o Outcome) bool {
		return o.Kind == "last_genus"
	})
}

// sweep writes, for each threshold, the share of labeled occurrences
// kept (expanded with a score at the threshold or above) and the
// agreement of kept ones with the labels.
func sweep(w io.Writer, title string, res []Outcome, in func(Outcome) bool) {
	fmt.Fprintf(w, "\n%s, by threshold\n", title)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(tw, "threshold\tkept\tagree\t\n")
	for _, t := range Thresholds {
		var all, kept, agreed float64
		for _, o := range res {
			if !in(o) {
				continue
			}
			all += o.Weight
			if o.Genus != "" && o.Score >= t {
				kept += o.Weight
				if o.Genus == o.Label {
					agreed += o.Weight
				}
			}
		}
		if all == 0 {
			return
		}
		agree := "-"
		if kept > 0 {
			agree = pct(agreed / kept)
		}
		fmt.Fprintf(tw, "%g\t%s\t%s\t\n", t, pct(kept/all), agree)
	}
	tw.Flush()
}

func pct(f float64) string {
	return fmt.Sprintf("%.1f%%", 100*f)
}

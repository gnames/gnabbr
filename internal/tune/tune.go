package tune

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gnames/gnabbr/internal/iognames"
	"github.com/gnames/gnabbr/pkg/abbrname"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/gnabbr"
)

// Options of a tuning run.
type Options struct {
	// DataDir has docs/, docs-masked/, items.jsonl and labels.jsonl.
	DataDir string

	// Docs is the directory of documents in DataDir: `docs`, or
	// `docs-masked` for documents without antecedents of abbreviated
	// names, which labels come from.
	Docs string

	// CachePath is the SQLite file with cached abbr_names answers.
	CachePath string

	// OutPath, if not empty, gets one Outcome per line.
	OutPath string

	// Split is `tune`, `holdout` or `all`.
	Split string

	// Items is the number of items to run, all of the split if 0. Items
	// are taken in a random order given by Seed.
	Items int
	Seed  uint64

	// Jobs is the number of items reconciled at once.
	Jobs int

	// Config has gnabbr settings. Its threshold is ignored: every best
	// genus is accepted, and thresholds are compared in the report.
	Config *config.Config
}

// Outcome is the result for one labeled occurrence.
type Outcome struct {
	ItemID       int     `json:"itemId"`
	OccurrenceID int     `json:"occurrenceId"`
	Kind         string  `json:"kind"`
	Label        string  `json:"label,omitempty"`
	Weight       float64 `json:"weight"`
	Genus        string  `json:"genus,omitempty"`
	Score        float64 `json:"score"`
	Group        int     `json:"group"`

	// Split is true if the name was expanded by a run of a split mixed
	// group.
	Split bool `json:"split,omitempty"`

	// Status is `expanded`, `not abbreviated` (gnabbr does not see the
	// name as abbreviated) or `no genus` (abbr_names found no genus for
	// its group).
	Status string `json:"status"`
}

// Run reconciles the items of a split and returns outcomes of their
// selected labeled occurrences.
func Run(ctx context.Context, opts Options) ([]Outcome, error) {
	items, err := readItems(opts.DataDir)
	if err != nil {
		return nil, err
	}
	labels, err := readLabels(opts.DataDir)
	if err != nil {
		return nil, err
	}
	var run []Item
	for _, it := range items {
		if opts.Split == "all" || it.Split == opts.Split {
			run = append(run, it)
		}
	}
	rand.New(rand.NewPCG(opts.Seed, 0)).Shuffle(len(run), func(i, j int) {
		run[i], run[j] = run[j], run[i]
	})
	if opts.Items > 0 && opts.Items < len(run) {
		run = run[:opts.Items]
	}

	cfg := *opts.Config
	cfg.Threshold = math.Inf(-1)
	version, err := gnamesVersion(cfg.GnamesURL)
	if err != nil {
		return nil, err
	}
	c, err := newCache(opts.CachePath, version, iognames.New(cfg.GnamesURL))
	if err != nil {
		return nil, err
	}
	defer c.Close()
	core := gnabbr.New(&cfg, c)
	slog.Info("Tuning run started", "items", len(run), "split", opts.Split,
		"docs", opts.Docs,
		"gnames_version", version, "cache", opts.CachePath)

	start := time.Now()
	jobs := make(chan Item)
	var mu sync.Mutex
	var res []Outcome
	var firstErr error
	var done int
	var wg sync.WaitGroup
	for range max(opts.Jobs, 1) {
		wg.Go(func() {
			for it := range jobs {
				outs, err := runItem(ctx, core, opts, it, labels[it.ItemID])
				mu.Lock()
				done++
				if err != nil && firstErr == nil {
					firstErr = fmt.Errorf("item %d: %w", it.ItemID, err)
				}
				res = append(res, outs...)
				// Progress goes to STDERR, so STDOUT has only the report.
				fmt.Fprintf(os.Stderr, "\r%d/%d items, %d outcomes, %.0fs",
					done, len(run), len(res), time.Since(start).Seconds())
				mu.Unlock()
			}
		})
	}
	for _, it := range run {
		jobs <- it
	}
	close(jobs)
	wg.Wait()
	fmt.Fprintln(os.Stderr)
	slog.Info("Tuning run finished", "outcomes", len(res),
		"cache_hits", c.hits, "cache_misses", c.misses,
		"seconds", time.Since(start).Seconds())
	if firstErr != nil {
		return res, firstErr
	}

	if opts.OutPath != "" {
		err = writeOutcomes(opts.OutPath, res)
	}
	return res, err
}

func runItem(
	ctx context.Context,
	core gnabbr.GNabbr,
	opts Options,
	it Item,
	labels []Label,
) ([]Outcome, error) {
	d, err := readDoc(opts.DataDir, opts.Docs, it.ItemID)
	if err != nil {
		return nil, err
	}
	out, err := core.Resolve(ctx, d)
	if err != nil {
		return nil, err
	}

	type expanded struct {
		genus string
		score float64
		group int
		split bool
	}
	exp := make(map[string]expanded)
	for _, g := range out.Reconciled {
		for _, n := range g.Names {
			exp[n.ID] = expanded{g.Genus.Name, g.Genus.Score.Total, g.Index, g.Split}
		}
	}

	names := make(map[string]string, len(d.Names))
	for _, n := range d.Names {
		names[n.ID] = n.Name
	}

	w := it.Weight
	if it.GroupQ > 0 {
		w /= it.GroupQ
	}
	res := make([]Outcome, 0, len(labels))
	for _, l := range labels {
		o := Outcome{
			ItemID: it.ItemID, OccurrenceID: l.OccurrenceID,
			Kind: l.Kind(), Weight: w, Group: -1,
		}
		if l.Genus != nil {
			o.Label = *l.Genus
		}
		id := strconv.Itoa(l.OccurrenceID)
		e, ok := exp[id]
		_, abbr := abbrname.Parse(names[id])
		switch {
		case ok:
			o.Genus, o.Score, o.Group, o.Split = e.genus, e.score, e.group, e.split
			o.Status = "expanded"
		case !abbr:
			o.Status = "not abbreviated"
		default:
			o.Status = "no genus"
		}
		res = append(res, o)
	}
	return res, nil
}

// gnamesVersion returns the version and build of gnames at url, a part
// of cache keys: answers of another gnames build are not reused.
func gnamesVersion(url string) (string, error) {
	resp, err := http.Get(strings.TrimSuffix(url, "/") + "/version")
	if err != nil {
		return "", fmt.Errorf("gnames version: %w", err)
	}
	defer resp.Body.Close()
	var v struct{ Version, Build string }
	if err = json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return "", fmt.Errorf("gnames version: %w", err)
	}
	return v.Version + " " + v.Build, nil
}

func writeOutcomes(path string, res []Outcome) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(f)
	for _, o := range res {
		if err = enc.Encode(o); err != nil {
			f.Close()
			return err
		}
	}
	return f.Close()
}

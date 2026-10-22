package cmd

import (
	"os"
	"path/filepath"

	"github.com/gnames/gnabbr/internal/tune"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/spf13/cobra"
)

func getTuneCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tune",
		Short: "Runs the BHL sample and compares expanded names with labels",
		Long: `Reconciles documents of the BHL sample exported by abbrctx/export.py
(docs/<item>.json, items.jsonl, labels.jsonl) and prints agreement of
expanded genera with labels, by label kind and by threshold, weighted by
sample weights. abbr_names answers are cached, so reruns with the same
groups and hints make no calls. Temporary: goes away after tuning.`,
		Args: cobra.NoArgs,
		RunE: runTune,
	}
	cmd.Flags().StringP("data", "D", "", "directory of the exported sample (required)")
	cmd.Flags().String("docs", "docs", "documents: docs, or docs-masked (no antecedents)")
	cmd.Flags().StringP("cache", "c", "", "abbr_names cache (default <cache dir>/abbr_names.sqlite)")
	cmd.Flags().StringP("out", "o", "", "file for outcomes, one JSON per line")
	cmd.Flags().String("groups", "", "file for groups of the run items, one JSON per line")
	cmd.Flags().StringP("split", "s", "tune", "tune, holdout or all")
	cmd.Flags().IntP("items", "n", 0, "number of random items (0: all of the split)")
	cmd.Flags().Uint64("seed", 1, "seed of the random order of items")
	cmd.Flags().IntP("jobs", "j", 4, "items reconciled at once")
	cmd.Flags().StringP("url", "u", "", "gnames API URL")
	cmd.Flags().BoolP("features", "f", false, "ask abbr_names for score features")
	_ = cmd.MarkFlagRequired("data")
	return cmd
}

func runTune(cmd *cobra.Command, _ []string) error {
	opts := cfg.ToOptions()
	if s, _ := cmd.Flags().GetString("url"); s != "" {
		opts = append(opts, config.OptGnamesURL(s))
	}
	if b, _ := cmd.Flags().GetBool("features"); b {
		opts = append(opts, config.OptWithFeatures(b))
	}
	c := config.New()
	c.Update(opts)

	f := cmd.Flags()
	var o tune.Options
	o.Config = c
	o.DataDir, _ = f.GetString("data")
	o.Docs, _ = f.GetString("docs")
	o.CachePath, _ = f.GetString("cache")
	o.OutPath, _ = f.GetString("out")
	o.GroupsPath, _ = f.GetString("groups")
	o.Split, _ = f.GetString("split")
	o.Items, _ = f.GetInt("items")
	o.Seed, _ = f.GetUint64("seed")
	o.Jobs, _ = f.GetInt("jobs")
	if o.CachePath == "" {
		o.CachePath = filepath.Join(config.CacheDir(cfg.HomeDir), "abbr_names.sqlite")
	}

	res, err := tune.Run(cmd.Context(), o)
	if err != nil {
		return err
	}
	tune.Report(os.Stdout, res)
	return nil
}

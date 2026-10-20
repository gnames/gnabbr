package cmd

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/gnames/gnabbr"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/spf13/cobra"
)

func getResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve [file]",
		Short: "Resolves abbreviated names from a file or STDIN",
		Long: `Reads names in reading order, one per line, from a file or STDIN,
and prints the resolved abbreviated names as JSON. A line is a name or
'page<TAB>name'. Full names give genus hints to abbreviated ones.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runResolve,
	}
	cmd.Flags().StringP("url", "u", "", "gnames API URL")
	cmd.Flags().BoolP("features", "f", false, "return score features")
	cmd.Flags().IntP("genera-limit", "g", 0, "candidate genera per group")
	return cmd
}

func runResolve(cmd *cobra.Command, args []string) error {
	var opts []config.Option
	if s, _ := cmd.Flags().GetString("url"); s != "" {
		opts = append(opts, config.OptGnamesURL(s))
	}
	if b, _ := cmd.Flags().GetBool("features"); b {
		opts = append(opts, config.OptWithFeatures(b))
	}
	if i, _ := cmd.Flags().GetInt("genera-limit"); i > 0 {
		opts = append(opts, config.OptGeneraLimit(i))
	}
	cfg.Update(opts)

	r := io.Reader(os.Stdin)
	if len(args) == 1 {
		f, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer f.Close()
		r = f
	}
	d, err := readDocument(r)
	if err != nil {
		return err
	}

	out, err := gnabbr.New(append(cfg.ToOptions(), opts...)...).
		Resolve(cmd.Context(), d)
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func readDocument(r io.Reader) (doc.Document, error) {
	var d doc.Document
	sc := bufio.NewScanner(r)
	var n int
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		n++
		page := 1
		if p, name, ok := strings.Cut(line, "\t"); ok {
			if i, err := strconv.Atoi(p); err == nil {
				page, line = i, strings.TrimSpace(name)
			}
		}
		d.Occurrences = append(d.Occurrences, doc.Occurrence{
			ID: strconv.Itoa(n), Name: line, Page: page, Offset: n,
		})
	}
	return d, sc.Err()
}

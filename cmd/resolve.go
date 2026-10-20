package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/gnames/gnabbr"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
	"github.com/spf13/cobra"
)

func getResolveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resolve [file]",
		Short: "Resolves abbreviated names from a file or STDIN",
		Long: `Reads a document from a file or STDIN and prints reconciled and
failed groups of abbreviated names as JSON.

The input is a JSON document (names with rune offsets, optional sections
and TaxonStats, see pkg/doc), or plain text with one name per line in
reading order; then the offset of a name is the offset of its line.
Full names give genus hints to abbreviated ones.`,
		Args: cobra.MaximumNArgs(1),
		RunE: runResolve,
	}
	cmd.Flags().StringP("url", "u", "", "gnames API URL")
	cmd.Flags().BoolP("features", "f", false, "return score features")
	cmd.Flags().BoolP("details", "d", false, "return abbr_names input and output")
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
	if b, _ := cmd.Flags().GetBool("details"); b {
		opts = append(opts, config.OptWithDetails(b))
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
	b, err := io.ReadAll(r)
	if err != nil {
		return d, err
	}
	if t := bytes.TrimSpace(b); len(t) > 0 && t[0] == '{' {
		err = json.Unmarshal(t, &d)
		return d, err
	}

	sc := bufio.NewScanner(bytes.NewReader(b))
	var start int
	for sc.Scan() {
		line := sc.Text()
		if name := strings.TrimSpace(line); name != "" {
			d.Names = append(d.Names, doc.Name{Name: name, Start: start})
		}
		start += utf8.RuneCountInString(line) + 1
	}
	return d, sc.Err()
}

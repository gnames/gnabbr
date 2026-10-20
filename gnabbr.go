// Package gnabbr reconciles abbreviated genera of scientific names
// (`C. elegans` -> Caenorhabditis elegans) found in documents.
//
// It is the entry point of the library: it joins the core logic
// (pkg/gnabbr) with the internal client of gnames abbr_names, so callers
// never deal with gnames.
//
// Usage:
//
//	ga := gnabbr.New(config.OptGnamesURL("http://localhost:8888/api/v1/"))
//	out, err := ga.Resolve(ctx, d)
package gnabbr

import (
	"github.com/gnames/gnabbr/internal/iognames"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/gnabbr"
)

// GNabbr reconciles abbreviated names of a document, see pkg/gnabbr.
type GNabbr = gnabbr.GNabbr

// Output is the result of Resolve, see pkg/gnabbr.
type Output = gnabbr.Output

// New creates GNabbr from default settings changed by opts. It calls
// abbr_names at the gnames URL of the settings.
func New(opts ...config.Option) GNabbr {
	cfg := config.New()
	cfg.Update(opts)
	return gnabbr.New(cfg, iognames.New(cfg.GnamesURL))
}

# gnabbr

gnabbr resolves abbreviated genera of scientific names found in
documents: `C. elegans` -> *Caenorhabditis elegans*. It is a Go library
for [gnfinder] and [bhlindex], with a command-line tool used to tune it
on data of the Biodiversity Heritage Library (BHL).

Status: early development.

## How it works

1. **Find** names with an abbreviated genus: a capital letter with up to
   two lowercase letters and a period, then an epithet (`C. alba`,
   `Cae. alba`).
2. **Group** them locally: the same abbreviation in the same part
   (article) of the document, split where pages are far apart. Names in
   a group usually share a genus.
3. **Hint**: full names mentioned shortly before (`Carex nigra`) give
   genera the abbreviation may stand for.
4. **Resolve** each group with [gnames] `abbr_names`, adding the
   taxonomic context of the part (dominant class, main taxon).
5. **Accept** the best genus when its score reaches a threshold.

## Library

```go
import (
	"github.com/gnames/gnabbr"
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/doc"
)

ga := gnabbr.New(config.OptGnamesURL("http://localhost:8888/api/v1/"))
out, err := ga.Resolve(ctx, doc.Document{
	ID: "item-1",
	Occurrences: []doc.Occurrence{
		{ID: "1", Name: "Caenorhabditis elegans", Page: 1},
		{ID: "2", Name: "C. briggsae", Page: 2},
	},
})
```

Callers do not deal with gnames: the root package `gnabbr` joins the
core logic (`pkg/gnabbr`) with the internal `abbr_names` client. `pkg/`
never imports `internal/` and has no CLI, configuration file or
database dependencies.

## Command line

```bash
just install
printf '1\tCarex alba\n2\tC. nigra\n' | gnabbr resolve -u http://localhost:8888/api/v1/
```

A line is a name or `page<TAB>name`. Settings are in
`~/.config/gnabbr/config.yaml` or `GNABBR_*` environment variables.

## Layout

```
gnabbr.go           entry point: New(opts...) joins pkg/gnabbr and internal/iognames
cmd/gnabbr          main
cmd                 CLI (Cobra)
pkg/gnabbr          core: GNabbr, AbbrNamer, Resolve, Output
pkg/config          settings (functional options, no I/O)
pkg/doc             input: documents and name occurrences
pkg/abbrname        parsing abbreviated names
pkg/group           local groups
pkg/hints           genus hints
internal/iognames   abbr_names REST client (implements pkg/gnabbr.AbbrNamer)
internal/iofs       config directories and file (CLI)
internal/iologger   logging setup (CLI)
internal/tune       temporary tuning client (BHL sample, weights, threshold)
```

## Tuning

Score weights live in gnames `abbr_names`; the threshold and calibration
live here. They are tuned on a weighted sample of 1,127 BHL items with
labels from the items themselves; see `internal/tune`.

[gnfinder]: https://github.com/gnames/gnfinder
[bhlindex]: https://github.com/gnames/bhlindex
[gnames]: https://github.com/gnames/gnames

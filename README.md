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
2. **Group** them locally: the same abbreviation in the same section
   (article, chapter) of the document, split where names are far apart
   (in characters). Names in a group usually share a genus; a repeated
   name is reconciled once per group.
3. **Hint**: full names mentioned shortly before (`Carex nigra`) give
   genera the abbreviation may stand for.
4. **Reconcile** each group with [gnames] `abbr_names`, adding the
   taxonomic statistics of the section or the document (main taxon,
   dominant class), given by the caller.
5. **Accept** the best genus when its score reaches a threshold (8 by
   default, which favors precision), and expand the group's names with
   it (`C. alba` -> `Carex alba`).
   Callers verify expanded names with their other names.

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
	Names: []doc.Name{
		{ID: "1", Name: "Caenorhabditis elegans", Start: 0},
		{ID: "2", Name: "C. briggsae", Start: 2400},
	},
	TaxonStats: taxonStats, // gnstats stats.Stats of the document
})
// out.Reconciled: groups with an accepted genus and expanded names;
// out.Failed: groups without one, with their verbatims.
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

The input is a JSON document (see `pkg/doc`) or one name per line;
with lines, the offset of a name is the offset of its line. Settings are in
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
internal/tune       temporary tuning client (`gnabbr tune`)
```

## Tuning

Score weights live in gnames `abbr_names`; the threshold and calibration
live here. They are tuned on a weighted sample of 1,127 BHL items with
labels from the items themselves; see `internal/tune`.

```bash
gnabbr tune -D path/to/bhl-abbr-export -u http://localhost:8888/api/v1/ \
  -o outcomes.jsonl
```

`tune` reconciles the exported sample (`docs/<item>.json`, `items.jsonl`,
`labels.jsonl`), caches `abbr_names` answers by gnames build and request,
and prints weighted agreement with labels by label kind and threshold.

[gnfinder]: https://github.com/gnames/gnfinder
[bhlindex]: https://github.com/gnames/bhlindex
[gnames]: https://github.com/gnames/gnames

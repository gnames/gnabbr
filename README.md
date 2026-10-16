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
	"github.com/gnames/gnabbr/pkg/config"
	"github.com/gnames/gnabbr/pkg/ent/doc"
	"github.com/gnames/gnabbr/pkg/gnabbr"
)

cfg := config.New()
ga := gnabbr.New(cfg)
out, err := ga.Resolve(ctx, doc.Document{
	ID: "item-1",
	Occurrences: []doc.Occurrence{
		{ID: "1", Name: "Caenorhabditis elegans", Page: 1},
		{ID: "2", Name: "C. briggsae", Page: 2},
	},
})
```

`pkg/` and `internal/io` have no CLI, configuration file or database
dependencies.

## Command line

```bash
just install
printf '1\tCarex alba\n2\tC. nigra\n' | gnabbr resolve -u http://localhost:8888/api/v1/
```

A line is a name or `page<TAB>name`. Settings are in
`~/.config/gnabbr/config.yaml` or `GNABBR_*` environment variables.

## Layout

```
pkg/gnabbr        public API: GNabbr, Resolve, Output
pkg/config        settings (functional options, no I/O)
pkg/ent/doc       input: documents and name occurrences
pkg/ent/abbrname  parsing abbreviated names
pkg/ent/group     local groups
pkg/ent/hints     genus hints
internal/io/...   abbr_names REST client
internal/tune     temporary tuning client (BHL sample, weights, threshold)
cmd               CLI
```

## Tuning

Score weights live in gnames `abbr_names`; the threshold and calibration
live here. They are tuned on a weighted sample of 1,127 BHL items with
labels from the items themselves; see `internal/tune`.

[gnfinder]: https://github.com/gnames/gnfinder
[bhlindex]: https://github.com/gnames/bhlindex
[gnames]: https://github.com/gnames/gnames

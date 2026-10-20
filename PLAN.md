# gnabbr plan

## Scope

gnabbr reconciles names with abbreviated genera to full names known to
gnverifier data sources (`C. elegans` -> *Caenorhabditis elegans*), among
names already detected in a document. It does not detect names or split
documents into sections; callers do that.

Output follows the gnfinder model (verbatim name, name), so expanded
names merge with other names. Callers verify expanded names together
with their other names. When a name comes from a taxonomic data source
(e.g. Catalogue of Life), verification may also give the current name.
This is a convenience, not a guarantee.

Pipeline: name finding (gnfinder) -> reconciliation of abbreviated names
(gnabbr) -> verification of expanded names with all other names
(caller).

Goal: a significant improvement over the current state, not a perfect
result. Some mistakes are acceptable. Currently BHL does not reconcile
abbreviated names at all, so every correct reconciliation is a gain;
wrong ones are the cost to keep low.

### Terms

- **Document**: a text processed as a unit, given as detected names in
  reading order. The outermost scope.
- **Section** (optional): an autonomous part of a document (article,
  chapter). Its abbreviations refer only to genera introduced in it.
  Not the taxonomic rank.
- **Offset**: position of a name from the start of the document, in
  characters. The only position gnabbr uses; there are no pages.
- **TaxonStats**: taxonomic statistics of a document or section (main
  taxon, most prevalent class), as calculated by gnstats. Always given
  by the caller.
- **Group**: abbreviated names in one scope that are expected to share a
  genus.

### Out of scope

- Name detection (gnfinder).
- Verification of expanded names (callers, in one batch with their other
  names).
- Source-specific terms (BHL items, parts, pages); callers map them to
  documents, sections and offsets.
- Finding section boundaries.
- Calculating TaxonStats. Callers (gnfinder, bhlindex) already verify
  names and have the taxonomic data it needs; gnabbr would have to
  get the same data and repeat their work.

## Implementation

### Input contract

```
Document
  ID          string        optional
  Names       []Name        required, any order (gnabbr sorts by Start)
  Sections    []Section     optional
  TaxonStats  TaxonStats    required, whole document

Name                         (subset of gnfinder output.Name)
  ID           string        optional; index in Names if empty
  Verbatim     string        optional, carried to output
  Name         string        required (normalized, e.g. `C. alba`)
  Start        int           required, runes from document start

Section
  ID          string
  Start, End  int            rune range in the document
  TaxonStats  *TaxonStats    optional; document's if empty
```

- Sections are offset ranges, not tags on names: names stay plain
  gnfinder names, and gnabbr assigns them to sections.
- Offsets are in runes (gnfinder can give bytes; those are rejected).
- TaxonStats are required for the document: callers calculate them
  (gnstats) from their verified names. A section without its own
  TaxonStats uses the document's.
- Minimum input is `Names` with `Start` and the document's TaxonStats;
  the rest improves results or output readability. Callers join output
  to their data by `ID`.
- Callers pass all detected names: full names give hints.
- Callers pass every occurrence, repeated names included. gnabbr
  deals with repeats itself: a name string is reconciled once per
  group and the result applies to all its occurrences.

### Output contract

```
Output
  Meta         version, settings, counts, timing
  Reconciled   []Group
  Failed       []Group

Group                           (shaped like abbr_names output)
  Index, Section, Abbr, Start, End
  TaxonStats, Hints             as used
  Genus        abbr.Genus       best candidate
  Candidates   []abbr.Genus     filtered runners-up
  Names        []Name           expanded names (Reconciled groups only)
  Verbatims    []string         verbatims (Failed groups only)
  Details      raw abbr_names input/output, optional

Name                            (mirrors gnfinder output.Name)
  ID, Verbatim, Name, AbbrName, Start
```

- Callers (gnfinder, bhlindex) update only reconciled names; everything
  else stays as it was. So only names of reconciled groups carry full
  records; failed groups give distinct verbatims, for inspection.
- Expanded names are candidates until the caller verifies them (see
  Reconciliation).
- Own types with gnfinder's JSON tags, not gnfinder's types: gnfinder
  calls gnabbr, so gnabbr cannot import it. Shared types come from gnlib.
- Candidates far below the best genus are dropped, the rest capped at a
  few; the cut-off is decided after the first run.

### Steps

1. **Group** abbreviated names: the same abbreviation in the same section
   (or the whole document), split where names are far apart.
2. **Hint**: full names near an abbreviation suggest genera.
3. **Split** (hypothetical) groups whose hints point to different genera.
4. **Reconcile** each group with gnames `abbr_names`, adding TaxonStats.
5. **Accept** one best genus per group when its score reaches a
   threshold. A paper does not use the same abbreviation for two genera
   in one place.

### Hints

First run, with the current `abbr_names` API:

- Genera of full multinomials and full uninomials before the
  abbreviation, in the same section, nearest first.
- Hints are merged per group, so a hint before the first name of a group
  helps all its names.
- gnabbr collects hints; gnames `abbr_names` weighs them.
- Names carry no rank: a uninomial may be a higher taxon (*Coleoptera*
  before `C. auratus`). gnames knows the genera, so `abbr_names` numbers
  hint positions only among known genera; other hints are skipped and
  do not take hint slots or lower the weight of real ones. Currently
  they do (`gnames/pkg/ent/abbr/query.go`, `addHints`). This gnames
  change is needed before the first tuning run.

Possible expansions, to decide after the first run shows real problems
and opportunities:

- Hints after the abbreviation (e.g. a footnote at the bottom of the
  page), as weaker evidence.
- Distance of each hint, and whether its epithet matches the abbreviated
  one (`Carex nigra` ... `C. nigra`).
- Hints from the rest of the document as a weaker fallback.

These need structured hints in `abbr_names` (gnlib, gnames).

### Reconciliation

- The reconciled name is the accepted genus with the epithets as written
  (`C. albus` -> *Carex albus*); the matched name is evidence, not a
  replacement.
- gnabbr gives expanded names of accepted groups. Callers verify them in
  one batch with their other names (bhlindex verifies each unique name
  string once), with the same settings.
- Caller rule: keep an expanded name only if verification gives an exact
  or fuzzy match; skip it otherwise. It is better to miss an occurrence
  than to invent one. A fuzzy match is backed by its group and is most
  likely real. A partial match (genus only) is not: `C.` may be dirt on
  the page after a capital C.
- gnabbr's own runs (CLI, tuning) verify expanded names in the tuning
  code to measure precision.
- Genus, score and hints are reported per group (see Output contract).

### Acceptance

- Calibration is needed; its direction (e.g. weights of score features,
  margin over the second genus, group size) is decided after the first
  run.
- The threshold starts permissive (recall first) and is tightened on
  evidence.
- Rejected groups are still reported with their best genus, as
  candidates, not as reconciled names.

### Overlapping sections

Sections may overlap (BHL has duplicated parts and parts sharing pages).
For now each section is processed independently: a name in several
sections is reconciled in each of them, and the better match wins, so
each occurrence has one result.

### Undetected sections

Two unrelated sections, not marked by the caller, may follow each other
and use the same abbreviation for different genera (`C.` for *Carex*,
then for *Coleophora*). Distance alone puts them into one group. Step 3
would split such a group where its hints change from one genus to
another, or where names before and after a point are matched by
different genera. This splits groups only; it does not detect sections.

Hypothetical: added only if tuning data shows such groups are common.

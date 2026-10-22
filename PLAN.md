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
- TaxonStats may be empty when the caller has nothing to calculate them
  from (in the BHL sample, 64 of 1,127 items and 787 of 5,607 parts have
  no main taxon). gnabbr then reconciles without them.
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
3. **Reconcile** each group with gnames `abbr_names`, adding TaxonStats.
4. **Split** a group into runs of names that share a genus when its
   genus is missing or not the one its hints support (see Mixed
   groups), and reconcile each run.
5. **Accept** one best genus per group or run when its score reaches a
   threshold. A paper does not use the same abbreviation for two genera
   in one place.

### Hints

First run, with the current `abbr_names` API:

- Full names before the abbreviation, in the same section, nearest
  first: for each of the nearest genera starting with the abbreviation
  its nearest multinomial (`Carex nigra`), or the genus for a
  uninomial, and its names with the abbreviation's epithet (`Carex
  alba` for `C. alba`). `abbr_names` counts positions by distinct
  genera and weighs a name hint 1.5 times a genus hint, 2 times with
  the same epithet.
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
- Distance of each hint.
- Hints from the rest of the document as a weaker fallback.

These need structured hints in `abbr_names` (gnlib, gnames).

### Reconciliation

- The reconciled name is the accepted genus with the epithets as written
  (`C. albus` -> *Carex albus*); the matched name is evidence, not a
  replacement. Only names the genus matched are expanded.
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
  Runs of a split group are reported as groups marked `split`.

### Acceptance

- Calibration is needed; its direction (e.g. weights of score features,
  margin over the second genus, group size) is decided after the first
  run.
- The threshold started permissive (recall first) and was tightened on
  evidence. The default is 8: precision matters more than recall. On
  the BHL holdout gold set (reviewed labels) it gives 96.8% precision
  and 84.5% recall; 7 gives 96.0% and 88.0%.
- Rejected groups are still reported with their best genus, as
  candidates, not as reconciled names.

### Overlapping sections

Sections may overlap (BHL has duplicated parts and parts sharing pages).
For now each section is processed independently: a name in several
sections is reconciled in each of them, and the better match wins, so
each occurrence has one result.

### Mixed groups

A text may use one abbreviation for several genera close to each other
(`G.` for *Gastrophilus*, then for *Glossina*), and distance alone puts
them into one group. `abbr_names` keeps a genus only if at most
`tolerance` (20%) of the group's names are unmatched, so such a group,
or one with a few OCR errors, gets no genus at all.

First tuning run (tune split, 786 items): 35% of groups got no genus,
and so did 18.7% of names with high-confidence labels. Asked again with
`tolerance` 100, nearly all groups of 5+ names got genera. In 114 of
200 such groups, two or more genera matched names that also had them as
hints. Genera mostly follow each other: in reading order of distinct
names, ignoring one-name interruptions, 61% of these groups have runs
of 3+ names of a genus, 26% runs of 2-3, and 12% alternate (checklists,
indexes).

Groups that got a genus can be wrong the same way: in a small mixed
group, a genus that matches all names by chance wins over the genus the
text names (*Ficus* over *Formica* for `F. rufa`, *Cantharis* over
*Cornus* for `C. sanguinea`). Such cases were most of the expanded
names that disagreed with good labels in both plain and masked runs.

So every group of two or more distinct names is checked against hints:

1. Ask `abbr_names` again with `tolerance` 100 and more genera, so every
   genus that matched a name comes back. Tolerance filters genera only;
   scores do not change.
2. A name gets the genus that matched it and is nearest among its own
   hints. A name none of whose matching genera are among its hints gets
   none (OCR errors matching some genus with the letter).
3. In reading order, cut the group where the genus changes and the new
   genus holds for at least two names, or is the nearest hint of the
   name. Other interrupting names are left out of the runs. When genera
   alternate (checklists, indexes), few runs start.
4. If the runs are one run of the group's genus, or there are none, the
   group stays as it is.
5. Otherwise reconcile each run as a group of its own with the normal
   `tolerance`, so its score is comparable with the threshold. The
   group keeps names left out of all runs, without a genus.

An accepted genus expands only the names it matched; its other names
stay as they were (unmatched by the genus, 1% of expanded names with
good labels agreed with them only 24% of the time). This splits groups
only; it does not detect sections.

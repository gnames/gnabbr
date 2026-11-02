# Changelog

## [v0.0.1] - 2026-11-02

- Add: Go library that resolves abbreviated genera of scientific names in
  a document (`C. alba` -> `Carex alba`). The root package
  `gnabbr.New(opts...)` joins the core (`pkg/gnabbr`) with an internal
  gnames `abbr_names` client; `Resolve` takes a `doc.Document` and returns
  reconciled and failed groups.
- Add: input as document-wide rune offsets with optional sections and
  gnstats taxonomic statistics of the document or of each section.
- Add: abbreviated name parsing (a capital with up to two lowercase
  letters and a period, then an epithet).
- Add: local groups of the same abbreviation within a section, cut by the
  gap between names and by the group's span.
- Add: genus hints from full names mentioned before an abbreviation: the
  nearest multinomial of each nearby genus and its names with the same
  epithet.
- Add: mixed groups are split into runs of one genus; every group of two
  or more distinct names is checked against its hints, and an accepted
  genus expands only the names it matched.
- Add: names a mixed group leaves without a genus are reconciled one
  distinct name at a time with their own hints (groups marked `alone`).
- Add: OCR rereading of abbreviations with two capitals (`CS.`, `LL.`,
  `PL.`); reread names are grouped only with each other.
- Add: genus acceptance threshold, 8 by default (96.8% precision, 84.5%
  recall on the BHL holdout gold set).
- Remove: CLI (`resolve`, `tune`), configuration file and logging;
  gnabbr is a library only.

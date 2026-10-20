// Package doc defines the input of gnabbr: scientific names detected in
// a document (a book, an article, a web page) and, optionally, its
// sections. Positions are rune offsets from the start of the document;
// there are no pages.
package doc

import "github.com/gnames/gnstats/ent/stats"

// Document is a text with names detected in it, for example by gnfinder.
type Document struct {
	// ID of the document, for example a BHL item id.
	ID string `json:"id,omitempty"`

	// Names are all detected names, abbreviated or not, in any order.
	// Full names give genus hints to abbreviated ones, so they should be
	// kept. Repeated names are kept too: every occurrence is a Name.
	Names []Name `json:"names"`

	// Sections are optional autonomous parts of the document (articles,
	// chapters). An abbreviation refers only to genera of its section.
	Sections []Section `json:"sections,omitempty"`

	// TaxonStats are taxonomic statistics of the whole document, as
	// calculated by gnstats. They are required, but may be empty when
	// the caller has no verified names to calculate them from.
	TaxonStats stats.Stats `json:"taxonStats"`
}

// Name is one detected name, a subset of gnfinder's output.Name.
type Name struct {
	// ID of the name; the index of the name in Document.Names if empty.
	ID string `json:"id,omitempty"`

	// Verbatim is the name as it is in the text, carried to the output.
	Verbatim string `json:"verbatim,omitempty"`

	// Name is the normalized name: `C. alba`, `Carex alba`.
	Name string `json:"name"`

	// Start is the position of the name in runes from the start of the
	// document.
	Start int `json:"start"`
}

// Section is an autonomous part of a document given as a rune range.
// Sections may overlap.
type Section struct {
	// ID of the section, for example a BHL part id.
	ID string `json:"id,omitempty"`

	// Start and End are the rune range [Start, End) of the section.
	Start int `json:"start"`
	End   int `json:"end"`

	// TaxonStats of the section; the document's when nil.
	TaxonStats *stats.Stats `json:"taxonStats,omitempty"`
}

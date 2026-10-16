// Package doc defines the input of gnabbr: scientific names found in a
// document (a BHL item, a PDF, a web page), in reading order.
package doc

// Document is a text with names detected in it, for example by gnfinder.
type Document struct {
	// ID of the document, for example a BHL item id.
	ID string

	// Occurrences of all detected names, abbreviated or not. Full names
	// give genus hints for abbreviated ones, so they should be kept.
	Occurrences []Occurrence

	// Contexts maps a part id (see Occurrence.Part) to the taxonomic
	// context of that part. The empty key is the context of the whole
	// document.
	Contexts map[string]Context
}

// Occurrence is one detected name.
type Occurrence struct {
	// ID of the occurrence; it must be unique in the document.
	ID string

	// Name as detected: `C. alba`, `Carex alba L.`.
	Name string

	// Page is the position of the page in the document (sequence
	// order), starting from any number.
	Page int

	// Offset is the position of the name on its page.
	Offset int

	// Part is the id of a part of the document (an article, a chapter),
	// if known. Abbreviations are resolved within their part.
	Part string
}

// Context is the taxonomic context of a document or one of its parts.
type Context struct {
	// MainTaxon is the lowest taxon that contains most of the names.
	MainTaxon string

	// DominantClass is the class with most of the names.
	DominantClass string
}

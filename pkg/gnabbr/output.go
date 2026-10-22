package gnabbr

import (
	"github.com/gnames/gnlib/ent/abbr"
)

// Output is the result of reconciling a document.
type Output struct {
	Meta `json:"metadata"`

	// Reconciled are groups whose best genus reached the threshold, with
	// their expanded names.
	Reconciled []Group `json:"reconciled"`

	// Failed are groups without an accepted genus, with their distinct
	// verbatims for inspection. Their names stay as they were.
	Failed []Group `json:"failed"`
}

// Meta describes a run.
type Meta struct {
	// Version of gnabbr.
	Version string `json:"version"`

	// DocumentID is the id of the document.
	DocumentID string `json:"documentId,omitempty"`

	// Settings used for the run.
	Settings Settings `json:"settings"`

	// NamesNum is the number of input names.
	NamesNum int `json:"namesNum"`

	// AbbrNamesNum is the number of input names with an abbreviated
	// genus.
	AbbrNamesNum int `json:"abbrNamesNum"`

	// ReconciledNum is the number of expanded names.
	ReconciledNum int `json:"reconciledNum"`

	// TotalSec is the duration of the run.
	TotalSec float32 `json:"totalSec"`
}

// Settings are the settings of a run that change its results.
type Settings struct {
	GapChars    int     `json:"gapChars"`
	SpanChars   int     `json:"spanChars"`
	HintWindow  int     `json:"hintWindow"`
	HintsLimit  int     `json:"hintsLimit"`
	GeneraLimit int     `json:"generaLimit"`
	Threshold   float64 `json:"threshold"`
}

// Group is a group of abbreviated names with the same abbreviation that
// share a genus.
type Group struct {
	// Index of the group among all groups of the document.
	Index int `json:"index"`

	// Section is the id of the group's section, empty for the document.
	Section string `json:"section,omitempty"`

	// Abbr is the abbreviation without its period: `C`, `Cae`.
	Abbr string `json:"abbr"`

	// Start and End are the offsets of the first and the last name.
	Start int `json:"start"`
	End   int `json:"end"`

	// TaxonStats sent to abbr_names.
	TaxonStats TaxonStats `json:"taxonStats"`

	// Split is true for a run of names that share a genus, cut out of a
	// group whose genus, if any, its hints did not support as a whole (a
	// mixed group).
	Split bool `json:"split,omitempty"`

	// Hints are the hints of the group, nearest first: full names, or
	// genera for uninomials.
	Hints []string `json:"hints,omitempty"`

	// Genus is the best candidate genus, nil if there is none.
	Genus *abbr.Genus `json:"genus,omitempty"`

	// Candidates are the runners-up.
	Candidates []abbr.Genus `json:"candidates,omitempty"`

	// Names are the expanded names (Reconciled groups only).
	Names []Name `json:"names,omitempty"`

	// Verbatims are distinct verbatims of the group's names that stay as
	// they were: all names of a Failed group, names of a Reconciled group
	// its genus did not match.
	Verbatims []string `json:"verbatims,omitempty"`

	// Details are raw abbr_names input and output (WithDetails).
	Details *Details `json:"details,omitempty"`
}

// TaxonStats are the taxonomic statistics abbr_names uses.
type TaxonStats struct {
	MainTaxon     string `json:"mainTaxon,omitempty"`
	DominantClass string `json:"dominantClass,omitempty"`
}

// Name is an expanded name; it mirrors gnfinder's output.Name.
type Name struct {
	// ID of the input name.
	ID string `json:"id"`

	// Verbatim of the input name.
	Verbatim string `json:"verbatim,omitempty"`

	// Name is the accepted genus with the rest of the name as written:
	// `Carex alba` for `C. alba`.
	Name string `json:"name"`

	// AbbrName is the input name: `C. alba`, or its reading if its
	// abbreviation has two capitals in the text (`C. alba` for `JC. alba`).
	AbbrName string `json:"abbrName"`

	// Start of the name in runes from the start of the document.
	Start int `json:"start"`
}

// Details are raw abbr_names input and output of a group.
type Details struct {
	Input  abbr.Input  `json:"input"`
	Output abbr.Output `json:"output"`
}

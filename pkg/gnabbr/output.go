package gnabbr

import (
	"github.com/gnames/gnabbr/pkg/ent/group"
	"github.com/gnames/gnlib/ent/abbr"
)

// Output is the result of resolving a document.
type Output struct {
	// DocumentID is the id of the resolved document.
	DocumentID string `json:"documentId,omitempty"`

	// Results has one entry per abbreviated occurrence, in the order of
	// the document's occurrences.
	Results []Result `json:"results"`

	// Groups are the groups sent to abbr_names, with its answers.
	Groups []GroupOutput `json:"groups"`
}

// Result is the best genus for one abbreviated occurrence.
type Result struct {
	// OccurrenceID is the id of the occurrence in the document.
	OccurrenceID string `json:"occurrenceId"`

	// Name is the abbreviated name: `C. elegans`.
	Name string `json:"name"`

	// Genus is the best candidate genus, empty when none was found.
	Genus string `json:"genus,omitempty"`

	// Score is the total score of the genus.
	Score float64 `json:"score"`

	// Accepted is true when the score reaches the threshold.
	Accepted bool `json:"accepted"`

	// Group is the index of the occurrence's group in Output.Groups.
	Group int `json:"group"`
}

// GroupOutput is a group of abbreviated names with the abbr_names
// answer for it.
type GroupOutput struct {
	group.Group
	Input  abbr.Input  `json:"input"`
	Output abbr.Output `json:"output"`
}

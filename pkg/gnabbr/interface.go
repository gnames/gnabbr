package gnabbr

import (
	"context"

	"github.com/gnames/gnabbr/pkg/ent/doc"
	"github.com/gnames/gnlib/ent/abbr"
)

// GNabbr resolves abbreviated genera of scientific names (`C. elegans`)
// found in a document.
type GNabbr interface {
	// Resolve groups abbreviated names of a document, sends each group
	// to gnames abbr_names and returns the best genus for every
	// abbreviated occurrence.
	Resolve(ctx context.Context, d doc.Document) (Output, error)
}

// AbbrNamer finds candidate genera for a group of names with the same
// abbreviated genus. gnames abbr_names is its main implementation.
type AbbrNamer interface {
	AbbrNames(ctx context.Context, inp abbr.Input) (abbr.Output, error)
}

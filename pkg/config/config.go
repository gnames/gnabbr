// Package config provides settings of gnabbr.
//
// This package has no I/O dependencies (no file operations, no network calls).
//
// # Design Principles
//
// - Default config (from New()) is always valid - no validation needed
// - All mutations go through Option functions - the only way to modify Config
// - Invalid options are rejected with slog.Warn() - config remains in valid state
package config

// Config represents settings of gnabbr.
type Config struct {
	// GnamesURL is the gnames REST API that provides abbr_names.
	GnamesURL string

	// GapChars: a group of abbreviated names ends where the next name
	// with the same abbreviation is more than GapChars runes away.
	GapChars int

	// SpanChars is the largest number of runes a group spans.
	SpanChars int

	// HintWindow is how many runes back full names give genus hints.
	HintWindow int

	// HintsLimit is the largest number of genus hints per name.
	HintsLimit int

	// GeneraLimit is the largest number of candidate genera abbr_names
	// returns for a group.
	GeneraLimit int

	// Threshold is the lowest score of an accepted genus.
	Threshold float64

	// WithFeatures asks abbr_names for raw score features.
	WithFeatures bool

	// WithDetails adds raw abbr_names input and output to each group.
	WithDetails bool
}

// New creates a Config with sensible default values.
// The returned config is always valid and ready to use.
// Default values can be overridden using Option functions via Update().
func New() *Config {
	res := &Config{
		GnamesURL: "https://verifier.globalnames.org/api/v1/",
		// A BHL page has about 2,500 runes: 10, 50 and 50 pages.
		GapChars:    25_000,
		SpanChars:   125_000,
		HintWindow:  125_000,
		HintsLimit:  5,
		GeneraLimit: 5,
		// Precision over recall: on the BHL holdout gold set 8 gives
		// 96.8% precision and 84.5% recall, 7 gives 96.0% and 88.0%.
		Threshold: 8,
	}
	return res
}

// Update applies configuration options.
// This is the only way to modify Config after creation.
// Invalid options are ignored (warnings via slog.Warn()).
func (c *Config) Update(opts []Option) {
	for _, opt := range opts {
		opt(c)
	}
}

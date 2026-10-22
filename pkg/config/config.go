// Package config provides configuration management for the application.
//
// This package has no I/O dependencies (no file operations, no network calls).
// Validation functions log warnings via slog.
//
// # Configuration Sources
//
// Precedence (highest to lowest): CLI flags > env vars > config.yaml > defaults
//
// # Design Principles
//
// - Default config (from New()) is always valid - no validation needed
// - All mutations go through Option functions - the only way to modify Config
// - Invalid options are rejected with slog.Warn() - config remains in valid state
// - ToOptions() converts persistent fields (those in config.yaml)
// - Environment variables match ToOptions() fields exactly
//
// # Persistent vs Runtime Fields
//
// Persistent fields (in ToOptions, config.yaml, and env vars):
//   - Log: level, format, destination
//   - General: jobs_number
//   - Resolver: gnames_url, gap_chars, span_chars, hint_window,
//     hints_limit, genera_limit, threshold
//
// Runtime-only fields (CLI flags only):
//   - HomeDir (set once at startup)
//   - WithFeatures, WithDetails
//
// # Environment Variables
//
// Use GNABBR_ prefix with underscores for nesting:
//
//	GNABBR_LOG_LEVEL=info
//	GNABBR_JOBS_NUMBER=8
package config

import (
	"runtime"
)

// Config represents the complete application configuration.
type Config struct {
	// Log contains logging configuration.
	Log LogConfig `mapstructure:"log" yaml:"log"`

	// JobsNumber is the number of concurrent workers for parallel operations.
	// Default value is set according to the number of available threads.
	JobsNumber int `mapstructure:"jobs_number" yaml:"jobs_number"`

	// GnamesURL is the gnames REST API that provides abbr_names.
	GnamesURL string `mapstructure:"gnames_url" yaml:"gnames_url"`

	// GapChars: a group of abbreviated names ends where the next name
	// with the same abbreviation is more than GapChars runes away.
	GapChars int `mapstructure:"gap_chars" yaml:"gap_chars"`

	// SpanChars is the largest number of runes a group spans.
	SpanChars int `mapstructure:"span_chars" yaml:"span_chars"`

	// HintWindow is how many runes back full names give genus hints.
	HintWindow int `mapstructure:"hint_window" yaml:"hint_window"`

	// HintsLimit is the largest number of genus hints per name.
	HintsLimit int `mapstructure:"hints_limit" yaml:"hints_limit"`

	// GeneraLimit is the largest number of candidate genera abbr_names
	// returns for a group.
	GeneraLimit int `mapstructure:"genera_limit" yaml:"genera_limit"`

	// Threshold is the lowest score of an accepted genus.
	Threshold float64 `mapstructure:"threshold" yaml:"threshold"`

	// WithFeatures asks abbr_names for raw score features (tuning).
	WithFeatures bool

	// WithDetails adds raw abbr_names input and output to each group.
	WithDetails bool

	// HomeDir determines where config, cache and logs directories reside.
	// It must be set by CLI during init, there is no default value for it.
	HomeDir string
}

// LogConfig provides typical settings for application logs.
type LogConfig struct {
	// Format can be 'json', 'text' or 'tint' (user-facing and colored).
	Format string `mapstructure:"format"      yaml:"format"`
	// Level of logging -- 'error', 'warn', 'info', 'debug'
	Level string `mapstructure:"level"       yaml:"level"`
	// Destination can be a log file (to default place), STDERR or STDOUT
	Destination string `mapstructure:"destination" yaml:"destination"`
}

// New creates a Config with sensible default values.
// The returned config is always valid and ready to use.
// Default values can be overridden using Option functions via Update().
func New() *Config {
	res := &Config{
		Log: LogConfig{
			Format:      "json",
			Level:       "info",
			Destination: "file",
		},
		JobsNumber: runtime.NumCPU(), // Default to number of CPU threads
		GnamesURL:  "https://verifier.globalnames.org/api/v1/",
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

// ToOptions converts the persistent configuration fields to Option functions.
// Used for config.yaml round-trip (load from file, save to file).
// Only includes fields that should be persisted in config.yaml.
// Runtime-only fields (HomeDir) are excluded.
func (c *Config) ToOptions() []Option {
	var res []Option
	var s string
	var i int
	s = c.Log.Format
	if s != "" {
		res = append(res, OptLogFormat(s))
	}
	s = c.Log.Level
	if s != "" {
		res = append(res, OptLogLevel(s))
	}
	s = c.Log.Destination
	if s != "" {
		res = append(res, OptLogDestination(s))
	}

	i = c.JobsNumber
	if i > 0 {
		res = append(res, OptJobsNumber(i))
	}

	if c.GnamesURL != "" {
		res = append(res, OptGnamesURL(c.GnamesURL))
	}
	for _, v := range []struct {
		i   int
		opt func(int) Option
	}{
		{c.GapChars, OptGapChars},
		{c.SpanChars, OptSpanChars},
		{c.HintWindow, OptHintWindow},
		{c.HintsLimit, OptHintsLimit},
		{c.GeneraLimit, OptGeneraLimit},
	} {
		if v.i > 0 {
			res = append(res, v.opt(v.i))
		}
	}
	res = append(res, OptThreshold(c.Threshold))
	return res
}

package config

import (
	"slices"
	"strings"

	"github.com/gnames/gn"
)

// Option is a function that modifies Config.
// This is the functional options pattern.
type Option func(*Config)

// OptLogFormat sets the log format.
func OptLogFormat(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Format", s, []string{"json", "text", "tint"}) {
			c.Log.Format = s
		}
	}
}

// OptLogLevel sets the log level.
func OptLogLevel(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Level", s, []string{"debug", "info", "warn", "error"}) {
			c.Log.Level = s
		}
	}
}

// OptLogDestination sets the log destination.
func OptLogDestination(s string) Option {
	return func(c *Config) {
		if isValidEnum("Log Destination", s, []string{"file", "stdout", "stderr"}) {
			c.Log.Destination = s
		}
	}
}

// OptJobsNumber sets the number of concurrent workers.
func OptJobsNumber(i int) Option {
	return func(c *Config) {
		if isValidInt("Jobs Number", i) {
			c.JobsNumber = i
		}
	}
}

// OptGnamesURL sets the gnames REST API URL.
func OptGnamesURL(s string) Option {
	s = strings.TrimSpace(s)
	return func(c *Config) {
		if isValidString("Gnames URL", s) {
			c.GnamesURL = s
		}
	}
}

// OptGapPages sets the page gap that splits groups.
func OptGapPages(i int) Option {
	return func(c *Config) {
		if isValidInt("Gap Pages", i) {
			c.GapPages = i
		}
	}
}

// OptSpanPages sets the largest number of pages in a group.
func OptSpanPages(i int) Option {
	return func(c *Config) {
		if isValidInt("Span Pages", i) {
			c.SpanPages = i
		}
	}
}

// OptHintWindow sets how many pages back genus hints come from.
func OptHintWindow(i int) Option {
	return func(c *Config) {
		if isValidInt("Hint Window", i) {
			c.HintWindow = i
		}
	}
}

// OptHintsLimit sets the largest number of genus hints per name.
func OptHintsLimit(i int) Option {
	return func(c *Config) {
		if isValidInt("Hints Limit", i) {
			c.HintsLimit = i
		}
	}
}

// OptGeneraLimit sets the largest number of candidate genera per group.
func OptGeneraLimit(i int) Option {
	return func(c *Config) {
		if isValidInt("Genera Limit", i) {
			c.GeneraLimit = i
		}
	}
}

// OptThreshold sets the lowest score of an accepted genus.
func OptThreshold(f float64) Option {
	return func(c *Config) {
		c.Threshold = f
	}
}

// OptWithFeatures asks abbr_names for raw score features.
// Runtime-only field - not in ToOptions().
func OptWithFeatures(b bool) Option {
	return func(c *Config) {
		c.WithFeatures = b
	}
}

// OptHomeDir sets the home directory for config, cache, and log locations.
// Set once at startup from os.UserHomeDir().
// Runtime-only field - not in ToOptions().
func OptHomeDir(s string) Option {
	s = strings.TrimSpace(s)
	return func(c *Config) {
		if isValidString("Home Directory", s) {
			c.HomeDir = s
		}
	}
}

// Validation helpers

func isValidString(name, value string) bool {
	if value == "" {
		gn.Warn("%s cannot be empty", name)
		return false
	}
	return true
}

func isValidInt(name string, value int) bool {
	if value < 0 {
		gn.Warn("%s must be non-negative, got %d", name, value)
		return false
	}
	return true
}

func isValidEnum(name, value string, allowed []string) bool {
	if slices.Contains(allowed, value) {
		return true
	}
	gn.Warn("%s must be one of %v, got '%s'", name, allowed, value)
	return false
}

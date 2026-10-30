package config

import (
	"log/slog"
	"strings"
)

// Option is a function that modifies Config.
// This is the functional options pattern.
type Option func(*Config)

// OptGnamesURL sets the gnames REST API URL.
func OptGnamesURL(s string) Option {
	s = strings.TrimSpace(s)
	return func(c *Config) {
		if isValidString("Gnames URL", s) {
			c.GnamesURL = s
		}
	}
}

// OptGapChars sets the gap in runes that splits groups.
func OptGapChars(i int) Option {
	return func(c *Config) {
		if isValidInt("Gap Chars", i) {
			c.GapChars = i
		}
	}
}

// OptSpanChars sets the largest number of runes in a group.
func OptSpanChars(i int) Option {
	return func(c *Config) {
		if isValidInt("Span Chars", i) {
			c.SpanChars = i
		}
	}
}

// OptHintWindow sets how many runes back genus hints come from.
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
func OptWithFeatures(b bool) Option {
	return func(c *Config) {
		c.WithFeatures = b
	}
}

// OptWithDetails adds raw abbr_names input and output to the groups.
func OptWithDetails(b bool) Option {
	return func(c *Config) {
		c.WithDetails = b
	}
}

// Validation helpers

func isValidString(name, value string) bool {
	if value == "" {
		slog.Warn("Option rejected: value cannot be empty", "option", name)
		return false
	}
	return true
}

func isValidInt(name string, value int) bool {
	if value < 0 {
		slog.Warn("Option rejected: value must be non-negative",
			"option", name, "value", value)
		return false
	}
	return true
}
